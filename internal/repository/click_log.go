package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/panhao/url-shortener/internal/model"
)

type ClickLogRepo struct {
	pool *pgxpool.Pool
}

func NewClickLogRepo(pool *pgxpool.Pool) *ClickLogRepo {
	return &ClickLogRepo{pool: pool}
}

func (r *ClickLogRepo) BatchInsert(ctx context.Context, logs []model.ClickLog) error {
	if len(logs) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, log := range logs {
		batch.Queue(
			`INSERT INTO click_logs (link_id, ip, user_agent, referer, country, city, device_type, browser, os, clicked_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			log.LinkID, log.IP, log.UserAgent, log.Referer,
			log.Country, log.City, log.DeviceType, log.Browser, log.OS, log.ClickedAt,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(logs); i++ {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (r *ClickLogRepo) GetLinkStats(ctx context.Context, linkID int64, since time.Time) (map[string]any, error) {
	stats := make(map[string]any)

	var total int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM click_logs WHERE link_id = $1`, linkID,
	).Scan(&total)
	if err != nil {
		return nil, err
	}
	stats["total_clicks"] = total

	var uniqueIPs int64
	err = r.pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT ip) FROM click_logs WHERE link_id = $1`, linkID,
	).Scan(&uniqueIPs)
	if err != nil {
		return nil, err
	}
	stats["unique_ips"] = uniqueIPs

	rows, err := r.pool.Query(ctx, `
		SELECT DATE_TRUNC('hour', clicked_at) as t, COUNT(*) as c
		FROM click_logs
		WHERE link_id = $1 AND clicked_at >= $2
		GROUP BY t ORDER BY t`, linkID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trend []map[string]any
	for rows.Next() {
		var t time.Time
		var c int
		if err := rows.Scan(&t, &c); err != nil {
			return nil, err
		}
		trend = append(trend, map[string]any{"time": t, "clicks": c})
	}
	stats["trend"] = trend

	geoRows, err := r.pool.Query(ctx, `
		SELECT COALESCE(country, 'Unknown'), COUNT(*)
		FROM click_logs WHERE link_id = $1 AND country IS NOT NULL AND country != ''
		GROUP BY country ORDER BY COUNT(*) DESC LIMIT 20`, linkID)
	if err != nil {
		return nil, err
	}
	defer geoRows.Close()

	var geo []map[string]any
	for geoRows.Next() {
		var country string
		var count int
		if err := geoRows.Scan(&country, &count); err != nil {
			return nil, err
		}
		geo = append(geo, map[string]any{"country": country, "clicks": count})
	}
	stats["geo"] = geo

	refRows, err := r.pool.Query(ctx, `
		SELECT COALESCE(referer, 'Direct'), COUNT(*)
		FROM click_logs WHERE link_id = $1
		GROUP BY referer ORDER BY COUNT(*) DESC LIMIT 10`, linkID)
	if err != nil {
		return nil, err
	}
	defer refRows.Close()

	var referrers []map[string]any
	for refRows.Next() {
		var ref string
		var count int
		if err := refRows.Scan(&ref, &count); err != nil {
			return nil, err
		}
		referrers = append(referrers, map[string]any{"referer": ref, "clicks": count})
	}
	stats["referrers"] = referrers

	return stats, nil
}

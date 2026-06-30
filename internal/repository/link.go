package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/panhao/url-shortener/internal/model"
)

type LinkRepo struct {
	pool *pgxpool.Pool
}

func NewLinkRepo(pool *pgxpool.Pool) *LinkRepo {
	return &LinkRepo{pool: pool}
}

func (r *LinkRepo) Create(ctx context.Context, link *model.Link) error {
	query := `
		INSERT INTO links (short_code, original_url, title, description, user_id, domain, expire_at, password_hash, rules, redirect_type)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		link.ShortCode, link.OriginalURL, link.Title, link.Description,
		link.UserID, link.Domain, link.ExpireAt, link.PasswordHash,
		link.Rules, link.RedirectType,
	).Scan(&link.ID, &link.CreatedAt, &link.UpdatedAt)
}

func (r *LinkRepo) GetByShortCode(ctx context.Context, code string) (*model.Link, error) {
	query := `
		SELECT id, short_code, original_url, title, description, user_id, domain,
		       expire_at, password_hash, rules, redirect_type, is_active, click_count,
		       created_at, updated_at
		FROM links WHERE short_code = $1`

	link := &model.Link{}
	var userID *int64
	err := r.pool.QueryRow(ctx, query, code).Scan(
		&link.ID, &link.ShortCode, &link.OriginalURL, &link.Title, &link.Description,
		&userID, &link.Domain, &link.ExpireAt, &link.PasswordHash, &link.Rules,
		&link.RedirectType, &link.IsActive, &link.ClickCount,
		&link.CreatedAt, &link.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	link.UserID = userID
	return link, err
}

func (r *LinkRepo) GetByID(ctx context.Context, id int64) (*model.Link, error) {
	query := `
		SELECT id, short_code, original_url, title, description, user_id, domain,
		       expire_at, password_hash, rules, redirect_type, is_active, click_count,
		       created_at, updated_at
		FROM links WHERE id = $1`

	link := &model.Link{}
	var userID *int64
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&link.ID, &link.ShortCode, &link.OriginalURL, &link.Title, &link.Description,
		&userID, &link.Domain, &link.ExpireAt, &link.PasswordHash, &link.Rules,
		&link.RedirectType, &link.IsActive, &link.ClickCount,
		&link.CreatedAt, &link.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	link.UserID = userID
	return link, err
}

func (r *LinkRepo) IsShortCodeExists(ctx context.Context, code string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM links WHERE short_code = $1)`, code).Scan(&exists)
	return exists, err
}

func (r *LinkRepo) ListByUser(ctx context.Context, userID int64, page, pageSize int) ([]model.Link, int64, error) {
	var total int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM links WHERE user_id = $1`, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := `
		SELECT id, short_code, original_url, title, description, user_id, domain,
		       expire_at, rules, redirect_type, is_active, click_count,
		       created_at, updated_at
		FROM links WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var links []model.Link
	for rows.Next() {
		var l model.Link
		var uid *int64
		if err := rows.Scan(
			&l.ID, &l.ShortCode, &l.OriginalURL, &l.Title, &l.Description,
			&uid, &l.Domain, &l.ExpireAt, &l.Rules,
			&l.RedirectType, &l.IsActive, &l.ClickCount,
			&l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		l.UserID = uid
		links = append(links, l)
	}
	return links, total, nil
}

func (r *LinkRepo) Update(ctx context.Context, code string, updates map[string]any) error {
	query := `UPDATE links SET updated_at = NOW()`
	args := []any{code}
	idx := 2

	if v, ok := updates["title"]; ok {
		query += fmt.Sprintf(`, title = $%d`, idx)
		args = append(args, v)
		idx++
	}
	if v, ok := updates["description"]; ok {
		query += fmt.Sprintf(`, description = $%d`, idx)
		args = append(args, v)
		idx++
	}
	if v, ok := updates["expire_at"]; ok {
		query += fmt.Sprintf(`, expire_at = $%d`, idx)
		args = append(args, v)
		idx++
	}
	if v, ok := updates["is_active"]; ok {
		query += fmt.Sprintf(`, is_active = $%d`, idx)
		args = append(args, v)
		idx++
	}
	if v, ok := updates["redirect_type"]; ok {
		query += fmt.Sprintf(`, redirect_type = $%d`, idx)
		args = append(args, v)
		idx++
	}

	query += fmt.Sprintf(` WHERE short_code = $1`)
	_, err := r.pool.Exec(ctx, query, args...)
	return err
}

func (r *LinkRepo) Delete(ctx context.Context, code string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM links WHERE short_code = $1`, code)
	return err
}

func (r *LinkRepo) IncrementClickCount(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `UPDATE links SET click_count = click_count + 1 WHERE id = $1`, id)
	return err
}

func (r *LinkRepo) GetUserLink(ctx context.Context, code string, userID int64) (*model.Link, error) {
	query := `
		SELECT id, short_code, original_url, title, description, user_id, domain,
		       expire_at, rules, redirect_type, is_active, click_count,
		       created_at, updated_at
		FROM links WHERE short_code = $1 AND user_id = $2`

	link := &model.Link{}
	var uid *int64
	err := r.pool.QueryRow(ctx, query, code, userID).Scan(
		&link.ID, &link.ShortCode, &link.OriginalURL, &link.Title, &link.Description,
		&uid, &link.Domain, &link.ExpireAt, &link.Rules,
		&link.RedirectType, &link.IsActive, &link.ClickCount,
		&link.CreatedAt, &link.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	link.UserID = uid
	return link, err
}

func (r *LinkRepo) GetAllShortCodes(ctx context.Context) ([]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT short_code FROM links WHERE is_active = true`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var codes []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		codes = append(codes, code)
	}
	return codes, nil
}

func (r *LinkRepo) GetDashboardOverview(ctx context.Context, userID int64) (*model.DashboardOverview, error) {
	overview := &model.DashboardOverview{}

	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM links WHERE user_id = $1`, userID,
	).Scan(&overview.TotalLinks)
	if err != nil {
		return nil, err
	}

	err = r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(click_count), 0) FROM links WHERE user_id = $1`, userID,
	).Scan(&overview.TotalClicks)
	if err != nil {
		return nil, err
	}

	today := time.Now().Truncate(24 * time.Hour)
	err = r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(clicks), 0) FROM click_stats_hourly cs
		 JOIN links l ON cs.link_id = l.id
		 WHERE l.user_id = $1 AND cs.hour >= $2`, userID, today,
	).Scan(&overview.TodayClicks)
	if err != nil {
		return nil, err
	}

	err = r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM links WHERE user_id = $1 AND is_active = true
		 AND (expire_at IS NULL OR expire_at > NOW())`, userID,
	).Scan(&overview.ActiveLinks)
	if err != nil {
		return nil, err
	}

	err = r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM links WHERE user_id = $1 AND expire_at <= NOW()`, userID,
	).Scan(&overview.ExpiredLinks)
	if err != nil {
		return nil, err
	}

	var days int
	err = r.pool.QueryRow(ctx,
		`SELECT GREATEST(1, EXTRACT(DAY FROM NOW() - MIN(created_at))::int) FROM links WHERE user_id = $1`, userID,
	).Scan(&days)
	if err != nil || days == 0 {
		days = 1
	}
	overview.AvgClicksPerDay = float64(overview.TotalClicks) / float64(days)

	return overview, nil
}

func (r *LinkRepo) GetTrendData(ctx context.Context, userID int64, hours int) ([]map[string]any, error) {
	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	rows, err := r.pool.Query(ctx, `
		SELECT DATE_TRUNC('hour', cs.hour) as t, SUM(cs.clicks) as c
		FROM click_stats_hourly cs
		JOIN links l ON cs.link_id = l.id
		WHERE l.user_id = $1 AND cs.hour >= $2
		GROUP BY t ORDER BY t`, userID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]any
	for rows.Next() {
		var t time.Time
		var c int
		if err := rows.Scan(&t, &c); err != nil {
			return nil, err
		}
		result = append(result, map[string]any{"time": t, "clicks": c})
	}
	return result, nil
}

func (r *LinkRepo) GetGeoData(ctx context.Context, userID int64) ([]map[string]any, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT COALESCE(cl.country, 'Unknown'), COUNT(*)
		FROM click_logs cl
		JOIN links l ON cl.link_id = l.id
		WHERE l.user_id = $1 AND cl.country IS NOT NULL AND cl.country != ''
		GROUP BY cl.country ORDER BY COUNT(*) DESC LIMIT 20`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]any
	for rows.Next() {
		var country string
		var count int
		if err := rows.Scan(&country, &count); err != nil {
			return nil, err
		}
		result = append(result, map[string]any{"country": country, "clicks": count})
	}
	return result, nil
}

func (r *LinkRepo) GetDeviceData(ctx context.Context, userID int64) ([]map[string]any, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT COALESCE(cl.device_type, 'Unknown'), COUNT(*)
		FROM click_logs cl
		JOIN links l ON cl.link_id = l.id
		WHERE l.user_id = $1
		GROUP BY cl.device_type ORDER BY COUNT(*) DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]any
	for rows.Next() {
		var device string
		var count int
		if err := rows.Scan(&device, &count); err != nil {
			return nil, err
		}
		result = append(result, map[string]any{"device": device, "clicks": count})
	}
	return result, nil
}

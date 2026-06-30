package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/panhao/url-shortener/internal/model"
)

type APIKeyRepo struct {
	pool *pgxpool.Pool
}

func NewAPIKeyRepo(pool *pgxpool.Pool) *APIKeyRepo {
	return &APIKeyRepo{pool: pool}
}

func (r *APIKeyRepo) Create(ctx context.Context, key *model.APIKey) error {
	query := `INSERT INTO api_keys (user_id, name, key_prefix, key_hash, rate_limit)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`
	return r.pool.QueryRow(ctx, query, key.UserID, key.Name, key.KeyPrefix, key.KeyHash, key.RateLimit).
		Scan(&key.ID, &key.CreatedAt)
}

func (r *APIKeyRepo) ListByUser(ctx context.Context, userID int64) ([]model.APIKey, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, name, key_prefix, rate_limit, is_active, last_used_at, created_at
		 FROM api_keys WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []model.APIKey
	for rows.Next() {
		var k model.APIKey
		if err := rows.Scan(&k.ID, &k.UserID, &k.Name, &k.KeyPrefix, &k.RateLimit,
			&k.IsActive, &k.LastUsedAt, &k.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, nil
}

func (r *APIKeyRepo) Delete(ctx context.Context, id, userID int64) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM api_keys WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

func (r *APIKeyRepo) GetByHash(ctx context.Context, hash string) (*model.APIKey, error) {
	k := &model.APIKey{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, name, key_prefix, key_hash, rate_limit, is_active, last_used_at, created_at
		 FROM api_keys WHERE key_hash = $1`, hash).Scan(
		&k.ID, &k.UserID, &k.Name, &k.KeyPrefix, &k.KeyHash, &k.RateLimit,
		&k.IsActive, &k.LastUsedAt, &k.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return k, err
}

func (r *APIKeyRepo) UpdateLastUsed(ctx context.Context, id int64) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, `UPDATE api_keys SET last_used_at = $1 WHERE id = $2`, now, id)
	return err
}

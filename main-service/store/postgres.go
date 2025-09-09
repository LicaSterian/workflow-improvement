package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresStore(
	ctx context.Context,
	dsn string,
	maxConns int32,
) (Store, func(context.Context) error, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, nil, err
	}
	if maxConns > 0 {
		cfg.MaxConns = maxConns
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, nil, err
	}
	close := func(c context.Context) error {
		pool.Close()
		return nil
	}
	return &store{
		DB: pool,
	}, close, nil
}

type store struct {
	DB *pgxpool.Pool
}

func (s *store) CreateItem(ctx context.Context, name string) (int64, error) {
	var id int64
	// Example SQL; adjust for your schema
	row := s.DB.QueryRow(ctx, `INSERT INTO items(name) VALUES($1) RETURNING id`, name)
	return id, row.Scan(&id)
}

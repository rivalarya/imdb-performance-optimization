package database

import (
	"context"
	"fmt"

	"imdb-performance-optimization/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabasePools struct {
	OptimizedPool   *pgxpool.Pool
	UnoptimizedPool *pgxpool.Pool
}

func ConnectToPostgres(cfg *config.Config) (*DatabasePools, error) {
	optimizedDSN := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	optimizedPool, err := pgxpool.New(context.Background(), optimizedDSN)
	if err != nil {
		return nil, err
	}

	unoptimizedConfig, err := pgxpool.ParseConfig(optimizedDSN)
	if err != nil {
		optimizedPool.Close()
		return nil, err
	}

	// Set for every new connection. So each unoptimized pool will have this config
	unoptimizedConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, `
			SET enable_indexscan = OFF;
			SET enable_bitmapscan = OFF;
			SET enable_indexonlyscan = OFF;
		`)
		return err
	}

	unoptimizedPool, err := pgxpool.NewWithConfig(context.Background(), unoptimizedConfig)
	if err != nil {
		optimizedPool.Close()
		return nil, err
	}

	return &DatabasePools{
		OptimizedPool:   optimizedPool,
		UnoptimizedPool: unoptimizedPool,
	}, nil
}

func (dp *DatabasePools) Close() {
	if dp.OptimizedPool != nil {
		dp.OptimizedPool.Close()
	}
	if dp.UnoptimizedPool != nil {
		dp.UnoptimizedPool.Close()
	}
}

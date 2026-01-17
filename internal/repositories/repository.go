package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunInExplain(ctx context.Context, pool *pgxpool.Pool, query string, args ...interface{}) (string, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	explainQuery := fmt.Sprintf("EXPLAIN (ANALYZE, BUFFERS, VERBOSE, FORMAT TEXT) %s", query)

	var plan string
	var line string
	rows, err := tx.Query(ctx, explainQuery, args...)
	if err != nil {
		return "", fmt.Errorf("explain failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&line); err != nil {
			return "", err
		}
		plan += line + "\n"
	}

	return plan, nil
}

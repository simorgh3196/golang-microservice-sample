package testutil

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// SetupPostgres はテスト用の PostgreSQL (pgvector) コンテナを起動し、
// 指定された初期化 SQL スクリプトを実行した上で接続プール (*pgxpool.Pool) を返します。
// コンテナと接続プールはテスト終了時 (t.Cleanup) に自動的に破棄されます。
func SetupPostgres(t *testing.T, initScripts ...string) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	opts := []testcontainers.ContainerCustomizer{
		postgres.WithDatabase("agentforge_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("password"),
		postgres.BasicWaitStrategies(),
	}
	if len(initScripts) > 0 {
		opts = append(opts, postgres.WithInitScripts(initScripts...))
	}

	pgContainer, err := postgres.Run(ctx, "pgvector/pgvector:pg17", opts...)
	require.NoError(t, err, "failed to start postgres container")

	t.Cleanup(func() {
		if termErr := testcontainers.TerminateContainer(pgContainer); termErr != nil {
			t.Logf("failed to terminate pstgres container: %v", termErr)
		}
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "failed to get connection string")

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err, "failed to create pgxpool")

	t.Cleanup(func() {
		pool.Close()
	})

	require.NoError(t, pool.Ping(ctx), "failed to ping postgres container")

	return pool
}

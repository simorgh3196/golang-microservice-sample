//go:build integration

package db_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/simorgh3196/golang-microservice-sample/apps/auth-service/internal/db"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestGetApiKeyByHash_Integration(t *testing.T) {
	ctx := context.Background()

	schemaPath, err := filepath.Abs(filepath.Join("..", "..", "db", "schema.sql"))
	require.NoError(t, err)

	// PostgreSQL 17 コンテナの起動とスキーマ初期化
	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:17-alpine",
		tcpostgres.WithInitScripts(schemaPath),
		tcpostgres.WithDatabase("agentforge_test"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(15*time.Second),
		),
	)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, testcontainers.TerminateContainer(pgContainer))
	}()

	// コンテナの接続文字列を取得して pgxpool で接続
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)
	defer pool.Close()

	queries := db.New(pool)

	// テストデータの投入
	tenantID, err := uuid.NewV7()
	require.NoError(t, err)

	_, err = pool.Exec(ctx,
		"INSERT INTO tenants (id, name, plan) VALUES ($1, $2, $3)",
		tenantID, "Test Company", "enterprise",
	)
	require.NoError(t, err)

	keyID := "key_test_123"
	keyHash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	_, err = pool.Exec(ctx,
		"INSERT INTO api_keys (id, tenant_id, key_hash, name, role, is_active) VALUES ($1, $2, $3, $4, $5, $6)",
		keyID, tenantID, keyHash, "Test API Key", "admin", true,
	)
	require.NoError(t, err)

	// テストケースの実行
	t.Run("[正常系] 存在するキーハッシュで正しく取得できる", func(t *testing.T) {
		apiKey, err := queries.GetApiKeyByHash(ctx, keyHash)
		require.NoError(t, err)

		want := db.ApiKey{
			ID:       keyID,
			TenantID: tenantID,
			KeyHash:  keyHash,
			Name:     "Test API Key",
			Role:     "admin",
			IsActive: true,
		}

		opts := cmpopts.IgnoreFields(db.ApiKey{}, "CreatedAt", "ExpiresAt")
		if diff := cmp.Diff(want, apiKey, opts); diff != "" {
			t.Errorf("予期しないレスポンスです: -want +got\n%s", diff)
		}
	})

	t.Run("[異常系] 存在しないキーハッシュの場合、ErrNoRows を返す", func(t *testing.T) {
		nonExistentHash := "non-existent-key-hash"
		_, err := queries.GetApiKeyByHash(ctx, nonExistentHash)
		require.Error(t, err)
		require.ErrorIs(t, err, pgx.ErrNoRows)
	})
}

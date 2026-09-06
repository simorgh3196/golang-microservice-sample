//go:build integration

package db_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/simorgh3196/golang-microservice-sample/apps/auth-service/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testPool    *pgxpool.Pool
	testQueries *db.Queries
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	schemaPath, err := filepath.Abs(filepath.Join("..", "..", "db", "schema.sql"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get schema path: %v\n", err)
		os.Exit(1)
	}

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
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start postgres container: %v\n", err)
		os.Exit(1)
	}

	// コンテナの接続文字列を取得して pgxpool で接続
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get connection string: %v\n", err)
		os.Exit(1)
	}

	testPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create db pool: %v\n", err)
		os.Exit(1)
	}
	testQueries = db.New(testPool)

	// パッケージ内の全テスト実行
	code := m.Run()

	// テスト完了後にリソースを一括解放
	testPool.Close()
	_ = testcontainers.TerminateContainer(pgContainer)

	os.Exit(code)
}

func createTestTenant(t *testing.T, name, plan string) uuid.UUID {
	t.Helper()
	tenantID, err := uuid.NewV7()
	require.NoError(t, err)

	_, err = testPool.Exec(context.Background(),
		"INSERT INTO tenants (id, name, plan) VALUES ($1, $2, $3)",
		tenantID, name, plan,
	)
	require.NoError(t, err)
	return tenantID
}

func createTestApiKey(t *testing.T, tenantID uuid.UUID, keyID, keyHash, name, role string, isActive bool) string {
	t.Helper()

	_, err := testPool.Exec(context.Background(),
		"INSERT INTO api_keys (id, tenant_id, key_hash, name, role, is_active) VALUES ($1, $2, $3, $4, $5, $6)",
		keyID, tenantID, keyHash, name, role, isActive,
	)
	require.NoError(t, err)

	return keyID
}

func TestGetApiKeyByHash_Integration(t *testing.T) {
	ctx := context.Background()

	tenantID := createTestTenant(t, "Test Company", "enterprise")
	keyID := "key_test_123"
	keyHash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	createTestApiKey(t, tenantID, keyID, keyHash, "Test API Key", "admin", true)

	// テストケースの実行
	t.Run("[正常系] 存在するキーハッシュで正しく取得できる", func(t *testing.T) {
		apiKey, err := testQueries.GetApiKeyByHash(ctx, keyHash)
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
		_, err := testQueries.GetApiKeyByHash(ctx, nonExistentHash)
		require.Error(t, err)
		require.ErrorIs(t, err, pgx.ErrNoRows)
	})
}

func TestGetTenantByID_Integration(t *testing.T) {
	ctx := context.Background()
	tenantID := createTestTenant(t, "Acme Corporation", "enterprise")

	t.Run("[正常系] 存在するテナントIDで正しく取得できる", func(t *testing.T) {
		tenant, err := testQueries.GetTenantByID(ctx, tenantID)
		require.NoError(t, err)

		want := db.Tenant{
			ID:   tenantID,
			Name: "Acme Corporation",
			Plan: "enterprise",
		}

		opts := cmpopts.IgnoreFields(db.Tenant{}, "CreatedAt", "UpdatedAt")
		if diff := cmp.Diff(want, tenant, opts); diff != "" {
			t.Errorf("予期しないレスポンスです: -want +got\n%s", diff)
		}
		assert.True(t, tenant.CreatedAt.Valid)
		assert.True(t, tenant.UpdatedAt.Valid)
	})

	t.Run("[異常系] 存在しないテナントIDの場合、ErrNoRows を返す", func(t *testing.T) {
		nonExistentTenantID, err := uuid.NewV7()
		require.NoError(t, err)

		_, err = testQueries.GetTenantByID(ctx, nonExistentTenantID)
		require.ErrorIs(t, err, pgx.ErrNoRows)
	})
}

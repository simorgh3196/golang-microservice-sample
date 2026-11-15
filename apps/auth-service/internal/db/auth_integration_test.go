//go:build integration

package db_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"

	"github.com/simorgh3196/golang-microservice-sample/apps/auth-service/internal/db"
	"github.com/simorgh3196/golang-microservice-sample/pkg/testutil"
)

func TestQueries_Integration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// internal/dbディレクトリから見た db/schema.sql のパスを指定してコンテナ起動
	schemaPath := filepath.Join("..", "..", "db", "schema.sql")
	pool := testutil.SetupPostgres(t, schemaPath)
	queries := db.New(pool)

	tenantID := uuid.MustParse("018f2b34-8c7a-7b3f-8000-000000000001")
	_, err := pool.Exec(ctx, `
		INSERT INTO tenants (id, name, plan)
		VALUES ($1, 'Acme Enterprise', 'enterprise')
	`, tenantID)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO api_keys (id, tenant_id, key_hash, name, role, is_active)
		VALUES
			('key_active_01',   $1, 'hash_active_123',   'Active Key',   'admin',  true),
			('key_inactive_02', $1, 'hash_inactive_456', 'Inactive Key', 'member', false)
	`, tenantID)
	require.NoError(t, err)

	t.Run("GetTenantByID: 登録したテナントが取得できる", func(t *testing.T) {
		t.Parallel()
		tenant, err := queries.GetTenantByID(ctx, tenantID)
		require.NoError(t, err)
		require.Equal(t, tenantID, tenant.ID)
		require.Equal(t, "Acme Enterprise", tenant.Name)
		require.Equal(t, "enterprise", tenant.Plan)
	})

	t.Run("GetApiKeyByHash: 有効なAPIキー(is_active=true)が取得できる", func(t *testing.T) {
		t.Parallel()
		key, err := queries.GetApiKeyByHash(ctx, "hash_active_123")
		require.NoError(t, err)
		require.Equal(t, "key_active_01", key.ID)
		require.Equal(t, tenantID, key.TenantID)
		require.Equal(t, "admin", key.Role)
		require.True(t, key.IsActive)
	})

	t.Run("GetApiKeyByHash: 無効なAPIキー(is_active=false)は取得できず ErrNoRows になる", func(t *testing.T) {
		t.Parallel()
		_, err := queries.GetApiKeyByHash(ctx, "hash_inactive_456")
		require.ErrorIs(t, err, pgx.ErrNoRows)
	})

	t.Run("GetApiKeyByHash: 存在しないハッシュは ErrNoRows になる", func(t *testing.T) {
		t.Parallel()
		_, err := queries.GetApiKeyByHash(ctx, "hash_not_found")
		require.ErrorIs(t, err, pgx.ErrNoRows)
	})
}

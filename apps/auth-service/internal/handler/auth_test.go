package handler_test

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/simorgh3196/golang-microservice-sample/apps/auth-service/internal/db"
	"github.com/simorgh3196/golang-microservice-sample/apps/auth-service/internal/handler"
	authv1 "github.com/simorgh3196/golang-microservice-sample/pkg/gen/agentforge/auth/v1"
)

type mockStore struct {
	getApiKeyByHashFn func(ctx context.Context, keyHash string) (db.ApiKey, error)
	getTenantByIDFn   func(ctx context.Context, id uuid.UUID) (db.Tenant, error)
}

func (m *mockStore) GetApiKeyByHash(ctx context.Context, keyHash string) (db.ApiKey, error) {
	if m.getApiKeyByHashFn != nil {
		return m.getApiKeyByHashFn(ctx, keyHash)
	}
	return db.ApiKey{}, errors.New("unexpected call to GetApiKeyHash")
}

func (m *mockStore) GetTenantByID(ctx context.Context, id uuid.UUID) (db.Tenant, error) {
	if m.getTenantByIDFn != nil {
		return m.getTenantByIDFn(ctx, id)
	}
	return db.Tenant{}, errors.New("unexpected call to GetTenantByID")
}

func TestAuthHandler_ValidateApiKey(t *testing.T) {
	t.Parallel()

	validTenantID := uuid.MustParse("018f2b34-8c7a-7b3f-8000-000000000001")
	ctx := context.Background()

	tests := []struct {
		name          string
		apiKey        string
		store         *mockStore
		wantErrorCode connect.Code
		wantResponse  *authv1.ValidateApiKeyResponse
	}{
		{
			name:   "[正常系] 有効なAPIキーの場合、認証に成功してテナントが返る",
			apiKey: "agf_live_test_key_123",
			store: &mockStore{
				getApiKeyByHashFn: func(ctx context.Context, keyHash string) (db.ApiKey, error) {
					return db.ApiKey{
						ID:       "key_test_01",
						TenantID: validTenantID,
						Role:     "admin",
						IsActive: true,
					}, nil
				},
			},
			wantResponse: &authv1.ValidateApiKeyResponse{
				TenantId: validTenantID.String(),
				KeyId:    "key_test_01",
				Role:     "admin",
			},
		},
		{
			name:   "[準正常系] 無効なAPIキーの場合、IsValid=false が返る",
			apiKey: "invalid-api-key",
			store: &mockStore{
				getApiKeyByHashFn: func(ctx context.Context, keyHash string) (db.ApiKey, error) {
					return db.ApiKey{}, pgx.ErrNoRows
				},
			},
			wantErrorCode: connect.CodeUnauthenticated,
		},
		{
			name:          "[異常系] APIキーが空の場合、InvalidArgument エラーが返る",
			apiKey:        "",
			store:         &mockStore{},
			wantErrorCode: connect.CodeInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := handler.NewAuthHandler(tt.store)
			req := connect.NewRequest(&authv1.ValidateApiKeyRequest{
				ApiKey: tt.apiKey,
			})
			res, err := h.ValidateApiKey(ctx, req)

			if tt.wantErrorCode != 0 {
				require.Error(t, err)
				require.Equal(t, tt.wantErrorCode, connect.CodeOf(err))
				return
			}
			require.NoError(t, err)

			if diff := cmp.Diff(tt.wantResponse, res.Msg, protocmp.Transform()); diff != "" {
				t.Errorf("予期しないレスポンスです: -want +got\n%s", diff)
			}
		})
	}
}

func TestAuthHandler_GetTenant(t *testing.T) {
	t.Parallel()

	validTenantID := uuid.MustParse("018f2b34-8c7a-7b3f-8000-000000000001")
	ctx := context.Background()

	tests := []struct {
		name          string
		tenantID      string
		store         *mockStore
		wantErrorCode connect.Code
		wantResponse  *authv1.GetTenantResponse
	}{
		{
			name:     "[正常系] 有効なテナントIDの場合、テナント詳細が返る",
			tenantID: "018f2b34-8c7a-7b3f-8000-000000000001",
			store: &mockStore{
				getTenantByIDFn: func(ctx context.Context, id uuid.UUID) (db.Tenant, error) {
					return db.Tenant{
						ID:   validTenantID,
						Name: "Acme Enterprise Corp",
						Plan: "enterprise",
					}, nil
				},
			},
			wantResponse: &authv1.GetTenantResponse{
				TenantId: "018f2b34-8c7a-7b3f-8000-000000000001",
				Name:     "Acme Enterprise Corp",
				Plan:     "enterprise",
			},
		},
		{
			name:          "[異常系] テナントIDが空の場合、InvalidArgument エラーが返る",
			tenantID:      "",
			store:         &mockStore{},
			wantErrorCode: connect.CodeInvalidArgument,
		},
		{
			name:          "[異常系] 不正なUUID形式の場合、InvalidArgument エラーが返る",
			tenantID:      "invalid-uuid",
			store:         &mockStore{},
			wantErrorCode: connect.CodeInvalidArgument,
		},
		{
			name:     "[異常系] テナントが存在しない場合、NotFound エラーが返る",
			tenantID: "00000000-0000-0000-0000-000000000000",
			store: &mockStore{
				getTenantByIDFn: func(ctx context.Context, id uuid.UUID) (db.Tenant, error) {
					return db.Tenant{}, pgx.ErrNoRows
				},
			},
			wantErrorCode: connect.CodeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := handler.NewAuthHandler(tt.store)
			req := connect.NewRequest(&authv1.GetTenantRequest{
				TenantId: tt.tenantID,
			})
			res, err := h.GetTenant(ctx, req)

			if tt.wantErrorCode != 0 {
				require.Error(t, err)
				require.Equal(t, tt.wantErrorCode, connect.CodeOf(err))
				return
			}
			require.NoError(t, err)

			if diff := cmp.Diff(tt.wantResponse, res.Msg, protocmp.Transform()); diff != "" {
				t.Errorf("予期しないレスポンスです: -want +got\n%s", diff)
			}
		})
	}
}

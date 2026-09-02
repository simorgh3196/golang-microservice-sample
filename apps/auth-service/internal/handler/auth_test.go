package handler_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/simorgh3196/golang-microservice-sample/apps/auth-service/internal/handler"
	authv1 "github.com/simorgh3196/golang-microservice-sample/pkg/gen/agentforge/auth/v1"
)

func TestAuthHandler_ValidateApiKey(t *testing.T) {
	t.Parallel()

	h := handler.NewAuthHandler()
	ctx := context.Background()

	tests := []struct {
		name          string
		apiKey        string
		wantErrorCode connect.Code
		wantResponse  *authv1.ValidateApiKeyResponse
	}{
		{
			name:   "[正常系] 有効なAPIキーの場合、認証に成功してテナントが返る",
			apiKey: "test-agentforge-key-123",
			wantResponse: &authv1.ValidateApiKeyResponse{
				IsValid:  true,
				KeyId:    "key_test_01",
				TenantId: "018f2b34-8c7a-7b3f-8000-000000000001",
				Role:     "admin",
			},
		},
		{
			name:   "[準正常系] 無効なAPIキーの場合、IsValid=false が返る",
			apiKey: "invalid-api-key",
			wantResponse: &authv1.ValidateApiKeyResponse{
				IsValid: false,
			},
		},
		{
			name:          "[異常系] APIキーが空の場合、InvalidArgument エラーが返る",
			apiKey:        "",
			wantErrorCode: connect.CodeInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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

	h := handler.NewAuthHandler()
	ctx := context.Background()

	tests := []struct {
		name          string
		tenantID      string
		wantErrorCode connect.Code
		wantResponse  *authv1.GetTenantResponse
	}{
		{
			name:     "[正常系] 有効なテナントIDの場合、テナント詳細が返る",
			tenantID: "018f2b34-8c7a-7b3f-8000-000000000001",
			wantResponse: &authv1.GetTenantResponse{
				TenantId: "018f2b34-8c7a-7b3f-8000-000000000001",
				Name:     "Acme Enterprise Corp",
				Plan:     "enterprise",
			},
		},
		{
			name:          "[異常系] テナントIDが空の場合、InvalidArgument エラーが返る",
			tenantID:      "",
			wantErrorCode: connect.CodeInvalidArgument,
		},
		{
			name:          "[異常系] テナントIDが存在しない場合、NotFound エラーが返る",
			tenantID:      "not-found",
			wantErrorCode: connect.CodeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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

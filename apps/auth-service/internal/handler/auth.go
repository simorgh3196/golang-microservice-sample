package handler

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	authv1 "github.com/simorgh3196/golang-microservice-sample/pkg/gen/agentforge/auth/v1"
	"github.com/simorgh3196/golang-microservice-sample/pkg/gen/agentforge/auth/v1/authv1connect"
)

var _ authv1connect.AuthServiceHandler = (*AuthHandler)(nil)

// AuthHandler は認証及びテナント管理を行う Connect-RpPC ハンドラーです
type AuthHandler struct {
	authv1connect.UnimplementedAuthServiceHandler
}

// NewAuthHandler は AuthHandler の新しいインスタンスを生成します
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// ValidateApiKey はAPIキーを検証し、有効であればテナント情報と権限ロールを返します
func (h *AuthHandler) ValidateApiKey(
	ctx context.Context,
	req *connect.Request[authv1.ValidateApiKeyRequest],
) (*connect.Response[authv1.ValidateApiKeyResponse], error) {
	apiKey := req.Msg.GetApiKey()
	if apiKey == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("api_key is required"))
	}

	if apiKey != "test-agentforge-key-123" {
		return connect.NewResponse(&authv1.ValidateApiKeyResponse{
			IsValid: false,
		}), nil
	}

	res := connect.NewResponse(&authv1.ValidateApiKeyResponse{
		IsValid:  true,
		TenantId: "018f2b34-8c7a-7b3f-8000-000000000001",
		KeyId:    "key_test_01",
		Role:     "admin",
	})

	return res, nil
}

// GetTenant はテナントIDを指定してテナント詳細情報を取得します
func (h *AuthHandler) GetTenant(
	ctx context.Context,
	req *connect.Request[authv1.GetTenantRequest],
) (*connect.Response[authv1.GetTenantResponse], error) {
	tenantID := req.Msg.GetTenantId()
	if tenantID == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("tenant_id is required"))
	}

	if tenantID != "018f2b34-8c7a-7b3f-8000-000000000001" {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("tenant not found"))
	}

	res := connect.NewResponse(&authv1.GetTenantResponse{
		TenantId: tenantID,
		Name:     "Acme Enterprise Corp",
		Plan:     "enterprise",
	})
	return res, nil
}

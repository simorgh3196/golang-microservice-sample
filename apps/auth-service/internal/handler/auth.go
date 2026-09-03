package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/simorgh3196/golang-microservice-sample/apps/auth-service/internal/db"
	authv1 "github.com/simorgh3196/golang-microservice-sample/pkg/gen/agentforge/auth/v1"
	"github.com/simorgh3196/golang-microservice-sample/pkg/gen/agentforge/auth/v1/authv1connect"
	"github.com/simorgh3196/golang-microservice-sample/pkg/logging"
)

var _ authv1connect.AuthServiceHandler = (*AuthHandler)(nil)

// Store はハンドラが必要とする DB 操作を定義したインターフェースです
type Store interface {
	GetApiKeyByHash(ctx context.Context, keyHash string) (db.ApiKey, error)
	GetTenantByID(ctx context.Context, tenantID uuid.UUID) (db.Tenant, error)
}

// AuthHandler は認証及びテナント管理を行う Connect-RpPC ハンドラーです
type AuthHandler struct {
	authv1connect.UnimplementedAuthServiceHandler
	store Store
}

// NewAuthHandler は AuthHandler の新しいインスタンスを生成します
func NewAuthHandler(store Store) *AuthHandler {
	return &AuthHandler{
		store: store,
	}
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

	hash := sha256.Sum256([]byte(apiKey))
	keyHash := hex.EncodeToString(hash[:])

	key, err := h.store.GetApiKeyByHash(ctx, keyHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.WarnContext(ctx, "api key validation failed: key not found",
				slog.Any("api_key", logging.MaskedString(apiKey)),
			)

			// キーが存在しない、または無効化されている場合は 401 エラーを返します
			return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid api key"))
		}

		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to validate api key"))
	}

	// 有効期限のチェック(設定されている場合)
	if key.ExpiresAt.Valid && time.Now().After(key.ExpiresAt.Time) {
		slog.WarnContext(ctx, "api key validation failed: key expired",
			slog.Any("api_key", logging.MaskedString(apiKey)),
		)
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("api key has expired"))
	}

	// 監査ログ: 認証成功（安全にマスクしてキーを記録）
	slog.InfoContext(ctx, "api key validated successfully",
		slog.String("tenant_id", key.TenantID.String()),
		slog.String("key_id", key.ID),
		slog.Any("api_key", logging.MaskedString(apiKey)),
	)

	res := connect.NewResponse(&authv1.ValidateApiKeyResponse{
		TenantId: key.TenantID.String(),
		KeyId:    key.ID,
		Role:     key.Role,
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

	// 文字列を uuid.UUID に変換
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid tenant_id format"))
	}

	// DBからテナントを取得
	tenant, err := h.store.GetTenantByID(ctx, tenantUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("failed to get tenant"))
		}
		return nil, connect.NewError(connect.CodeInternal, errors.New(("failed to get tenant")))
	}

	res := connect.NewResponse(&authv1.GetTenantResponse{
		TenantId: tenantID,
		Name:     tenant.Name,
		Plan:     tenant.Plan,
	})
	return res, nil
}

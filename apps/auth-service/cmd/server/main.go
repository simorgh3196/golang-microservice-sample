package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/simorgh3196/golang-microservice-sample/apps/auth-service/internal/db"
	"github.com/simorgh3196/golang-microservice-sample/apps/auth-service/internal/handler"
	"github.com/simorgh3196/golang-microservice-sample/pkg/connectutil"
	"github.com/simorgh3196/golang-microservice-sample/pkg/gen/agentforge/auth/v1/authv1connect"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// グレースフルシャットダウン (SIGINT / SIGTERM を待機)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// DB 接続プールの初期化
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		logger.Error("DATABASE_URL environment variable is required")
		os.Exit(1)
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		logger.Error("Failed to create DB pool", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	// 起動時に DB への疎通確認
	if err := pool.Ping(ctx); err != nil {
		logger.Error("Failed to ping DB", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("Connected to database successfully")

	store := db.New(pool)
	authHandler := handler.NewAuthHandler(store)

	// ルーティングの登録
	mux := http.NewServeMux()
	path, h := authv1connect.NewAuthServiceHandler(
		authHandler,
		connect.WithCodec(connectutil.NewJSONCodec()),
		connect.WithInterceptors(
			connectutil.NewRecoveryInterceptor(logger),
			connectutil.NewLoggingInterceptor(logger),
		),
	)
	mux.Handle(path, h)

	// ヘルスチェック用のエンドポイント
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	addr := fmt.Sprintf(":%s", port)

	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	protocols.SetUnencryptedHTTP2(true)

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		Protocols:         protocols,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("Starting AuthService", slog.String("addr", addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// シャットダウン待機
	<-ctx.Done()
	logger.Info("Shutting down AuthService gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("AuthService stopped")
}

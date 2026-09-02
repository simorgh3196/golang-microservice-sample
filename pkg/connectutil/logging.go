package connectutil

import (
	"context"
	"log/slog"
	"time"

	"connectrpc.com/connect"
)

// NewLoggingInterceptor は各 RPC の実行結果・所要時間・ステータスコードを slog で記録するインターセプターを生成します。
func NewLoggingInterceptor(logger *slog.Logger) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			start := time.Now()
			procedure := req.Spec().Procedure

			res, err := next(ctx, req)

			duration := time.Since(start)
			durationMs := float64(duration.Microseconds()) / 1000.0

			attrs := []any{
				slog.String("procedure", procedure),
				slog.Float64("duration_ms", durationMs),
			}

			if err != nil {
				code := connect.CodeOf(err)
				attrs = append(attrs,
					slog.String("code", code.String()),
					slog.String("error", err.Error()),
				)
				logger.ErrorContext(ctx, "RPC failed", attrs...)
			} else {
				attrs = append(attrs, slog.String("code", "ok"))
				logger.InfoContext(ctx, "RPC success", attrs...)
			}

			return res, err
		}
	}
}

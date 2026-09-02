package connectutil

import (
	"context"
	"errors"
	"log/slog"
	"runtime/debug"

	"connectrpc.com/connect"
)

// NewRecoveryInterceptor は panic が発生した場合に捕捉し、適切なエラーレスポンスを返します
func NewRecoveryInterceptor(logger *slog.Logger) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (res connect.AnyResponse, err error) {
			defer func() {
				if r := recover(); r != nil {
					stack := string(debug.Stack())
					logger.ErrorContext(ctx, "Panic recoverd in RPC handler",
						slog.Any("panic", r),
						slog.String("procedure", req.Spec().Procedure),
						slog.String("stack", stack),
					)
					err = connect.NewError(connect.CodeInternal, errors.New("internal server error"))
				}
			}()

			return next(ctx, req)
		}
	}
}

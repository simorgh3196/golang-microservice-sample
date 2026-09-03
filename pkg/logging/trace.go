package logging

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// TraceHandler はコンテキストから OpenTelemetry の trace_id と span_id を抽出し、
// ログレコードに自動付与する slog.Handler のラッパーです。
type TraceHandler struct {
	slog.Handler
}

// // NewTraceHandler は指定したハンドラをラップする TraceHandler を生成します。
func NewTraceHandler(h slog.Handler) *TraceHandler {
	return &TraceHandler{Handler: h}
}

// Handle はログ出力時に context から Span 情報を取得し、trace_id と span_id をログに注入します。
func (h *TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		r.AddAttrs(
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
	}
	return h.Handler.Handle(ctx, r)
}

// WithAttrs はラップしている内部ハンドラの WithAttrs を呼び出します。
func (h *TraceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TraceHandler{Handler: h.Handler.WithAttrs(attrs)}
}

// WithGroup はラップしている内部ハンドラの WithGroup を呼び出します。
func (h *TraceHandler) WithGroup(name string) slog.Handler {
	return &TraceHandler{Handler: h.Handler.WithGroup(name)}
}

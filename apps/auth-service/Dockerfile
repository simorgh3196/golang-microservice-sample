FROM golang:1.27-bookworm AS builder

WORKDIR /workspace

# 依存モジュールの定義ファイルを先にコピーしてキャッシュを効かせる
COPY go.work go.work.sum ./
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY apps/auth-service/go.mod apps/auth-service/go.sum ./apps/auth-service/

# 依存パッケージのダウンロード
RUN go work sync

# ソースコードのコピー
COPY pkg/ ./pkg/
COPY apps/auth-service/ ./apps/auth-service/

# コンパイル
WORKDIR /workspace/apps/auth-service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /bin/auth-service \
    ./cmd/server/main.go

# Runtime Stage
FROM gcr.io/distroless/static-debian12:nonroot

USER nonroot:nonroot

COPY --from=builder --chown=nonroot:nonroot /bin/auth-service /app/auth-service

EXPOSE 8080

ENTRYPOINT ["/app/auth-service"]


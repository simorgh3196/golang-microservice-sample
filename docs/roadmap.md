# 段階的学習ロードマップ（詳細版）

本ドキュメントは [README.md](../README.md#🗺-段階的学習ロードマップ) のロードマップ概要を、Step ごとに「何を作るか」「どう壊すか」「何を説明できるようになるか」まで落とし込んだ詳細版です。

設計方針は次の 3 つです。

1. **深さ優先**: 各 Step に「必須」と「発展」を分け、必須だけ通しても一本筋が通るようにする。概念の数ではなく「一つの仕組みを障害ケース込みで説明できる深さ」を目標にする。
2. **壊してから直す**: すべての Step に **💥 障害シミュレーション（詰まり体験）** を組み込む。防御なしの状態で実際に事故を再現し、原因を自力で特定してから対策を入れる。
3. **言語化する**: Step ごとに ADR（設計判断）とポストモーテム（障害の振り返り）を書く。シニアに最も求められる「なぜそうしたかを説明する力」を鍛える。

---

## 📖 各 Step の読み方

| 記号 | 意味 |
| :--- | :--- |
| 🎯 **完了条件** | この Step を「終わった」と言える客観的な状態。全 Step 共通の条件は「必須がすべて動き、💥 を再現 → 修正 → ポストモーテム記述まで完了していること」 |
| ✅ **必須** | シニアとして外せない最小セット。後続 Step が依存する |
| ⭐ **発展** | 余力があれば取り組む。飛ばしても後続 Step は進められる |
| 💥 **障害シミュレーション** | わざと壊して詰まる体験。**再現手順だけ**が書いてあり、対策は書いていない。自力で原因を特定してから直す |
| 🔍 **深掘りの問い** | Step 終了時に、何も見ずに口頭で説明できるべき問い |
| 📝 **記録** | この Step で書く ADR / ポストモーテム |
| 📚 **一次情報** | 公式ドキュメントや原典。ブログ記事より先にこちらを読む |

---

## 🔥 障害シミュレーション（詰まり体験）の進め方

実務でシニアが信頼される理由は「事故を見たことがある」からです。本ロードマップでは、それを安全なローカル環境で意図的に再現します。

### 進行プロトコル

1. **AI が台本を出す**: 障害シナリオと再現手順のみを提示する。**対策は提示しない**。
2. **自分で壊す**: 手順どおりに再現し、症状を観察する（ログ、`pprof`、`EXPLAIN ANALYZE`、Jaeger、`pg_stat_activity` など）。
3. **仮説を立てる**: 「何が起きているか」「なぜか」を自分の言葉で AI に説明する。AI は段階的にヒントを出す（ヒント 1: 見るべき場所 → ヒント 2: 関係する仕組み → ヒント 3: 答え）。
4. **直して再実行**: 対策を実装し、同じ手順で再現しないことを確認する。
5. **ポストモーテムを書く**: [docs/postmortems/TEMPLATE.md](postmortems/TEMPLATE.md) に沿って、タイムライン・根本原因・再発防止を記録する。

### ツールボックス

| 障害の種類 | 使うもの | 例 |
| :--- | :--- | :--- |
| 依存サービスの停止 / 一時停止 | `docker compose stop` / `pause` / `unpause` | `docker compose pause postgres` |
| 遅延・帯域制限・接続切断 | [toxiproxy](https://github.com/Shopify/toxiproxy)（compose に常駐） | `toxiproxy-cli toxic add -t latency -a latency=5000 context-service` |
| プロセスの強制終了 | `kill -9` / `docker compose kill` | Saga の途中で落とす |
| DB のロック・遅いクエリ | `psql` で `BEGIN; LOCK TABLE ...;` を開いたまま放置、`pg_sleep()`、`pg_stat_activity` で観察 | マイグレーション中のロック待ち行列 |
| Redis の停止 | `docker compose pause redis` / `redis-cli CLIENT PAUSE 30000` | セッションストア喪失 |
| 負荷 | [k6](https://grafana.com/docs/k6/latest/) | キャッシュスタンピード、接続枯渇 |
| 観察 | `net/http/pprof`、`EXPLAIN (ANALYZE, BUFFERS)`、Jaeger、Grafana、構造化ログ | 原因特定はコードを読む前にこれらで行う |

### インシデント種別と再現する Step の対応表

[docs/architecture.md](architecture.md#🚨-重大インシデント防止設計) の防御パターンを、どの Step で「防御なしの事故 → 防御あり」として体験するかの対応表です。

| インシデント種別 | 再現する Step |
| :--- | :--- |
| 連鎖倒れ（Cascading Failure） | Step 6（デッドライン）, Step 8（Bulkhead）, Step 11（サーキットブレーカー） |
| データ不整合・二重実行 | Step 7（Saga）, Step 8（Idempotency-Key）, Step 11（Outbox） |
| ポイズンメッセージ | Step 11（DLQ） |
| 機密ナレッジ漏洩（Cross-Tenant） | Step 4（RLS）, Step 12（PgBouncer 越境） |
| 情報漏洩・監査不足 | Step 2（マスキング）, Step 9（監査ログ） |
| API 乱用 | Step 9（レートリミット） |
| デプロイ起因の断 | Step 13（Probe / Graceful Shutdown） |
| リソース枯渇（ディスク・メモリ） | Step 12（接続枯渇）, Step 14（予測アラート） |
| 検知の遅れ・アラート疲れ | Step 10（バーンレート）, Step 14（外形監視・抑制） |

---

## ロードマップ本体

### Step 0: 開発環境の土台作り

**状態: ✅ 完了**

- Go ワークスペース（`go.work`）と `pkg/` 共通パッケージのスケルトン作成
- Buf（`buf.yaml`, `buf.gen.yaml`）による Protobuf コード生成環境の整備
- Docker Compose による開発用 DB（PostgreSQL + `pgvector` と Redis）の起動確認

---

### Step 1: プロトコル定義

**状態: ✅ 完了**

- `proto/` に認証サービスの Protobuf スキーマ定義
- `buf generate` による型安全な Go インターフェース自動生成

---

### Step 2: 認証とテナント管理

**状態: ✅ 完了（復習課題あり）**

- `auth-service` 実装: API Key の SHA-256 ハッシュ照合、Connect-RPC ハンドラ実装
- ゼロ値も明示するカスタム JSON コーデック、`CodeUnauthenticated` エラーハンドリング
- PostgreSQL（`pgxpool` + `sqlc`）による型安全な永続化、google/uuid による UUIDv7 連携
- Func Mock パターンによる高速かつ並行な単体テスト（`golangci-lint` v2 完全パス）
- ログマスキング: `slog.LogValuer` による秘密情報（API Key 等）の平文ログ流出防止

💥 **障害シミュレーション（復習用、任意）**

1. **「DB が固まると全部固まる」**: auth-service を起動した状態で `docker compose pause postgres` を実行し、`ValidateApiKey` を curl で叩く。リクエストは返ってこない。`unpause` するまで何秒待たされるか、待っている間に goroutine 数がどうなるか（`/debug/pprof/goroutine` を有効にして）を観察する。→ この Step ではまだリクエスト単位のタイムアウトが無い。どの層で・何秒で打ち切るべきかを考え、`pkg/connectutil` にデッドライン用インターセプターを追加する（README のディレクトリ構成にある「デッドライン」がこれ）。

🔍 **深掘りの問い**

- API Key のハッシュに bcrypt / Argon2 ではなく SHA-256 を使ってよいのはなぜか（エントロピーと検索可能性の観点で）。
- `CodeUnauthenticated` と `CodePermissionDenied` の使い分けは何か。
- Func Mock パターンと `testify/mock` のトレードオフは何か。
- `MaskedString` を「ログ出力時にマスクする型」として実装したことで、何が防げて何が防げないか（例: `fmt.Println` や `%v` で出力した場合）。

---

### Step 3: 運用の土台

**テーマ: コンテナ化・最小トレーシング・スキーマ互換性ゲート・統合テスト基盤**

サービスが増える前に、全サービスに共通する「動かす・観る・壊さない」の土台を作ります。以降の Step でトレースを見ながらデバッグできるようになるのが狙いです。

🎯 **完了条件**: `docker compose up` で auth-service と Jaeger が起動し、curl で叩いた RPC のトレースが Jaeger に表示される。CI で後方互換性を壊す proto 変更が落ちる。実 DB を使った統合テストが `go test` で通る。

✅ **必須**

- `apps/auth-service/Dockerfile`（マルチステージビルド + distroless、非 root）と compose への追加（`depends_on: condition: service_healthy`）
- OpenTelemetry 最小導入: Connect インターセプターで W3C `traceparent` を伝播し、OTLP で Jaeger に送る。`slog` のハンドラで `trace_id` をログに付与する
- compose に Jaeger（OTLP 受信）と toxiproxy を追加する
- CI に `buf breaking`（main ブランチとの比較）と `govulncheck` を追加する
- 統合テスト基盤: `testcontainers-go` で PostgreSQL を起動し、`GetApiKeyByHash` を実 DB で検証する（`//go:build integration` タグで分離）。Step 4 の RLS テストの雛形になる

⭐ **発展**

- `Taskfile.yaml` で `task up` / `task test:integration` / `task trace` を整備する
- `pkg/telemetry` にリソース属性（`service.name`, `service.version`）とサンプラー設定を共通化する

💥 **障害シミュレーション**

1. **「DB より先に起動してクラッシュループ」**: compose の `depends_on` から `condition: service_healthy` を外して `docker compose up` する。auth-service が Ping 失敗で exit し、`restart: unless-stopped` により再起動を繰り返す。→ 「起動順を制御する」と「起動後に DB が一時的に落ちても復帰する」は別問題である。どちらをどこで解決するかを考える。
2. **「フィールド番号を付け替える」**: `auth.proto` の `tenant_id = 1` を `= 5` に変えてコミットし、CI（または `buf breaking --against '.git#branch=main'`）を実行する。落ちなければ、`buf breaking` の比較対象が正しく取得できていない（`fetch-depth`）。
3. **「トレースが途切れる」**: インターセプターをサーバー側にだけ入れ、curl から呼ぶ。Jaeger には単発スパンしか出ない。次に `traceparent` ヘッダーを手で付けて curl し、スパンが親に繋がることを確認する。Step 4 でサービスを跨いだときにこれが繋がるかが試金石になる。

🔍 **深掘りの問い**

- distroless イメージにはシェルが無い。コンテナ内でデバッグしたいとき、どうするか。
- `traceparent` ヘッダーの 4 つのフィールドはそれぞれ何を表すか。
- testcontainers と「compose で立てた共有 DB」をテストで使う場合、それぞれの利点と欠点は何か（並行実行、CI での再現性）。
- `buf breaking` が「壊れる変更」と判定する基準は何か。フィールドの削除・型変更・番号変更・名前変更のうち、ワイヤ互換性を壊すのはどれか。

📝 **記録**

- ADR-0004: Dockerfile のベースイメージ選定（distroless / alpine / scratch）

📚 **一次情報**

- [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/)
- [W3C Trace Context](https://www.w3.org/TR/trace-context/)
- [Buf: Breaking change detection](https://buf.build/docs/breaking/overview/)
- [testcontainers-go: PostgreSQL module](https://golang.testcontainers.org/modules/postgres/)

---

### Step 4: データ基盤と RLS

**テーマ: context-service ①（Database-per-Service、Goose マイグレーション、Row Level Security、Expand / Contract）**

🎯 **完了条件**: 「他テナントのデータを読むテスト」が必ず失敗することが `go test` で証明されている。旧バイナリを動かしたままカラムのリネームを完了できた。

✅ **必須**

- `context-service` のスキーマ設計。`tenants` テーブルは auth-service の所有物なので、context-service は `tenant_id` を「外部 ID」として保持し、FK も JOIN も張らない
- Goose によるバージョン管理マイグレーション（`db/migrations/`）。sqlc は migrations ディレクトリを直接読む
- RLS: `ENABLE ROW LEVEL SECURITY` と `FORCE ROW LEVEL SECURITY`、`CREATE POLICY ... USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid)`。アプリ用のロール（テーブルオーナーでもスーパーユーザーでもない）を作る
- `pkg/database`: トランザクション開始時に必ず `SET LOCAL app.current_tenant_id` を実行する `WithTenant(ctx, fn)` ヘルパー。Go の型（`TenantContext`）を通さないと DB に触れない設計にする
- RLS 統合テスト: テナント A で挿入 → テナント B で SELECT が 0 件、UPDATE が 0 行、INSERT がポリシー違反
- Expand / Contract の実演: `instruction_body` を `content` にリネームする 3 段階（Expand: 新カラム追加と二重書き → Migrate: バックフィル → Contract: 旧カラム削除）を、旧バイナリを動かしたまま実行する

⭐ **発展**

- Atlas の `migrate lint` で破壊的変更を CI で検知する
- マイグレーション用接続に `lock_timeout` / `statement_timeout` を設定する

💥 **障害シミュレーション**

1. **「WHERE 忘れ」**: RLS を有効化する前に、`tenant_id` 条件を忘れた `ListContexts` を書き、テナント B のデータが返ってくるテストを先に書く（赤）。RLS を入れてテストを緑にする。
2. **「`SET` と `SET LOCAL`」**: ヘルパーの `SET LOCAL` を `SET` に変え、`pool_max_conns=2` にした pgxpool で「テナント A → テナント B」の順にリクエストを送る。トランザクション外で `SET` した値は接続に残り、次にその接続を借りた別テナントのリクエストが A のデータを読む。再現できたら `SET LOCAL` に戻す。さらに、トランザクション外で `SET LOCAL` を実行すると WARNING だけ出て何も効かないことも確認する。
3. **「RLS を入れたのに漏れる」**: 接続ユーザーを `postgres`（スーパーユーザー）またはテーブルオーナーに変える。RLS を素通りして全テナントのデータが見える。
4. **「設定し忘れで全件消失」**: `SET LOCAL` を忘れた状態でクエリする。`current_setting(..., true)` は NULL を返し、比較結果が NULL になって 0 件になる。エラーにならないので「データが消えた」ように見える。エラーにするにはどうするか。
5. **「本番で止まるマイグレーション」**: 100 万行を seed した上で `ALTER TABLE ... ADD COLUMN x TEXT NOT NULL`（DEFAULT なし）を実行する。次に、別セッションで `BEGIN; SELECT ... FOR UPDATE;` を開いたまま `ALTER TABLE`（ACCESS EXCLUSIVE ロック）を打ち、その後ろに来た普通の SELECT まで待たされる「ロック待ち行列」を `pg_stat_activity` で観察する。
6. **「旧バイナリが落ちる」**: カラムを一発でリネームし、旧バイナリのまま叩く。`column does not exist` で 500 になる。Expand / Contract で解決する。

🔍 **深掘りの問い**

- `SET LOCAL` はなぜトランザクションに束縛されるべきか。コネクションプールとの関係で説明せよ。
- RLS の `USING` と `WITH CHECK` の違いは何か。
- `FORCE ROW LEVEL SECURITY` が必要なのはどういう場合か。
- PostgreSQL のロックキューで、なぜ後続の SELECT まで待たされるのか。
- Database-per-Service で FK を張らないなら、参照整合性はどこで・どう守るか。
- `ADD COLUMN ... NOT NULL DEFAULT 'x'` が高速に終わるのはなぜか。`SET NOT NULL` はなぜ遅いか。

📝 **記録**

- ADR-0005: RLS のテナント伝播方式（`SET LOCAL` / 接続ごとのロール切替 / スキーマ分離）
- ポストモーテム PM-001: `SET` によるテナント越境

📚 **一次情報**

- [PostgreSQL: Row Security Policies](https://www.postgresql.org/docs/current/ddl-rowsecurity.html)
- [PostgreSQL: Explicit Locking](https://www.postgresql.org/docs/current/explicit-locking.html)
- [Goose](https://github.com/pressly/goose)、[sqlc: Migrations](https://docs.sqlc.dev/en/latest/howto/ddl.html)
- [Expand / Contract パターン（Martin Fowler, ParallelChange）](https://martinfowler.com/bliki/ParallelChange.html)

---

### Step 5: ナレッジ検索とコンテキスト合成

**テーマ: context-service ②（pgvector HNSW、UUIDv7、ハイブリッド検索、階層型コンテキスト合成）**

🎯 **完了条件**: テナント内の類似検索が HNSW インデックス経由（`EXPLAIN` で Index Scan）で返り、企業 → プロジェクト → 個人の優先順位で合成されたプロンプトがトークン上限に収まる。

✅ **必須**

- `knowledge_docs` に `vector` 列と HNSW インデックス。主キーは UUIDv7（`uuid.NewV7`）
- 埋め込み生成は `tools/llm-mock` の `/v1/embeddings`（テキストのハッシュから決定的に生成）で開発する。実 API は任意
- ハイブリッド検索: `tsvector`（キーワード）とコサイン距離の結果を RRF（Reciprocal Rank Fusion）で合成する。1 本の SQL で書く
- 階層型コンテキスト合成エンジン: `CONTEXT_PROFILE` を scope / priority で調停し、トークン予算に収める。Connect RPC `ComposeContext` として公開する

⭐ **発展**

- Embedding キャッシュ（Redis、キーはテキストの SHA-256）
- チャンク分割戦略の比較（固定長 / 段落 / オーバーラップ）

💥 **障害シミュレーション**

1. **「インデックスが効かない」**: 1 万件投入して `ORDER BY embedding <=> $1 LIMIT 10` を `EXPLAIN ANALYZE` する（Seq Scan）。HNSW を作って Index Scan に変わることを確認する。次に、インデックスを `vector_l2_ops` で作り直して `<=>` で検索し、インデックスが使われないことを確認する。
2. **「小さいテナントほど検索結果が減る」**: 全体 1 万件のうちテナント B が 50 件しか持たない状態で、テナント B として `LIMIT 10` を検索する。RLS フィルタは HNSW の近似走査の**後**に適用されるため、返ってくる件数が 10 件に満たない。`hnsw.ef_search` を上げる、`hnsw.iterative_scan` を有効にする、テナント別パーシャルインデックスを張る、の 3 案を比較する。
3. **「UUIDv4 で挿入が遅くなる」**: v4 と v7 で 100 万件ずつ挿入し、`pgstatindex` のリーフ断片化と WAL 生成量を比較する。
4. **「トークン超過」**: 企業規約が 20 万トークンある状態でセッションを開始する。LLM モックが 400 を返す。優先度に基づく圧縮（要約 or 切り詰め）をどう設計するか。

🔍 **深掘りの問い**

- HNSW の `m` / `ef_construction` / `ef_search` とリコール・メモリ・ビルド時間のトレードオフは何か。
- RLS と近似最近傍インデックスの相性問題の本質は何か（フィルタの適用順序）。
- RRF の式（`1 / (k + rank)` の和）で、なぜスコアの正規化が不要なのか。
- UUIDv7 はタイムスタンプを含む。外部に露出する ID として使ってよいか。

📝 **記録**

- ADR-0006: ベクトル検索のテナントフィルタ戦略

📚 **一次情報**

- [pgvector README（HNSW、Filtering、Iterative Index Scans）](https://github.com/pgvector/pgvector)
- [RFC 9562: UUIDv7](https://www.rfc-editor.org/rfc/rfc9562)
- [PostgreSQL: Full Text Search](https://www.postgresql.org/docs/current/textsearch.html)

---

### Step 6: 思考ループとストリーミング

**テーマ: agent-service ①（ReAct 状態マシン、Server Streaming、Cascading Deadline、Keyset Pagination）**

🎯 **完了条件**: クライアントが切断したら 1 秒以内に LLM 呼び出しと DB クエリがキャンセルされ、goroutine 数が元に戻る。

✅ **必須**

- `tools/llm-mock`: OpenAI 互換の偽 LLM サーバー（`/v1/chat/completions`、`stream` 対応）。`?scenario=slow|429|500|hang|tool_call|loop` で挙動を切り替えられる。以降のすべての Step で LLM 依存の決定的テストに使う
- ReAct ループの状態マシン（Think → Act → Observe、最大ステップ数、停止条件）
- Connect Server Streaming `RunSession` で思考ログとトークンを逐次配信する
- Cascading Deadline: リクエストの `context` を LLM HTTP クライアント、pgx、下流 RPC に伝播する。Connect の `Connect-Timeout-Ms` ヘッダーで下流に残り時間を渡す
- 会話履歴の Keyset Pagination（`WHERE (created_at, id) < ($1, $2) ORDER BY created_at DESC, id DESC`）
- LLM 呼び出しのリトライ（429 / 5xx のみ。Exponential Backoff + Jitter。`Retry-After` を尊重）

⭐ **発展**

- 会話要約による Working Memory 圧縮
- ストリーミング中の部分結果の永続化（途中切断からの再開）

💥 **障害シミュレーション**

1. **「goroutine が帰ってこない」**: LLM を `context.Background()` で呼ぶ実装にし、`scenario=hang` でストリームを開始して curl を Ctrl-C で切る。`/debug/pprof/goroutine?debug=1` で goroutine が増え続けることを確認する。`pg_stat_activity` で DB 側のクエリも残っているか確認する。
2. **「ストリームが届かない」**: `http.ResponseController.Flush` を呼ばずに SSE を書く。クライアントには最後にまとめて届く。次に toxiproxy で帯域を絞り、遅いクライアントに対してサーバー側のメモリがどう増えるかを観察する（バックプレッシャー）。
3. **「OFFSET 地獄」**: 10 万件のメッセージで `OFFSET 90000` と Keyset を `EXPLAIN (ANALYZE, BUFFERS)` で比較する。
4. **「429 で連打」**: `scenario=429` でリトライなし → 即エラー。次にリトライを固定間隔にすると、複数クライアントが同時に再送する（thundering herd）。Jitter を入れて分散させる。
5. **「シャットダウンで切れる」**: ストリーム中に SIGTERM を送る。`server.Shutdown` は 5 秒のタイムアウト後にエラーを返し、進行中のセッションは異常終了する。長寿命ストリームと Graceful Shutdown をどう両立するか。
6. **「無限ループするエージェント」**: `scenario=loop`（常に tool_call を返す）で起動する。最大ステップ数、トークン予算、同一ツール呼び出しの検知のどれで止めるか。

🔍 **深掘りの問い**

- `context.Context` のキャンセルが pgx のクエリを実際に止める仕組みは何か（`pg_stat_activity` で確認せよ）。
- `Connect-Timeout-Ms` と `context.Deadline` の関係は何か。下流に渡す残り時間はどう計算されるか。
- Connect の Server Streaming とブラウザ向け SSE の違いは何か（プロトコル、再接続、プロキシとの相性）。
- ストリーミングのバックプレッシャーはどの層で効くか。

📝 **記録**

- ポストモーテム PM-002: goroutine リーク
- ADR-0007: LLM 呼び出しのリトライ方針

📚 **一次情報**

- [Go blog: Go Concurrency Patterns: Context](https://go.dev/blog/context)
- [Connect: Streaming](https://connectrpc.com/docs/go/streaming/)、[Connect: Timeouts](https://connectrpc.com/docs/go/deadlines/)
- [Keyset Pagination（use-the-index-luke）](https://use-the-index-luke.com/no-offset)
- [AWS Architecture Blog: Exponential Backoff And Jitter](https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/)

---

### Step 7: 分散制御

**テーマ: agent-service ②（Saga と補償トランザクション、MCP クライアントによるツール実行）**

🎯 **完了条件**: セッション開始の途中でプロセスを kill しても、クォータの「取りっぱなし」が残らない。MCP ツールがハングしてもセッション全体は制限時間内に失敗する。

✅ **必須**

- auth-service に `ReserveQuota` / `ReleaseQuota`（テナント別の同時セッション数）を追加する。proto を変更するので `buf breaking` を通す
- オーケストレーション型 Saga: Reserve → CreateSession → Start。途中失敗時は補償（Release）。再起動時に未完了 Saga を復旧するための `saga_state` テーブル
- MCP クライアント（`modelcontextprotocol/go-sdk`）で `MCP_SERVER_REGISTRY` のツールを動的に発見・実行する。ツールごとのタイムアウト、出力サイズ上限、許可リスト
- `tools/mcp-mock`: ハング・巨大出力・エラーを返す偽 MCP サーバー

⭐ **発展**

- 予約の TTL による自動回収（補償が繰り返し失敗した場合の最後の砦）
- ツール実行のサンドボックス化（別プロセス、権限分離）

💥 **障害シミュレーション**

1. **「取りっぱなしのクォータ」**: Reserve 成功直後に `kill -9` する。クォータが減ったまま戻らない。Saga の復旧処理と TTL 回収の両方で直す。
2. **「補償が失敗する」**: Release の直前に `docker compose stop auth-service` する。補償できない。補償のリトライと冪等性（同じ Release を 2 回投げても 1 回分しか戻らない）を実装する。
3. **「ツールがハング」**: ハングする MCP サーバーを登録する。セッション全体が固まる。ツール単位のデッドラインを入れる。
4. **「ツール出力で OOM」**: 100 MB を返すツールを呼ぶ。メモリが急増する。出力上限とストリーミング読み込みを入れる。

🔍 **深掘りの問い**

- Saga とローカルトランザクションの違いは何か。「補償」が「ロールバック」でないのはなぜか。
- オーケストレーション型とコレオグラフィ型の使い分けは何か。
- なぜ 2 相コミット（2PC）を使わないのか。
- ツール実行を agent-service のプロセス内で行うことのセキュリティ境界の問題は何か。

📝 **記録**

- ADR-0008: セッション開始 Saga の設計

📚 **一次情報**

- [microservices.io: Saga](https://microservices.io/patterns/data/saga.html)
- [MCP 仕様](https://modelcontextprotocol.io/specification/latest)、[MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)

---

### Step 8: BFF ゲートウェイ

**テーマ: bff-gateway（SSE、API Composition、縮退運転と Bulkhead、ハイブリッド認証、Idempotency-Key。発展: GraphQL + DataLoader）**

🎯 **完了条件**: context-service を止めてもチャットが続く。同じリクエストを二重送信してもセッションは 1 つしか作られない。ブラウザ（Vercel AI SDK の `useChat`）から対話できる。

✅ **必須**

- SSE エンドポイント（`POST /api/chat`。Vercel AI SDK が期待するストリーム形式に合わせる）
- API Composition: `errgroup` と個別タイムアウトで Auth / Context / Agent を並行に呼び出し、メモリ上で合成する
- 縮退運転: Context 障害時は「規約なし」で続行しつつ UI に警告を返す。Bulkhead（下流サービスごとの同時実行上限）
- ハイブリッド認証: ブラウザと BFF の間は `HttpOnly; Secure; SameSite` Cookie（セッションは Redis）。BFF と内部サービスの間は **auth-service が署名する短命 JWT（TTL 60 秒、`kid` 付き）**。auth-service が JWKS（公開鍵）を配布し、各サービスは公開鍵検証のみを行う。**BFF は秘密鍵を持たない**
- 冪等性: `Idempotency-Key` ヘッダー + Redis `SET NX` + レスポンスのキャッシュ（Stripe 方式）

⭐ **発展**

- GraphQL（gqlgen）で管理画面向けクエリ（セッション一覧とそれぞれのプロジェクト名）を実装し、DataLoader で N+1 を解消する。クエリ深度・複雑度の制限
- `SameSite=Lax` で防げない CSRF ケースと、Origin ヘッダー検証

💥 **障害シミュレーション**

1. **「1 サービス落ちで全滅」**: `docker compose stop context-service` する。チャットが 500 になる。縮退運転で「規約なしで継続」に変える。
2. **「遅い依存が全部を道連れ」**: toxiproxy で context-service に 10 秒の遅延を入れる。すべてのリクエストが 10 秒待ちになり、BFF の goroutine と下流接続が枯渇する。個別タイムアウトと Bulkhead で直す。
3. **「二重送信」**: 同じリクエストを curl で 2 回同時に送る。セッションが 2 つできる。Idempotency-Key で 1 つにする。
4. **「ログアウトしたのに使える」**: 内部 JWT の TTL を 1 時間にしてログアウトする。発行済みの JWT で内部 API が叩ける。短命 TTL と Redis セッション失効の役割分担を確認する。
5. **「Redis が落ちたら全員ログアウト」**: `docker compose pause redis` する。セッションが参照できない。フェイルクローズ（拒否）にするかフェイルオープンにするかを SLA と突き合わせて決める。
6. **「N+1」（⭐）**: セッション 50 件の一覧で下流 RPC が 51 回飛ぶ。Jaeger で可視化してから DataLoader で 2 回にする。

🔍 **深掘りの問い**

- Cookie と短命 JWT を使い分ける理由は何か（失効性とステートレス性）。
- 署名鍵をどこに置くべきか。署名者と検証者を分離すると何が守られるか。
- `SameSite` は CSRF をどこまで防ぐか。
- 冪等キーの保存期間はどう決めるか。キーが衝突したらどうするか。
- Bulkhead とサーキットブレーカーの違いは何か。

📝 **記録**

- ADR-0009: BFF の認証方式
- ポストモーテム PM-003: 遅い依存による道連れ

📚 **一次情報**

- [microservices.io: API Composition](https://microservices.io/patterns/data/api-composition.html)
- [Stripe: Designing robust and predictable APIs with idempotency](https://stripe.com/blog/idempotency)
- [OWASP: Session Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html)
- [Vercel AI SDK: Stream Protocols](https://ai-sdk.dev/docs/ai-sdk-ui/stream-protocol)

---

### Step 9: 外部公開ゲートウェイ

**テーマ: public-api（OpenAI 互換 API、透過的コンテキスト注入、RBAC / スコープ認可、レートリミット、監査ログ）**

🎯 **完了条件**: Cursor や `openai` SDK から `base_url` を差し替えるだけで対話できる。member キーで admin 操作をすると 403 になる。テナント別レートリミットが効く。「誰が何をしたか」を監査ログで追える。

✅ **必須**

- `POST /v1/chat/completions`（非ストリームと `stream: true` の OpenAI 互換 SSE チャンク）、`GET /v1/models`
- 透過的コンテキスト注入（API Key → テナント → `ComposeContext` → RAG）
- 管理 REST（`/v1/contexts`, `/v1/skills`, `/v1/webhooks`）と OpenAPI 定義
- `pkg/authz`: ロール（admin / member / read_only）とスコープ（`contexts:write` など）による認可ミドルウェア。Connect インターセプターで内部サービスにも伝播する
- Redis Token Bucket レートリミッター（Lua スクリプトで自前実装。テナント別・キー別）。`429` と `Retry-After`
- 監査ログ: 誰が（key_id / user）、何を（操作と対象 ID）、いつ、結果。追記専用テーブル。アプリログとは分離する

⭐ **発展**

- API Key の Redis キャッシュ（無効キーの負のキャッシュを含む）
- 監査ログのハッシュチェーンによる改ざん検知
- OpenAPI から Go / TypeScript クライアントを生成する

💥 **障害シミュレーション**

1. **「1 テナントが全体を落とす」**: k6 で 1 つの API Key から毎秒 500 リクエストを送る。DB 接続が枯渇して他テナントまで遅くなる。テナント別 Token Bucket で隔離する。
2. **「認可漏れ」**: member キーで `DELETE /v1/contexts/{id}` を叩くと 200 が返る。認可を「入れ忘れられない」構造（ルーター登録時に必須引数にする等）に変える。
3. **「誰が消した？」**: 監査ログなしでコンテキストが消えた状況を再現する。答えられない。監査ログを追加してから同じ操作をし、追跡できることを確認する。
4. **「ストリーミング形式のズレ」**: `openai` SDK の `stream=True` で `data: [DONE]` を返し忘れる、または `finish_reason` を省く。SDK 側がハングまたは例外になる。実 SDK で互換性の細部を検証する。
5. **「レートリミットの Redis が落ちた」**: `docker compose pause redis` する。フェイルオープン（全通し）にするかフェイルクローズにするか。プロセス内トークンバケットへのフォールバックを検討する。

🔍 **深掘りの問い**

- RBAC / ABAC / ReBAC の違いは何か。本サービスではなぜロール + スコープで足りるか。
- Token Bucket と Sliding Window Log の違いは何か。バーストをどう扱うか。
- 監査ログとアプリケーションログは何が違うか（用途、保持期間、改ざん耐性）。
- OpenAI 互換であることの制約は何か。独自機能（コンテキスト指定など）をどう露出するか。

📝 **記録**

- ADR-0010: 認可モデル

📚 **一次情報**

- [OpenAI API Reference: Chat Completions](https://platform.openai.com/docs/api-reference/chat)
- [Redis: Rate limiting patterns](https://redis.io/glossary/rate-limiting/)
- [OWASP: Logging Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Logging_Cheat_Sheet.html)

---

### Step 10: 可観測性の深掘り

**テーマ: Prometheus メトリクス（RED）、Grafana、SLO とアラート、pprof。Step 3 で入れた最小トレーシングを本番水準にする。ホスト / コンテナ / DB 側の監視（USE）とログ集約は [Step 14](#step-14-インフラの状態監視) で扱う**

🎯 **完了条件**: 遅延を注入したサービスを Jaeger と Grafana だけで特定できる。SLO 違反時にアラートが鳴る。

✅ **必須**

- Prometheus メトリクス（RED: Rate / Errors / Duration）を Connect インターセプターと HTTP ミドルウェアで全サービス共通化する
- Grafana ダッシュボード（サービス別 RED、DB プール使用率、goroutine 数、NATS のラグ）
- SLO（例: chat の p95 < 3 秒、可用性 99.9%）とエラーバジェット。Alertmanager のルール
- ログ・トレース・メトリクスの相関（ログに `trace_id`、Exemplar でメトリクスからトレースへ）
- `net/http/pprof` を全サービスに（内部ポートのみ）

⭐ **発展**

- 継続的プロファイリング（Pyroscope）
- OTel Collector を挟んだ tail-based sampling

💥 **障害シミュレーション**

1. **「どこが遅い？」**: toxiproxy で auth-service に 800 ms の遅延を入れる。コードを見ずに Jaeger のウォーターフォールだけで特定する。
2. **「Prometheus が死ぬ」**: メトリクスのラベルに `tenant_id` や `session_id` を入れる。時系列が爆発してメモリが増え続ける。カーディナリティの設計を直す。
3. **「アラートが鳴らない / 鳴りすぎる」**: 固定閾値のアラートと、エラーバジェットのバーンレートに基づくアラートを比較する。
4. **「サンプリングで消える」**: head-based 10% サンプリングでエラーのトレースが見つからない。tail-based に変える。

🔍 **深掘りの問い**

- RED、USE、Four Golden Signals の関係は何か。
- ヒストグラムのバケット設計と p99 の誤差の関係は何か。
- SLO とアラートのバーンレートの関係は何か。
- トレースのサンプリング戦略にはどんな選択肢があるか。

📝 **記録**

- ADR-0011: SLO 定義

📚 **一次情報**

- [Google SRE Book: Monitoring Distributed Systems](https://sre.google/sre-book/monitoring-distributed-systems/)
- [Google SRE Workbook: Alerting on SLOs](https://sre.google/workbook/alerting-on-slos/)
- [Prometheus: Histograms and summaries](https://prometheus.io/docs/practices/histograms/)

---

### Step 11: イベント駆動と Webhook

**テーマ: notification-service（Transactional Outbox、NATS JetStream、HMAC 署名 Webhook、リトライと DLQ、サーキットブレーカー）**

🎯 **完了条件**: DB コミットと NATS 停止がどの順番で起きてもイベントが失われず、幻のイベントも出ない。毒メッセージが DLQ に隔離される。Webhook 受信側が署名を検証できる。

✅ **必須**

- Transactional Outbox: agent-service が `outbox` テーブルに同一トランザクションで書き、リレー（ポーリング or `LISTEN/NOTIFY`）が JetStream に publish する
- JetStream コンシューマー（durable、ack explicit、`MaxDeliver`）。`MAX_DELIVERIES` の advisory を購読して DLQ ストリームへ退避する
- Webhook 配信: HMAC-SHA256（`t=timestamp,v1=signature` 形式）、Exponential Backoff + Jitter、配信ログ
- サーキットブレーカー（配信先ごと）
- `tools/webhook-receiver`: 署名検証とリプレイ防止を実装した受信側サンプル

⭐ **発展**

- Outbox テーブルのパーティショニングと古い行の削除
- テナント単位の subject 設計による順序保証
- at-least-once の受信側冪等化（`event_id` の記録）

💥 **障害シミュレーション**

1. **「イベント消失」**: Outbox を使わず「DB コミット → publish」の順に書き、publish 直前に `docker compose stop nats` する。セッションは完了しているのに Webhook が飛ばない。
2. **「幻の Webhook」**: 「publish → DB コミット」の順にし、コミットを失敗させる。存在しない完了通知が届く。
3. **「毒メッセージ」**: JSON が壊れたイベントを投入する。コンシューマーがクラッシュループする。DLQ に隔離する。
4. **「リトライの嵐」**: 受信先をハングさせ、100 イベントを投入する。配信 goroutine が積み上がる。サーキットブレーカーと同時配信上限で直す。
5. **「同じ Webhook が 2 回届く」**: ack の前にコンシューマーを kill する。再配信される。受信側の冪等性で吸収する。
6. **「リプレイ攻撃」**: 受信した署名付きリクエストを 1 時間後に再送する。通ってしまう。タイムスタンプの許容幅を入れる。

🔍 **深掘りの問い**

- デュアルライト問題とは何か。なぜ Outbox で解けるのか。
- at-least-once / at-most-once / exactly-once は実際にはどう実現される（されない）か。
- JetStream に「DLQ」という機能が無いのはなぜか。advisory ベースの実装で何が保証されるか。
- HMAC 署名になぜタイムスタンプが要るか。

📝 **記録**

- ポストモーテム PM-004: イベント消失
- ADR-0012: イベント配信の保証レベル

📚 **一次情報**

- [microservices.io: Transactional Outbox](https://microservices.io/patterns/data/transactional-outbox.html)
- [Debezium: Reliable Microservices Data Exchange With the Outbox Pattern](https://debezium.io/blog/2019/02/19/reliable-microservices-data-exchange-with-the-outbox-pattern/)
- [NATS JetStream](https://docs.nats.io/nats-concepts/jetstream)
- [Stripe: Webhook signatures](https://docs.stripe.com/webhooks#verify-official-libraries)

---

### Step 12: 高負荷と耐障害性

**テーマ: Singleflight、2 層キャッシュ、PgBouncer、k6 限界テストとチューニング**

🎯 **完了条件**: k6 で限界負荷を掛けたとき、キャッシュ失効・DB 接続枯渇・goroutine リークのいずれも起きない。ボトルネックを数値で説明できる。

✅ **必須**

- `pkg/resilience`: Singleflight（コンテキスト合成結果、Embedding）、Jitter リトライ、サーキットブレーカーを共通化する
- 2 層キャッシュ（プロセス内 LRU + Redis）と TTL ジッター
- PgBouncer（transaction pooling）を compose に導入し、全サービスの接続を集約する
- k6 シナリオ（[service-spec の 5 大高負荷シナリオ](service-spec.md#⚡-サービス特性に起因する5大高負荷シナリオ)）と限界テスト → pprof でボトルネック特定 → チューニングを Before / After で記録する

⭐ **発展**

- 負荷テストを CI（nightly）で回して回帰を検知する
- `GOMAXPROCS` / `GOMEMLIMIT` のコンテナ向け設定

💥 **障害シミュレーション**

1. **「キャッシュスタンピード」**: 人気キーの TTL を 10 秒にして k6 で 200 VU を掛ける。失効の瞬間に DB がスパイクする（Grafana で観察）。Singleflight と TTL ジッターで平らにする。
2. **「PgBouncer で RLS が壊れる」**: transaction pooling 下で `SET`（`LOCAL` なし）を使う。別テナントの接続に設定が混ざる。Step 4 の教訓の実戦編。
3. **「prepared statement does not exist」**: pgx の既定（拡張プロトコル + statement cache）と PgBouncer の相性で失敗する。pgx の `default_query_exec_mode` と PgBouncer の `max_prepared_statements` のどちらで解決するか。
4. **「接続枯渇」**: `pool_max_conns` を小さくして k6 を掛ける。待ち行列とタイムアウトが発生する。プールサイズの決め方（Little の法則）を考える。
5. **「goroutine リーク再訪」**: 長時間負荷の後に goroutine 数が戻らない箇所を pprof で特定する。

🔍 **深掘りの問い**

- Singleflight が効く条件と効かない条件は何か（キーの粒度）。
- TTL ジッターは何を解決するか。
- PgBouncer の 3 つのプールモードと、それぞれで使えなくなる機能は何か。
- Little の法則でプールサイズをどう算出するか。
- `GOMEMLIMIT` と OOM の関係は何か。

📝 **記録**

- 負荷試験レポート（Before / After）を `loadtests/reports/` に残す

📚 **一次情報**

- [Google SRE Book: Handling Overload](https://sre.google/sre-book/handling-overload/)、[Addressing Cascading Failures](https://sre.google/sre-book/addressing-cascading-failures/)
- [PgBouncer: Features](https://www.pgbouncer.org/features.html)
- [Go: Diagnostics](https://go.dev/doc/diagnostics)

---

### Step 13: Kubernetes 運用

**テーマ: k3d + Kustomize、Probe と Graceful Shutdown、HPA、mTLS、Distroless**

🎯 **完了条件**: `kubectl rollout restart` 中に k6 のエラー率が 0 になる。SIGTERM 時に進行中のストリームが完了してから Pod が消える。

✅ **必須**

- Kustomize（base / overlays）。Deployment / Service / Ingress / ConfigMap / Secret / HPA
- Probe 設計（startup: DB 接続、readiness: 下流疎通、liveness: プロセス生存のみ）
- Graceful Shutdown の連動（`preStop` の sleep、`terminationGracePeriodSeconds`、Go 側の Shutdown 待ち）
- サービス間 mTLS（自己署名 CA）または内部 JWT 検証の徹底
- Step 3 の Distroless イメージを再利用し、`readOnlyRootFilesystem` と非 root を強制する

⭐ **発展**

- NetworkPolicy による最小権限通信
- PodDisruptionBudget
- Argo Rollouts によるカナリアリリース

💥 **障害シミュレーション**

1. **「デプロイ中に 502」**: readiness なしで rollout する。起動直後の Pod にトラフィックが流れて失敗する。
2. **「ストリームがぶった切れる」**: `preStop` なしで rollout する。SSE が途中で切断される。k6 のストリームシナリオで検知する。
3. **「再起動の嵐」**: liveness に DB 疎通を入れる。DB 障害時に全 Pod が再起動を繰り返し、復旧が遅れる。
4. **「HPA が反応しない」**: CPU ベースの HPA でストリーミング負荷（接続数依存）を掛ける。スケールしない。カスタムメトリクスに変える。
5. **「OOMKilled の連鎖」**: resource limits なしで動かす。隣の Pod が巻き添えになる。
6. **「Secret が平文」**: ConfigMap に DB パスワードを書き、`kubectl get -o yaml` で見る。Secret と外部シークレット管理を検討する。

🔍 **深掘りの問い**

- 3 種類の Probe の役割分担は何か。liveness に依存先を入れてはいけないのはなぜか。
- Pod 終了シーケンス（SIGTERM、`preStop`、Endpoints からの削除）にはどんな競合があるか。
- HPA のメトリクスはどう選ぶか。
- mTLS と内部 JWT はどう使い分けるか。

📝 **記録**

- ADR-0013: サービス間認証（mTLS / JWT）
- ポストモーテム PM-005: デプロイ中の 502

📚 **一次情報**
- [Kubernetes: Pod Lifecycle](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/)
- [Kubernetes: Configure Liveness, Readiness and Startup Probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
- [k3d](https://k3d.io/)、[Kustomize](https://kustomize.io/)

---

### Step 14: インフラの状態監視

**テーマ: USE メトリクス（ホスト / コンテナ / Kubernetes / DB / ミドルウェア）、ログ集約、外形監視、予測アラート、アラート運用。Datadog 型の「インフラ監視」を OSS で再現する**

Step 10 は「サービスが健全か」（RED）を見た。この Step は「サービスが載っている資源が健全か」（USE）を見る。Datadog / New Relic のような SaaS は「集める・貯める・見せる・鳴らす」を 1 製品にまとめているが、ここでは層ごとに分かれた OSS で組み、仕組みを理解する。

🎯 **完了条件**: ディスク満杯・メモリリーク・DB の遅いクエリ・Ingress 断のいずれも、コードもログファイルも見ずに Grafana とアラートだけで「何が・どこで」起きているかを 5 分以内に特定できる。依存先ダウンで同時に鳴るアラートが根本原因の 1 通に抑制される。

✅ **必須**

- **収集層の設計**: Prometheus の pull 型（exporter を scrape）と、OTel Collector を DaemonSet として各ノードに置く push 型（Datadog Agent 相当）の両方を構築し、何に向くかを比較する
- **インフラメトリクス（USE）**: `node_exporter`（ホスト）、`cAdvisor`（コンテナ）、`kube-state-metrics`（Pod 再起動回数、Pending、OOMKilled、Node の圧迫）を Prometheus に取り込む。Grafana に USE ダッシュボード（Utilization / Saturation / Errors）を作る
- **データベース・ミドルウェア監視**: `postgres_exporter` と `pg_stat_statements`（接続数、ロック待ち、キャッシュヒット率、遅いクエリ上位）、`redis_exporter`、NATS の `/metrics`
- **ログ集約**: Loki と Promtail（または OTel Collector 経由）で全サービスのログを集約する。Grafana で trace_id からトレースとログを相互にジャンプできるようにし、Step 10 の相関を完成させる
- **外形監視**: `blackbox_exporter` で Ingress 経由のエンドポイントを定期的に叩き、内部メトリクスと独立した死活を見る
- **予測アラート**: `predict_linear` によるディスク・メモリの枯渇予測。閾値ではなく「あと何時間で枯渇するか」で鳴らす
- **アラート運用**: Alertmanager のルーティング（重大度別の通知先）、抑制（inhibition: 上流ダウン時に下流のアラートを止める）、グループ化、サイレンス（メンテナンス中）。各アラートに Runbook（何を見て、何をするか）へのリンクを付ける

⭐ **発展**

- Grafana ダッシュボードと Alertmanager ルールを Kustomize / ConfigMap でコード管理する（Dashboards as Code）
- 容量計画: 負荷テストの結果とメトリクスから「今のノードで何テナントまで捌けるか」を算出する
- Datadog / Grafana Cloud の無料枠に OTel Collector から同じデータを送り、SaaS と OSS の違いを体験する（任意）

💥 **障害シミュレーション**

1. **「ディスクが満杯になる」**: PostgreSQL のデータ領域と同じファイルシステムで `fallocate` により空き容量を削っていく。PostgreSQL が `No space left on device` で書けなくなり、読み取りは動くのに書き込みだけ失敗する奇妙な状態になる。`predict_linear(node_filesystem_avail_bytes[1h], 4 * 3600) < 0` で数時間前に予測アラートを出す。
2. **「メモリリークをホスト側から見つける」**: agent-service に意図的なリーク（グローバルなスライスへ追記し続ける等）を仕込んで負荷を掛ける。`container_memory_working_set_bytes` が右肩上がりになる。OOMKilled になる前に予測アラートで検知し、pprof の heap プロファイルで箇所を特定する。
3. **「内部は健全なのに外から繋がらない」**: Ingress のパスだけを壊す（タイポ）。アプリの RED メトリクスは正常のまま、`blackbox_exporter` だけが落ちる。内部監視だけでは気づけない障害があることを確認する。
4. **「どのクエリが遅いか分からない」**: `pg_stat_statements` なしで、インデックスの無いクエリを混ぜて負荷を掛ける。DB の CPU が張り付くが、どのクエリかは分からない。`pg_stat_statements` と `postgres_exporter` を入れてから同じ負荷を掛け、上位クエリを特定する。
5. **「ログが探せない」**: Loki 導入前に、5 サービスの `kubectl logs` から 1 つのリクエストを trace_id で追い、掛かった時間を計る。Loki 導入後に同じことを Grafana でやり、差を記録する。
6. **「深夜にアラートが 200 通」**: PostgreSQL を停止する。全サービスの RED アラート、DB 接続アラート、外形監視アラートが同時に鳴る。inhibition ルールで「DB ダウン」の 1 通に抑制する。
7. **「監視が監視を落とす」**: Prometheus の scrape 間隔を 1 秒にし、ラベルの多い exporter を増やす。Prometheus 自身のメモリと、監視される側の `/metrics` の負荷が増える。監視のコストを見積もる。

🔍 **深掘りの問い**

- USE と RED はどう使い分けるか。USE は「資源」に、RED は「サービス」に対して使うのはなぜか。
- pull 型（Prometheus）と push 型（Agent / OTLP）の利点と欠点は何か。短命なジョブやサーバーレスではどちらを選ぶか。
- 外形監視が内部監視で代替できないのはなぜか。
- `predict_linear` はどういう前提で予測しているか。どんな時に外れるか。
- アラートの「抑制」「グループ化」「サイレンス」の違いは何か。
- Runbook に書くべき最低限の項目は何か。
- Datadog のような SaaS を使う現場に入ったとき、この Step で学んだことのうち何がそのまま通用し、何が製品固有か。

📝 **記録**

- ADR-0014: ログ集約の方式（Loki + Promtail / OTel Collector / SaaS）
- ポストモーテム PM-006: ディスク満杯

📚 **一次情報**

- [Brendan Gregg: The USE Method](https://www.brendangregg.com/usemethod.html)
- [Prometheus: Alertmanager](https://prometheus.io/docs/alerting/latest/alertmanager/)、[node_exporter](https://github.com/prometheus/node_exporter)、[kube-state-metrics](https://github.com/kubernetes/kube-state-metrics)
- [Grafana Loki](https://grafana.com/docs/loki/latest/)
- [PostgreSQL: pg_stat_statements](https://www.postgresql.org/docs/current/pgstatstatements.html)
- [OpenTelemetry Collector: Deployment patterns（Agent / Gateway）](https://opentelemetry.io/docs/collector/deployment/)
- [Google SRE Book: Practical Alerting from Time-Series Data](https://sre.google/sre-book/practical-alerting/)

---

### Step 15: MCP サーバーと Skills 配布

**テーマ: AI-Native エコシステム（公式 MCP サーバー、Agent Skill 配布）**

🎯 **完了条件**: Claude や Cursor から `agentforge-mcp` を登録し、ナレッジ検索とセッション起動ができる。

✅ **必須**

- `agentforge-mcp`: Go 公式 SDK で MCP サーバー（stdio と Streamable HTTP）を実装する。ツールは public-api を経由し、認可・レートリミット・監査ログを共有する
- `ecosystem/skills/SKILL.md`: 外部開発者と自社フロント向けの AI 指示書
- `ecosystem/mcp/`: クライアント設定例

💥 **障害シミュレーション**

1. **「MCP 経由で認可を素通り」**: MCP サーバーが内部 RPC を直接叩く実装にする。RBAC と監査ログが抜ける。public-api 経由に統一する。
2. **「ツール結果でプロンプトインジェクション」**: ツールの返り値に指示文を混ぜる。エージェントがそれに従う。ツール結果を信頼境界の外として扱う設計を考える。

🔍 **深掘りの問い**

- MCP の stdio と HTTP トランスポートの違いは何か。
- ツール結果を信頼境界の外として扱うとは、具体的に何をすることか。

📚 **一次情報**

- [MCP 仕様](https://modelcontextprotocol.io/specification/latest)
- [OWASP Top 10 for LLM Applications](https://owasp.org/www-project-top-10-for-large-language-model-applications/)

---

### Step 16: Game Day

**テーマ: 総合障害訓練とポストモーテム。全 Step の集大成**

🎯 **完了条件**: 30 分の障害訓練で、注入された複合障害を検知 → 切り分け → 復旧 → ポストモーテム作成まで一人で完走できる。

✅ **必須**

- k3d 上でフルスタックを起動し、k6 で定常負荷を掛ける
- AI が「進行役」として障害を順に注入する台本（例: NATS 停止 → LLM 遅延 → Pod の OOM → RLS ロールの誤設定 → ディスク枯渇）。**何が起きるかは事前に知らされない**
- 検知は Alertmanager のアラートのみ（Step 10 の SLO アラートと Step 14 のインフラアラートの両方）。原因特定は Grafana（メトリクスと Loki のログ）、Jaeger、pprof のみを使う（コードは読まない縛り）
- ポストモーテム（タイムライン、根本原因、寄与要因、再発防止、アクションアイテム）
- 全 Step のポストモーテムを振り返り、「設計原則」として `docs/principles.md` にまとめる

💥 **障害シミュレーション**

- 単一障害の組み合わせで初めて起きる複合障害。例: 「Redis 停止 + レートリミットがフェイルオープン + LLM 遅延」でコストが暴走する。
- 監視の死角を突く障害。例: 「Prometheus 自身が OOM で停止 + その間に DB のディスクが枯渇」。監視が止まっていることに気づけるか（外形監視、Prometheus の self-monitoring）。

🔍 **深掘りの問い**

- `docs/principles.md` の各原則を「なぜ」で説明できるか。
- 本番で同じ障害が起きたとき、最初の 5 分で何をするか。

📚 **一次情報**

- [Google SRE Book: Postmortem Culture: Learning from Failure](https://sre.google/sre-book/postmortem-culture/)
- [Principles of Chaos Engineering](https://principlesofchaos.org/)

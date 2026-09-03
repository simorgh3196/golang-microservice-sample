# AgentForge Enterprise システム & インフラ アーキテクチャ設計書

本ドキュメントでは、エンタープライズAIエージェント基盤「**AgentForge Enterprise**」における**「① アプリケーション論理アーキテクチャ（ソフトウェア境界・通信）」** と **「② インフラ・ネットワーク構成（Kubernetes・ロードバランサー・配置）」** を明確に切り分けて解説します。

---

## 🎯 1. なぜマイクロサービスにするのか？（分割の必然性）

AIエージェントプラットフォームでは、コンポーネントごとに**負荷特性・稼働要件・セキュリティ境界が極端に異なる**ため、マイクロサービス化の必然性が極めて高くなります。

```mermaid
flowchart TD
    subgraph Reasons ["マイクロサービス分割の4大必然性"]
        R1["① 負荷・I/O特性の極端な違い\n(長時間のLLMストリーミング・思考ループ vs 高速なナレッジ検索)"]
        R2["② 機密性とセキュリティ境界の物理分離\n(全社RAG/暗号鍵・IAM vs サンドボックス内のツール実行)"]
        R3["③ 求められるSLAの違い\n(認証・API Gatewayは99.99%必須 vs バックグラウンド調査は遅延許容)"]
        R4["④ 外部障害の遮断\n(外部LLM APIやMCPツールの障害をコア業務に波及させない)"]
    end

    subgraph Services ["AgentForge サービス群"]
        AuthSvc["Auth & IAM Service\n【高可用性・テナント権限・API Key】"]
        AgentSvc["Agent Runtime Service\n【エージェント思考ループ・セッション管理】"]
        KnowledgeSvc["Knowledge & Context Service\n【全社RAG(pgvector)・Skills・AGENTS.md】"]
        NotifySvc["Event & Webhook Service\n【非同期通知・タスク完了Webhook・DLQ】"]
        PublicGateway["Public API & MCP Gateway\n【外部サンドボックス・レートリミット】"]
    end

    R2 --> AuthSvc
    R1 & R4 --> AgentSvc
    R2 --> KnowledgeSvc
    R4 --> NotifySvc
    R2 & R3 --> PublicGateway
```

---

## 🧩 2. アプリケーション論理アーキテクチャ（ソフトウェア設計）

各サービス間の責務、APIの境界、データフローを定義した論理構成です。

```mermaid
flowchart TB
    Frontend["Web App (自社チャットUI / 管理画面)"]
    AIAgent["外部 AI エージェント / Cursor / Claude"]
    ExternalCI["外部システム / CI/CD / GitHub"]

    subgraph APILayer ["API & Gateway Layer (論理境界)"]
        BFF["BFF Gateway\n(Go HTTP/2 SSE & GraphQL)\n【自社チャットUI / Vercel AI SDK直結】"]
        PublicGateway["Public API Gateway\n(Go REST / OpenAI互換)\n【/v1/chat/completions & 管理REST・レートリミット】"]
        MCPServer["Official MCP Server\n(Go / Stdio & SSE)\n【エージェント用ツールサーバー】"]
    end

    subgraph CacheLayer ["キャッシュ & レジリエンス"]
        RedisCache[("Redis\nセッションストア / レートリミット / キャッシュ")]
    end

    subgraph InternalServices ["Internal Microservices (Connect-RPC / HTTP/2)"]
        AuthSvc["Auth & IAM Service\n(内部JWT発行・JWKS公開 / API Key / RBAC / クォータ)"]
        AgentSvc["Agent Runtime Service\n(エージェント思考ループ / セッション / Singleflight)"]
        KnowledgeSvc["Context & Knowledge Service\n(階層型コンテキスト合成 / RAG / pgvector / Skills / RLS)"]
        NotifySvc["Notification & Webhook Service\n(タスク完了通知 / HMAC署名 / DLQ)"]
    end

    subgraph DataLayer ["Data & Storage Layer"]
        AuthDB[(PostgreSQL\nAuth DB - 物理分離)]
        KnowledgeDB[(PostgreSQL RLS + pgvector\nナレッジ・Skills・コンテキスト DB)]
        SessionDB[(PostgreSQL RLS\n会話履歴・メッセージ・Outbox DB)]
        EventBus[(NATS / Redis Streams\n非同期イベントバス)]
    end

    subgraph ObservabilityStack ["監視 & 計測 (SREプラクティス)"]
        OTelCollector["OTel Collector\n(Agent / DaemonSet)"]
        Exporters["Exporters\n(node / cAdvisor / kube-state / postgres / redis / blackbox)"]
        Jaeger["Jaeger (Tracing)"]
        Prometheus["Prometheus (Metrics: RED + USE)"]
        Loki["Loki (Logs)"]
        Alertmanager["Alertmanager\n(ルーティング / 抑制 / Runbook)"]
        Grafana["Grafana\n(Dashboard / trace⇄log 相関)"]
    end

    Frontend -->|"HTTPS / SSE & GraphQL (HttpOnly Cookie)"| BFF
    AIAgent -->|"MCP Protocol / SSE"| MCPServer
    ExternalCI -->|"HTTPS / REST (API Key)"| PublicGateway

    BFF -->|"セッション検証 & キャッシュ"| RedisCache
    PublicGateway -->|"レートリミット & API Keyキャッシュ"| RedisCache
    MCPServer --> PublicGateway

    BFF -->|"Connect-RPC (短命JWTの発行要求 / セッション検証)"| AuthSvc
    BFF -->|"Connect-RPC (短命JWT / メタデータ伝播)"| AgentSvc
    BFF -->|"Connect-RPC (短命JWT / メタデータ伝播)"| KnowledgeSvc
    PublicGateway -->|"Connect-RPC (スコープ検証済みコンテキスト)"| AgentSvc

    AgentSvc -->|"ナレッジ・Skills取得 (Connect-RPC)"| KnowledgeSvc
    KnowledgeSvc --> KnowledgeDB
    AgentSvc --> SessionDB
    AuthSvc --> AuthDB

    AgentSvc -.->|"タスク完了 / Outbox Event Pub"| EventBus
    EventBus -.->|"Event Sub"| NotifySvc
    NotifySvc -->|"HMAC署名付きWebhook"| ExternalCI

    InternalServices -.->|"Trace & Metrics"| OTelCollector
    BFF -.->|"Trace & Metrics"| OTelCollector
    PublicGateway -.->|"Trace & Metrics"| OTelCollector
    OTelCollector --> Jaeger
    OTelCollector --> Prometheus
    OTelCollector --> Loki
    Exporters --> Prometheus
    Prometheus --> Alertmanager
    Prometheus --> Grafana
    Loki --> Grafana
    Jaeger --> Grafana
```

---

## 🏗 3. インフラストラクチャ & ネットワーク構成（Kubernetes 物理配置）

```mermaid
flowchart TB
    InternetUser["インターネット / 外部トラフィック"]

    subgraph EdgeNetwork ["外部ネットワーク & ロードバランシング"]
        ExtLB["External L7 Load Balancer / Ingress Controller\n(Traefik / AWS ALB / Cloud LB)\n【TLS終端 / パスルーティング / 外部負荷分散】"]
    end

    subgraph K8sCluster ["Kubernetes Cluster (k3d / EKS / GKE)"]
        subgraph IngressRouting ["Ingress ルーティング"]
            IngressRule["K8s Ingress\n(/chat, /graphql ➔ BFF)\n(/v1/* ➔ Public Gateway)\n(/mcp/* ➔ MCP Server)"]
        end

        subgraph GatewayPods ["Gateway Pods (水平スケール HPA)"]
            Pod_BFF["Pod: bff-gateway (複数台)"]
            Pod_Pub["Pod: public-api (複数台)"]
            Pod_MCP["Pod: agentforge-mcp (複数台)"]
        end

        subgraph InternalLB ["内部RPC通信のロードバランシング"]
            HeadlessSVC["K8s Headless Service / CoreDNS\n【HTTP/2 クライアントサイド L7 ロードバランシング】"]
        end

        subgraph MicroservicePods ["Backend Microservice Pods (水平スケール HPA)"]
            Pod_Auth["Pod: auth-service (複数台)"]
            Pod_Agent["Pod: agent-service (複数台)"]
            Pod_Knowledge["Pod: context-service (複数台)"]
            Pod_Notify["Pod: notification-service (複数台)"]
        end

        subgraph MiddlewarePods ["クラスタ内ミドルウェア"]
            Pod_Redis["Pod: Redis (キャッシュ/レートリミット)"]
            Pod_NATS["Pod: NATS (イベントバス)"]
            Pod_PgBouncer["Pod: PgBouncer\n【DBコネクションプーリング】"]
        end
    end

    subgraph Databases ["データベース層"]
        DB_Postgres[("PostgreSQL\n(Auth / Context pgvector / Session RLS)\n※ サービスごとに論理DBを分離")]
    end

    InternetUser --> ExtLB
    ExtLB --> IngressRule
    IngressRule --> Pod_BFF
    IngressRule --> Pod_Pub
    IngressRule --> Pod_MCP

    Pod_BFF --> HeadlessSVC
    Pod_Pub --> HeadlessSVC
    HeadlessSVC -->|"HTTP/2 ラウンドロビン"| Pod_Auth
    HeadlessSVC -->|"HTTP/2 ラウンドロビン"| Pod_Agent
    HeadlessSVC -->|"HTTP/2 ラウンドロビン"| Pod_Knowledge

    Pod_Agent --> Pod_PgBouncer
    Pod_Knowledge --> Pod_PgBouncer
    Pod_Auth --> Pod_PgBouncer
    Pod_PgBouncer --> DB_Postgres

    Pod_Agent -.-> Pod_NATS
    Pod_NATS -.-> Pod_Notify
    Pod_BFF --> Pod_Redis
    Pod_Pub --> Pod_Redis
```

---

## 🚨 重大インシデント防止設計

| インシデント種別 | 想定リスク | 実装する防御パターン |
| :--- | :--- | :--- |
| **連鎖倒れ (Cascading Failure)** | 外部LLMやMCPツールの障害・遅延から全サービスが共倒れ | **Exponential Backoff + Jitter**（指数リトライ制御）<br>**サーキットブレーカー**（障害外部APIの即時遮断）<br>**デッドライン伝播**（タイムアウトの全伝播） |
| **データ不整合・二重実行** | ネットワーク切断による再試行でエージェントタスクが二重起動 | **Idempotency-Key（冪等性キー）** による重複実行ガード<br>**Transactional Outbox** によるセッション/イベントのアトミック更新 |
| **ポイズンメッセージ** | 不正なプロンプト/データでWorkerが無限クラッシュループ | **Dead Letter Queue (DLQ)** による異常メッセージの隔離 |
| **機密ナレッジ漏洩 (Cross-Tenant)** | クエリのWHERE句忘れによる他テナントの機密RAGデータ漏洩 | **PostgreSQL RLS (Row Level Security)** でDB層から完全遮断<br>Go型システムによる `TenantContext` 必須化 |
| **情報漏洩・監査不足** | ログへのプロンプト内個人情報(PII)/API Key出力 | **ログサニタイズ・マスキング**（zap/slogフィルター）<br>**改ざん不可な構造化監査ログ** の記録 |
| **API乱用・DDoS** | 外部連携APIへの大量リクエストによるリソース枯渇 | **Redis Token Bucket レートリミッター**<br>**Webhook HMAC署名検証** |
| **デプロイ起因の断** | Rolling Update 中の 502、SIGTERM による進行中ストリームの切断 | **Readiness / Startup Probe**<br>**preStop + Graceful Shutdown の連動** |
| **リソース枯渇・検知遅れ** | ディスク満杯で DB が書けない、メモリリークで OOMKilled、内部は健全なのに外から繋がらない | **USE メトリクス + `predict_linear` 予測アラート**<br>**外形監視（blackbox_exporter）**<br>**Alertmanager の抑制で根本原因 1 通に集約** |

上記の防御パターンはすべて、ロードマップの **💥 障害シミュレーション** で「防御なしの状態で事故を再現 → 防御を入れて再発しないことを確認 → ポストモーテムを書く」という順序で体得する。インシデント種別と Step の対応表は [docs/roadmap.md](roadmap.md#インシデント種別と再現する-step-の対応表) を参照。

---

## 🔄 ゼロダウンタイム運用 & マルチクラウドDR

### 1. 🔄 Expand / Contract パターン（無停止マイグレーション）

- **スキーマ安全移行**: プロンプトテンプレートやSkillsテーブルの構造変更時も、旧バージョンと新バージョンを並行運用してダウンタイムゼロで移行。
- **Kubernetes Rolling Update & Graceful Shutdown**:
  - `PreStop Hook` と Goの `SIGTERM` ハンドラにより、実行中エージェントセッションの完了を待ってからPodを安全停止。

### 2. 🌐 マルチクラウド & ディザスタリカバリ (DR)

- **クラウド中立設計**: **Kubernetes + Connect-RPC** で構築し、特定クラウドの独自サービスに依存しないポータビリティを確保する。
- **本リポジトリのスコープ外**: Terraform によるクラウドプロビジョニングと、TiDB / Turso 等の分散DBによるDR検証は、学習の深さを優先して対象外とした（[ADR-0002](adr/0002-scope-out-turso-tidb-terraform.md)）。ロードマップ完走後にクラウドへ展開する段階で、新しい ADR として再検討する。

### 3. 🔐 認証トークンの発行者と検証者の分離

- **署名鍵は auth-service だけが持つ**: BFF や Public Gateway は内部JWTの発行を auth-service に要求し、自らは署名しない。鍵がゲートウェイ層に漏れると、ゲートウェイの侵害が全サービスの侵害に直結するため。
- **検証は公開鍵のみで行う**: auth-service が JWKS（`kid` 付き公開鍵）を配布し、各マイクロサービスは公開鍵検証だけでステートレスに認証する。鍵のローテーションは `kid` の切り替えで無停止に行う。

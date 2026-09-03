# ADR-0003: DB マイグレーションに Goose を採用する

- **Status**: Accepted
- **Date**: 2026-09-03
- **関連 Step**: Step 4（データ基盤と RLS）

## Context（背景）

Step 2 の auth-service は `db/schema.sql` を `CREATE TABLE IF NOT EXISTS` で直接適用しており、マイグレーションのバージョン管理が無い。Step 4 以降で Expand / Contract（無停止マイグレーション）を学ぶには、バージョン管理されたマイグレーションが必要になる。

ドキュメント間でツールの記載が食い違っていた（技術スタック表: Atlas / Goose、旧ロードマップ Step 3: golang-migrate）。1 つに決める必要がある。

要件は次のとおり。

- Expand / Contract の 3 段階を **手で SQL を書いて** 体験できること（学習目的）。
- sqlc がマイグレーションファイルからスキーマを読めること（`schema.sql` の二重管理を避ける）。
- Go から埋め込み（`embed`）で実行でき、コンテナ起動時や統合テストで使えること。

## Options（検討した選択肢）

| 選択肢 | 利点 | 欠点 |
| :--- | :--- | :--- |
| **Goose** | up / down を SQL で明示的に書く。Go マイグレーションも書ける。`embed` 対応。sqlc が `-- +goose Up` 注釈を解釈する | 宣言的なスキーマ差分は出さない（自分で SQL を書く） |
| Atlas | スキーマを宣言的に書き、差分 SQL を自動生成。`migrate lint` で破壊的変更を検知 | 差分生成が便利すぎて「なぜこの SQL になるか」を考えなくなる。学習目的には抽象度が高い |
| golang-migrate | 最もシンプル。多くの DB に対応 | Go マイグレーションが書きにくい。メンテナンス頻度が Goose より低い |

## Decision（決定）

**Goose を採用する。Atlas は `migrate lint`（破壊的変更の検知）に限定して Step 4 の発展課題で使う。**

- 決め手 1: Expand / Contract は「どの SQL がどのロックを取るか」を理解することが本質で、SQL を手で書く Goose が学習に合う。
- 決め手 2: sqlc が Goose のマイグレーションファイルを直接読めるので、`schema.sql` との二重管理を無くせる。
- 決め手 3: `embed` で Go バイナリに同梱でき、統合テスト（testcontainers）で同じマイグレーションを流せる。

## Consequences（結果）

- **良くなること**: マイグレーションの履歴がバージョン管理される。Step 4 の 💥（ロック待ち行列、旧バイナリの破壊）を再現しやすい。
- **受け入れるトレードオフ**: 差分 SQL は自分で書く。テーブルが増えると記述量が増える。
- **将来見直す条件**: サービス数が増えてマイグレーションの記述コストが学習を妨げるようになったら、Atlas の差分生成を本採用に格上げする。

## References

- [Goose](https://github.com/pressly/goose)
- [sqlc: Handling SQL migrations](https://docs.sqlc.dev/en/latest/howto/ddl.html)
- [Atlas: migrate lint](https://atlasgo.io/versioned/lint)

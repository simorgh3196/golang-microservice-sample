# ADR-0001: アーキテクチャ決定を ADR として記録する

- **Status**: Accepted
- **Date**: 2026-09-03
- **関連 Step**: 全 Step

## Context（背景）

本リポジトリは「シニアバックエンドエンジニアに求められる実装力と判断力を身につける」ための学習リポジトリである。実装力はコードに残るが、判断力（なぜその設計にしたか、何を捨てたか）はコードに残らない。

ロードマップのレビューで、mindmap・技術スタック表・設計書の間で採否の判断が食い違っている箇所が複数見つかった（GraphQL の扱い、Turso / TiDB、Terraform、マイグレーションツールの 3 択並立）。判断を記録する場所が無いことが原因だった。

## Options（検討した選択肢）

| 選択肢 | 利点 | 欠点 |
| :--- | :--- | :--- |
| tech-selection.md を随時更新する | 一覧性が高い | 変更履歴が残らない。「いつ・なぜ変えたか」が消える |
| コミットメッセージに書く | 追加作業が無い | 検索しづらく、設計判断とコード変更が混ざる |
| ADR（1 決定 1 ファイル、不変） | 履歴が残る。判断の言語化を毎回強制する | ファイルが増える |

## Decision（決定）

**設計上の重要な決定は `docs/adr/` に ADR として記録する。**

- 形式は [template.md](template.md)（Context / Options / Decision / Consequences）。
- 一度 Accepted にしたら書き換えず、新しい ADR で置き換える。
- ロードマップの各 Step に「この Step で書く ADR」を明示する。

## Consequences（結果）

- **良くなること**: 判断の履歴が残る。Step を進めるたびに「選んだ理由と捨てた理由」を言語化する訓練になる。リポジトリがそのままポートフォリオになる。
- **受け入れるトレードオフ**: 1 Step あたり 10〜20 分の記述コストが増える。
- **将来見直す条件**: ADR が形骸化して「決定後に埋めるだけの書類」になったら、書くタイミング（迷った瞬間に書く）を見直す。

## References

- [Documenting Architecture Decisions（Michael Nygard）](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)

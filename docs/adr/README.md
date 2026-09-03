# ADR（Architecture Decision Record）

## ADR とは

**設計上の重要な決定を、1 件 1 ファイルで残す短い記録**です。Michael Nygard が 2011 年に提唱した形式で、次の 4 点を 1 ページに書きます。

1. **Context（背景）**: なぜこの決定が必要になったか。制約や前提は何か。
2. **Options（選択肢）**: 何を比較検討したか。
3. **Decision（決定）**: 何を選んだか。
4. **Consequences（結果）**: その決定で何が良くなり、何を受け入れる（トレードオフ）か。

## なぜ書くのか

- **未来の自分と他人のため**: 数か月後に「なぜ Goose なんだっけ」と思ったとき、コードには理由が書かれていない。ADR にはある。
- **言語化の訓練**: シニアエンジニアに最も求められるのは「選んだ理由と捨てた理由を説明できる」ことで、ADR はそれを毎回強制する。
- **ポートフォリオ**: 判断の履歴そのものが、技術力の証明になる。

## ルール

- **不変**: 一度 Accepted にした ADR は書き換えない。方針が変わったら新しい ADR を書き、古い方の Status を `Superseded by ADR-XXXX` にする。
- **短く**: 1 ページ。長くなるなら決定が複数混ざっている。
- **連番**: `NNNN-kebab-case-title.md`。番号は欠番を作らない。
- **書くタイミング**: ロードマップの各 Step にある 📝 の項目、または「選択肢が 2 つ以上あって迷った」瞬間。

## [docs/tech-selection.md](../tech-selection.md) との違い

| | tech-selection.md | ADR |
| :--- | :--- | :--- |
| 内容 | プロジェクト開始時の技術選定の一覧 | 開発中に都度下した個別の判断 |
| 更新 | 随時更新する | 不変。置き換えで履歴を残す |
| 粒度 | 領域ごとの比較表 | 1 決定 1 ファイル |

## 一覧

| 番号 | タイトル | Status |
| :--- | :--- | :--- |
| [0001](0001-record-architecture-decisions.md) | アーキテクチャ決定を ADR として記録する | Accepted |
| [0002](0002-scope-out-turso-tidb-terraform.md) | Turso / TiDB / Terraform を学習スコープ外とする | Accepted |
| [0003](0003-use-goose-for-migrations.md) | DB マイグレーションに Goose を採用する | Accepted |

新しい ADR は [template.md](template.md) をコピーして書く。

## 参考

- [Documenting Architecture Decisions（Michael Nygard）](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
- [ADR GitHub organization](https://adr.github.io/)

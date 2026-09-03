# ポストモーテム（障害の振り返り）

ロードマップの **💥 障害シミュレーション** で再現した障害について、本番のインシデントと同じ形式で振り返りを残す場所です。

## なぜ書くのか

- 障害を「直した」だけでは、次に同じ形の障害を見たときに気づけない。**タイムライン・根本原因・寄与要因**を言語化して初めて、パターンとして記憶に残る。
- シニアエンジニアの面接や設計レビューで最も問われるのは「どんな障害を見て、何を学んだか」である。
- 全 Step のポストモーテムを Step 16 で振り返り、`docs/principles.md`（設計原則）にまとめる。

## ルール

- **非難しない（Blameless）**: 「誰が悪いか」ではなく「どういう仕組みが事故を許したか」を書く。
- **再現手順を残す**: 同じ障害を後から再現できるように、コマンドと観察結果を書く。
- **書くタイミング**: 対策を実装し、再現しないことを確認した直後。

## 命名

`PM-NNN-kebab-case-title.md`。番号はロードマップの 📝 に対応する。

## 一覧

| 番号 | タイトル | 関連 Step |
| :--- | :--- | :--- |
| （未記入） | | |

新しいポストモーテムは [TEMPLATE.md](TEMPLATE.md) をコピーして書く。

## 参考

- [Google SRE Book: Postmortem Culture: Learning from Failure](https://sre.google/sre-book/postmortem-culture/)
- [Google SRE Book: Example Postmortem](https://sre.google/sre-book/example-postmortem/)

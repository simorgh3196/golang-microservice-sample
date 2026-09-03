# ADR-0002: Turso / TiDB / Terraform を学習スコープ外とする

- **Status**: Accepted
- **Date**: 2026-09-03
- **関連 Step**: Step 4（データ基盤）、Step 13（Kubernetes 運用）

## Context（背景）

当初の計画では、PostgreSQL に加えて Turso（エッジ分散 SQLite）と TiDB（分散 NewSQL）を「比較検証」として組み込み、Terraform でクラウド無料枠をプロビジョニングする構想だった。

ロードマップのレビューで、次の問題が見つかった。

- 3 つとも mindmap・技術スタック表・architecture.md には登場するが、**ロードマップのどの Step にも割り当てられていない**。
- ロードマップは既に 16 Step あり、1 Step あたりの主題が多い。最大のリスクは「幅が広すぎて深さが犠牲になる」ことだと評価された。
- Turso / TiDB を触っても、ローカル完結の学習環境では「動かしてみた」以上の知見（分散 DB 固有の障害、DR 切り替えの実運用）には到達しにくい。
- Terraform は対象となるクラウド資源がローカルに無く、宣言的 IaC の本質は Kustomize マニフェストで学べる。

## Options（検討した選択肢）

| 選択肢 | 利点 | 欠点 |
| :--- | :--- | :--- |
| 3 つとも残し、発展 Step を追加する | 技術の幅が広がる | 深さがさらに犠牲になる。ローカルでは浅い体験に留まる |
| Turso / TiDB を落とし、Terraform は残す | クラウド IaC の経験が得られる | 無料枠でも課金・認証設定の手間が学習の主題を邪魔する |
| 3 つとも落とし、PostgreSQL と Kustomize に集中する | RLS / pgvector / PgBouncer / Expand-Contract を一本で深掘りできる | 分散 DB とクラウド IaC の経験は別途必要になる |

## Decision（決定）

**Turso / TiDB / Terraform の 3 つを本リポジトリの学習スコープ外とし、データベースは PostgreSQL 一本、IaC は k3d + Kustomize に集中する。**

- 決め手 1: シニアとして説明できる知見は「PostgreSQL の RLS がプールで壊れる条件」のような深い一点から生まれる。DB を 3 つ触るより 1 つを壊し尽くす方が近道。
- 決め手 2: tech-selection.md での比較検討そのものは価値があるので、比較表は残し「不採用（スコープ外）」と明記する。
- 決め手 3: 削除ではなく「保留」であり、ロードマップ完走後に別 ADR で再検討できる。

## Consequences（結果）

- **良くなること**: ロードマップと設計書の整合が取れる。Step 4 / 5 / 12 で PostgreSQL の深掘り（RLS、HNSW、PgBouncer）に時間を割ける。
- **受け入れるトレードオフ**: 分散 DB（水平スケール、マルチリージョン）とクラウド IaC の経験は本リポジトリでは得られない。
- **将来見直す条件**: Step 16（Game Day）まで完走し、実クラウドに展開する段階になったら Terraform を再検討する。単一 PostgreSQL の限界（書き込みスケール）を実測で確認できたら分散 DB を再検討する。

## References

- [docs/tech-selection.md §3, §7](../tech-selection.md)
- ロードマップレビュー（2026-09-03）

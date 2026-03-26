# golangci-lint の導入方針

## 背景・課題 (Background/Problem)
- バックエンドの大規模なリファクタリングを予定しており、移行前にコード品質の基準を確立する必要がある
- lint が未導入であるため、コード品質の担保が難しい。特に、バグやセキュリティリスクの検出が不十分である。

## 決定事項 (Decision)
- golangci-lint v2 を採用し、`default: standard` をベースにlintを適宜追加する構成とした。

### 理由 (Reasons)
- `default: standard` はノイズが少なく標準的なルールであるため、導入しやすい。
- 今後のリファクタリングや新規コード追加の際に、必要に応じてルールを追加していくことで、段階的にコード品質を向上させることができる。

### 受け入れるトレードオフ (Accepted Trade-offs)
- `default: all` と比べて有用な linter を見落とす可能性がある。
- `default: all` と比べて.golangci.yamlのメンテナンスコストが発生する。

## 検討した別の選択肢 (Alternatives Considered)
- **`default: all`**: 全ての linter を有効にする構成。警告が数百件発生し、個別に妥当性を評価して対応するコストが高すぎるため、不採用とした。

## 参考 (References)
- https://golangci-lint.run/usage/configuration/

## 議論 (Discussion)
- 今後開発していく中で、必要に応じてルールを編集する。

# CLAUDE.md

このファイルはClaude Codeがこのリポジトリで作業する際のガイドです。

## ビルド・テストコマンド

```bash
go build ./...          # ビルド
go test ./...           # 全テスト実行
go test ./... -v        # 詳細出力
go vet ./...            # 静的解析
go test ./internal/solver/ -bench .  # ソルバーベンチマーク
```

## プロジェクト構成

```
main.go                          # エントリポイント
cmd/
  root.go                        # Cobraルートコマンド
  generate.go                    # generateサブコマンド（フラグ定義）
internal/
  puzzle/puzzle.go               # データ型: Puzzle, Difficulty, AlphabetSize
  regex/builder.go               # 正規表現パターン生成エンジン
  regex/difficulty.go            # 難易度別ストラテジー定義
  solver/solver.go               # 制約伝播 + バックトラッキングソルバー
  generator/generator.go         # パズル生成オーケストレーション
  display/display.go             # ターミナル表示フォーマット
```

## 技術的決定事項

技術的な課題に直面した際の意思決定について、 ./docs/design.md にまとめている。
設計について考える時に参照する。

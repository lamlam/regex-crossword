# regex-crossword

正規表現クロスワードパズルの問題生成CLIツール。

2Dグリッドの各行・各列に正規表現ヒントがあり、すべてのヒントを満たすようにセルを埋めるパズルを生成する。生成されたパズルは一意解が保証される。

## インストール

```bash
go install github.com/lamlam/regex-crossword@latest
```

または直接ビルド:

```bash
git clone https://github.com/lamlam/regex-crossword.git
cd regex-crossword
go build -o regex-crossword .
```

## 使い方

### パズルの生成

```bash
# デフォルト（3x3, medium）
regex-crossword generate

# サイズと難易度を指定
regex-crossword generate --rows 4 --cols 4 --difficulty easy

# 解答付きで表示
regex-crossword generate -r 5 -c 5 -d hard --show-solution

# シードを指定して再現可能な生成
regex-crossword generate -r 3 -c 3 --seed 42 --show-solution

# JSON形式で出力
regex-crossword generate -f json --seed 42
```

### JSON形式のパズル表示

生成したJSONパズルを表形式で表示します。

```bash
# stdinから読み込み
regex-crossword generate -f json | regex-crossword display

# ファイルから読み込み
regex-crossword display puzzle.json

# 解答付きで表示
regex-crossword display --show-solution puzzle.json

# パイプラインでの利用
cat saved-puzzle.json | regex-crossword display --show-solution
```

### generate コマンドのフラグ

| フラグ | 短縮 | デフォルト | 説明 |
|--------|------|-----------|------|
| `--rows` | `-r` | 3 | 行数 (1-8) |
| `--cols` | `-c` | 3 | 列数 (1-8) |
| `--difficulty` | `-d` | medium | 難易度 (easy / medium / hard) |
| `--seed` | `-s` | 0 (ランダム) | 再現用シード値 |
| `--format` | `-f` | text | 出力形式 (text / json) |
| `--show-solution` | | false | 解答を表示する |

### display コマンドのフラグ

| フラグ | デフォルト | 説明 |
|--------|-----------|------|
| `--show-solution` | false | 解答を表示する |

**引数**: ファイルパス（省略時はstdinから読み取り）

## 出力例

```
Regex Crossword  [4x4, Easy]
Seed: 42
Alphabet: 126MSXZ

            C1             C2             C3             C4
       ^2M[6MS][SZ]$  ^[12]S[16Z]1$  ^[MS][XZ]61$    ^[MSZ]S2S$
     +---------------+---------------+---------------+---------------+
R1   |               |               |               |               |  ^[2Z]1[SZ]M$
     +---------------+---------------+---------------+---------------+
R2   |               |               |               |               |  ^MS[MX][12S]$
     +---------------+---------------+---------------+---------------+
R3   |               |               |               |               |  ^6[6S][6SZ]2$
     +---------------+---------------+---------------+---------------+
R4   |               |               |               |               |  ^[1Z]11[6SX]$
     +---------------+---------------+---------------+---------------+
```

## 難易度

- **Easy**: リテラル文字が多く、小さな文字クラス（2要素）。シンプルな選択パターン
- **Medium**: やや広い文字クラス、量指定子 `{n}`、グループ化
- **Hard**: 否定文字クラス `[^...]`、ドット、より広い文字クラス

## 仕様

- 文字種: パズルごとにA-Z, 0-9からランダムに選ばれたサブセット（4〜10文字）
- グリッドサイズ: 最大8x8
- 正規表現: Go標準 `regexp`（RE2構文）
- 一意解保証: ソルバーによる検証済み

## 開発

```bash
# テスト
go test ./...

# ベンチマーク
go test ./internal/solver/ -bench .

# 静的解析
go vet ./...
```

# dot-claude-sync

`.claude`ディレクトリを複数のプロジェクト間で同期するCLIツール

## 概要

git worktreeを使った開発では、`.claude`ディレクトリの内容（プロンプト、コマンド、スキルなど）を各worktree間で共有するのが面倒です。`dot-claude-sync`はこの問題を解決し、複数のプロジェクト間で`.claude`の内容を簡単に同期できます。

## インストール

```bash
go install github.com/yugo-ibuki/dot-claude-sync@latest
```

または、ソースからビルド：

```bash
git clone https://github.com/yugo-ibuki/dot-claude-sync.git
cd dot-claude-sync
go build
```

## クイックスタート

### 1. 設定ファイルの作成

```bash
dot-claude-sync init
```

または手動で作成：

```bash
mkdir -p ~/.config/dot-claude-sync
vim ~/.config/dot-claude-sync/config.yaml
```

### 2. 設定例

```yaml
groups:
  web-projects:
    paths:
      main: ~/projects/main/.claude
      feature-a: ~/projects/feature-a/.claude
      feature-b: ~/projects/feature-b/.claude
    priority:
      - main  # 最優先（重複時はこのファイルを採用）
```

### 3. 同期実行

```bash
dot-claude-sync push web-projects
```

## コマンド

| コマンド | 説明 |
|---------|------|
| `init` | 設定ファイルを対話的に作成 |
| `detect <dir> --group <name>` | git worktreeから`.claude`を自動検出してグループに追加 |
| `push <group>` | グループ内のすべてのプロジェクトでファイルを同期 |
| `rm <group> <path>` | グループ内のすべてのプロジェクトからファイルを削除 |
| `mv <group> <from> <to>` | グループ内のすべてのプロジェクトでファイルを移動/リネーム |
| `list [group]` | グループ一覧、または特定グループの詳細を表示 |
| `config <subcommand>` | 設定の管理（グループやプロジェクトの追加/削除など） |

### グローバルオプション

```bash
--config <path>   # 設定ファイルのパスを指定
--dry-run         # 実行のシミュレーション（変更なし）
--verbose         # 詳細なログを出力
--force           # 確認プロンプトをスキップ
```

## よくある使い方

### git worktreeの自動検出

```bash
# worktreeから.claudeディレクトリを自動検出してグループに追加
dot-claude-sync detect ~/projects/my-app --group my-app

# 確認
dot-claude-sync list my-app

# 同期
dot-claude-sync push my-app
```

### ファイルの配布

```bash
# メインプロジェクトで新しいプロンプトを作成
cd ~/projects/main/.claude/prompts
vim new-feature.md

# グループ全体に配布
dot-claude-sync push web-projects
```

### ファイルの削除

```bash
# 削除前に確認
dot-claude-sync rm web-projects prompts/old.md --dry-run

# 実行
dot-claude-sync rm web-projects prompts/old.md
```

### 設定の管理

```bash
# 新しいグループを作成
dot-claude-sync config add-group mobile-projects

# プロジェクトを追加
dot-claude-sync config add-project mobile-projects ios ~/projects/ios-app/.claude
dot-claude-sync config add-project mobile-projects android ~/projects/android-app/.claude

# 優先順位を設定
dot-claude-sync config set-priority mobile-projects ios android

# 確認
dot-claude-sync config show mobile-projects
```

## 優先順位のルール

- `priority`リストの順序で優先度が決まる
- リストに含まれないプロジェクトは最低優先度
- `priority`が指定されていない場合は、`paths`の順序が優先度になる
- 重複するファイル名は、高優先度のプロジェクトのファイルで上書きされる

## 設定ファイルの場所

デフォルト: `~/.config/dot-claude-sync/config.yaml`

`--config`フラグで別の設定ファイルを指定可能

## 注意事項

- 初回実行前に`.claude`ディレクトリのバックアップを推奨
- `rm`コマンドは取り消せないため、`--dry-run`で事前確認を推奨
- ファイルの重複時は優先度の高いプロジェクトのファイルで上書きされる

## アンインストール

```bash
# バイナリの削除
rm $(which dot-claude-sync)

# 設定ディレクトリの削除
rm -rf ~/.config/dot-claude-sync
```

## ライセンス

MIT

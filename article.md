# claude-sync: 複数プロジェクトの.claudeディレクトリを同期するCLIツール

## TL;DR

- `.claude`ディレクトリを複数のプロジェクト/worktree間で同期するGo製CLIツール
- グループ管理と優先度システムで柔軟な同期戦略を実現
- git worktree環境でのClaude Code活用を劇的に改善

## はじめに

Claude Codeを使っていると、`.claude`ディレクトリに便利なプロンプトやスキル、プロジェクトのコンテキストを保存することが多くなります。しかし、git worktreeで複数ブランチを同時に開発していると、`.claude`はgitignoreされているため、各worktree間で同期されません。

**claude-sync**は、この問題を解決するために開発したツールです。

## 背景と課題

### Claude Codeでの.claude活用

Claude Codeを使った開発では、以下のような情報を`.claude`ディレクトリに保存すると便利です：

- プロジェクト固有のプロンプト
- よく使うスキルやコマンド
- 実装仕様書やTODOリスト
- プロジェクトのコンテキスト情報

これらは**gitで管理しない**（gitignore）ことで、リポジトリを汚さずにClaude専用の長期コンテキストとして活用できます。

### git worktreeでの課題

しかし、git worktreeを活用した開発フローでは問題が発生します：

```bash
my-project/
├── main/           # メインworktree
│   └── .claude/
│       └── prompts/useful-prompt.md
├── feature-a/      # feature-a worktree
│   └── .claude/    # 空っぽ！
└── feature-b/      # feature-b worktree
    └── .claude/    # 空っぽ！
```

- worktree間で`.claude`の内容が共有されない
- 各worktreeで個別に設定が必要
- 便利なプロンプトを都度コピーする手間

## claude-syncの解決策

claude-syncは「**グループ**」という概念でプロジェクトをまとめ、**優先度システム**で同期戦略を制御します。

### 基本的な使い方

```bash
# 1. インストール
go install github.com/yugo-ibuki/dot-claude-sync@latest

# 2. 初期設定（対話式）
claude-sync init

# 3. worktreeの.claudeディレクトリを自動検出
claude-sync detect ~/projects/my-app --group my-app

# 4. 同期実行
claude-sync push my-app
```

### 設定ファイルの例

`~/.config/claude-sync/config.yaml`:

```yaml
groups:
  my-app:
    paths:
      main: ~/projects/my-app/main/.claude
      feature-a: ~/projects/my-app/feature-a/.claude
      feature-b: ~/projects/my-app/feature-b/.claude
    priority:
      - main  # mainを最優先（マスター設定）
```

この設定で`claude-sync push my-app`を実行すると：

1. 全プロジェクトから`.claude`ディレクトリのファイルを収集
2. ファイル名が重複する場合は、優先度の高いプロジェクト（main）のファイルを採用
3. 収集したファイルを全プロジェクトに配布

## 主要機能

### 🔍 detect - worktree自動検出

```bash
claude-sync detect ~/projects/my-app --group my-app
```

`git worktree list`を実行し、各worktreeの`.claude`ディレクトリを自動検出して設定に追加します。手動で各パスを設定する手間が省けます。

### 📤 push - 同期実行

```bash
claude-sync push my-app
```

グループ内の全プロジェクトから`.claude`ファイルを収集し、優先度に基づいて配布します。

### 🗑️ rm - 一括削除

```bash
claude-sync rm my-app prompts/old-prompt.md
```

グループ内の全プロジェクトから指定ファイルを削除します。

### 📝 mv - 一括リネーム

```bash
claude-sync mv my-app old-name.md new-name.md
```

グループ内の全プロジェクトでファイルをリネーム/移動します。

### ⚙️ config - 設定管理

```bash
# グループ追加
claude-sync config add-group new-group

# プロジェクト追加
claude-sync config add-project my-app feature-c ~/projects/my-app/feature-c/.claude

# 優先度設定
claude-sync config set-priority my-app main feature-a feature-b feature-c
```

コマンドラインから設定ファイルを直接編集できます。

## 実装のポイント

### Go + CobraでのCLI設計

```go
// cmd/root.go
var rootCmd = &cobra.Command{
    Use:   "claude-sync",
    Short: "Synchronize .claude directories across projects",
}

func init() {
    rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Preview changes")
    rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Verbose output")
    rootCmd.PersistentFlags().BoolVar(&force, "force", false, "Skip confirmations")
}
```

全コマンド共通のグローバルフラグ（`--dry-run`, `--verbose`, `--force`）を提供しています。

### 柔軟な優先度システム

```go
// config/config.go
type Group struct {
    Paths    interface{} `yaml:"paths"`    // map[string]string or []string
    Priority []string    `yaml:"priority"` // optional
}
```

パスの指定方法を2通りサポート：

1. **エイリアス付き**（map形式）: 可読性が高く、優先度指定が簡単
2. **シンプル**（配列形式）: 素早く設定できる

優先度ルール：

- `priority`が指定されていればその順序
- なければ`paths`の順序をデフォルト優先度とする

### git worktree統合

```bash
# detect commandの内部処理
git worktree list --porcelain
# ↓ 各worktreeパスを取得
# ↓ 各worktree/.claudeの存在確認
# ↓ 設定ファイルに自動追加
```

`git worktree list`の出力をパースし、`.claude`ディレクトリを自動検出します。

## ユースケース

### ケース1: worktree環境の即座のセットアップ

```bash
# worktreeプロジェクトの.claudeディレクトリを自動検出・追加
claude-sync detect ~/projects/my-app --group my-app

# すぐに同期開始
claude-sync push my-app
```

### ケース2: 共通設定の配布

```yaml
groups:
  web-projects:
    paths:
      shared: ~/projects/shared-config/.claude  # 共通設定マスター
      frontend: ~/projects/frontend/.claude
      backend: ~/projects/backend/.claude
    priority:
      - shared  # shared が最優先
```

```bash
# shared プロジェクトの設定を全プロジェクトに配布
claude-sync push web-projects
```

### ケース3: クライアントプロジェクトのテンプレート管理

```yaml
groups:
  clients:
    paths:
      template: ~/templates/client/.claude
      client-a: ~/clients/a/.claude
      client-b: ~/clients/b/.claude
    priority:
      - template
```

テンプレートプロジェクトの設定を各クライアントプロジェクトに展開できます。

## 技術スタック

- **言語**: Go 1.23+
- **CLI**: github.com/spf13/cobra
- **設定**: YAML (gopkg.in/yaml.v3)
- **テスト**: Go標準testing + testify
- **CI/CD**: GitHub Actions

## 今後の展開

現在の実装は基本機能に焦点を当てていますが、以下の機能を検討中です：

- [ ] 双方向同期（conflict resolution）
- [ ] ファイル差分表示（push前のプレビュー）
- [ ] 選択的同期（特定ファイルのみ同期）
- [ ] 除外パターン（.gitignoreライクな仕組み）
- [ ] rollback機能（誤操作からの復元）

## まとめ

claude-syncは、git worktree環境でのClaude Code活用を改善する小さなツールです。

**こんな人におすすめ**：

- ✅ git worktreeを活用している
- ✅ 複数プロジェクトで共通の.claude設定を使いたい
- ✅ プロンプトやスキルの管理を効率化したい

興味を持っていただけた方は、ぜひ試してみてください！

## リンク

- GitHub: https://github.com/yugo-ibuki/dot-claude-sync
- インストール: `go install github.com/yugo-ibuki/dot-claude-sync@latest`

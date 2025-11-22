# claude-sync 実装TODO

## 📑 目次

1. [実装状況サマリー](#実装状況サマリー)
2. [現在の状況](#現在の状況)
3. [未実装機能](#未実装機能)
   - [1. Core Sync Logic (syncer/)](#1-core-sync-logic-syncer)
   - [2. Utility Functions (utils/)](#2-utility-functions-utils)
   - [3. Commands Implementation](#3-commands-implementation)
   - [4. Error Handling & Edge Cases](#4-error-handling--edge-cases)
   - [5. Testing](#5-testing)
   - [6. Documentation](#6-documentation)
   - [7. Future Enhancements](#7-future-enhancements-優先度低)
4. [実装優先順位](#実装優先順位)
5. [進捗トラッキング](#進捗トラッキング)

---

## 実装状況サマリー

- ✅ **実装済み**: 基本構造、設定管理、init/list/pushコマンド (60%)
- 🚧 **未実装**: rm/mvコマンド、テスト、ドキュメント (40%)

---

## 現在の状況

### ✅ 実装完了

#### プロジェクト構造
```
claude-sync/
├── main.go                ✅ エントリーポイント
├── cmd/
│   ├── root.go           ✅ ルートコマンド、グローバルフラグ
│   ├── init.go           ✅ 設定ファイル初期化（完全動作）
│   ├── list.go           ✅ グループ一覧・詳細表示（完全動作）
│   ├── push.go           ✅ 完全動作（収集・競合解決・配布）
│   ├── rm.go             🚧 スケルトンのみ（TODOコメント付き）
│   └── mv.go             🚧 スケルトンのみ（TODOコメント付き）
├── config/
│   └── config.go         ✅ YAML設定読み込み、優先度解決
├── syncer/
│   ├── collector.go      ✅ ファイル収集ロジック
│   ├── resolver.go       ✅ 競合解決ロジック
│   └── syncer.go         ✅ ファイル配布ロジック
├── utils/
│   └── file.go           ✅ ファイル操作（コピー、削除、移動）
├── spec/
│   ├── doc.md            ✅ 仕様書
│   └── todo.md           ✅ このファイル
├── README.md             ✅ ユーザードキュメント
├── .gitignore            ✅ Git除外設定
└── go.mod                ✅ 依存関係管理
```

#### 機能別実装状況

| 機能 | 状態 | 進捗 | 備考 |
|------|------|------|------|
| **設定管理** | ✅ | 100% | YAML読み込み、優先度解決完了 |
| **グローバルフラグ** | ✅ | 100% | --config, --dry-run, --verbose, --force |
| **initコマンド** | ✅ | 100% | インタラクティブ設定作成、完全動作 |
| **listコマンド** | ✅ | 100% | グループ一覧・詳細表示、完全動作 |
| **pushコマンド** | ✅ | 100% | 収集・競合解決・配布、完全動作 |
| **rmコマンド** | 🚧 | 20% | 引数解析のみ、ロジック未実装 |
| **mvコマンド** | 🚧 | 20% | 引数解析のみ、ロジック未実装 |
| **ファイル収集** | ✅ | 100% | syncer/collector.go 完成 |
| **競合解決** | ✅ | 100% | syncer/resolver.go 完成 |
| **ファイル配布** | ✅ | 100% | syncer/syncer.go 完成 |
| **ファイル操作** | ✅ | 100% | utils/file.go 完成 (IsDirectory, FormatSize追加) |
| **確認プロンプト** | ✅ | 100% | cmd/rm.go内で実装済み |
| **テスト** | ❌ | 0% | 全パッケージでテスト未実装 |

#### 動作確認済みコマンド

```bash
# ✅ 完全動作
claude-sync init              # 設定ファイル初期化
claude-sync list              # グループ一覧表示
claude-sync list <group>      # グループ詳細表示
claude-sync push <group>      # ファイル同期（収集・競合解決・配布）
claude-sync rm <group> <path> # ファイル/ディレクトリ削除（完全実装）
claude-sync --help            # ヘルプ表示
claude-sync --version         # バージョン表示

# ✅ 完全動作
claude-sync mv <group> <from> <to>  # ファイル移動・リネーム

# 🚧 スケルトンのみ（引数解析は動作）
claude-sync rm <group> <path> # 実行はできるが何もしない
claude-sync mv <group> <from> <to>  # 実行はできるが何もしない
```

### 🎯 次のマイルストーン

**目標**: `claude-sync mv`コマンドを実装する

**必要な実装**:
1. cmd/mv.go (移動ロジック実装)

**推定工数**: 2-3時間

---

## 未実装機能

### 1. Core Sync Logic (syncer/)

#### 1.1 ファイル収集 (syncer/collector.go)
- [x] `Collector` 構造体の定義
- [x] `.claude`ディレクトリ配下のファイルリストアップ
- [x] 相対パスの正規化（`.claude/`からの相対パス）
- [x] ファイル情報の収集（パス、サイズ、ハッシュ）
- [x] エラーハンドリング（存在しないディレクトリ、読み込み権限）

**主要メソッド:**
```go
type FileInfo struct {
    RelPath string  // .claudeからの相対パス
    AbsPath string  // 絶対パス
    Project string  // プロジェクトエイリアス
    Priority int    // 優先度
}

func CollectFiles(projects []ProjectPath) ([]FileInfo, error)
```

#### 1.2 競合解決 (syncer/resolver.go)
- [x] `Resolver` 構造体の定義
- [x] 同名ファイルのグルーピング
- [x] 優先度による競合解決
- [x] 解決結果のレポート生成

**主要メソッド:**
```go
type ResolvedFile struct {
    RelPath  string
    Source   string  // 採用されたプロジェクト
    Priority int
}

type Conflict struct {
    RelPath    string
    Candidates []FileInfo
    Resolved   FileInfo
}

func ResolveConflicts(files []FileInfo) ([]ResolvedFile, []Conflict, error)
```

#### 1.3 ファイル配布 (syncer/syncer.go)
- [x] `Syncer` 構造体の定義
- [x] ファイルコピーロジック
- [x] ディレクトリ作成
- [x] 既存ファイルの上書き
- [x] dry-runモード対応
- [x] 進捗表示

**主要メソッド:**
```go
type SyncResult struct {
    Project      string
    NewFiles     int
    Overwritten  int
    Failed       int
    Errors       []error
}

func SyncFiles(resolved []ResolvedFile, projects []ProjectPath, dryRun bool) ([]SyncResult, error)
```

---

### 2. Utility Functions (utils/)

#### 2.1 ファイル操作 (utils/file.go)
- [x] ファイルコピー
- [x] ディレクトリ再帰的コピー
- [x] ファイル削除
- [x] ディレクトリ再帰的削除
- [x] ファイル移動/リネーム
- [x] ファイル存在チェック
- [x] ディレクトリ作成（親ディレクトリ含む）
- [x] ファイルハッシュ計算（競合検出用）

**主要関数:**
```go
func CopyFile(src, dst string) error
func CopyDir(src, dst string) error
func RemoveFile(path string) error
func RemoveDir(path string) error
func MoveFile(src, dst string) error
func EnsureDir(path string) error
func FileExists(path string) bool
func FileHash(path string) (string, error)
```

#### 2.2 確認プロンプト
- [x] Yes/No確認プロンプト
- [x] 削除確認プロンプト（ファイルリスト表示）
- [x] forceフラグ対応

**注**: rmコマンド内で直接実装済み（utils/prompt.go は不要）

---

### 3. Commands Implementation

#### 3.1 push コマンド (cmd/push.go)
**現在の状態**: ✅ 完全動作

**実装タスク**:
- [x] ファイル収集フェーズの実装
  - [x] グループ内の全プロジェクトから収集
  - [x] 収集結果の表示
- [x] 競合解決フェーズの実装
  - [x] 同名ファイルの検出
  - [x] 優先度による解決
  - [x] 競合レポート表示
- [x] 配布フェーズの実装
  - [x] 全プロジェクトへの配布
  - [x] 進捗表示
  - [x] エラーハンドリング
- [x] サマリー表示
  - [x] 総ファイル数
  - [x] 各プロジェクトの新規/上書きファイル数
  - [x] エラーサマリー

**期待される出力例**:
```
Collecting files...
✓ web: 10 files (priority: 1)
✓ mobile: 8 files (priority: 2)
✓ admin: 5 files (priority: 3)

Resolving conflicts...
- config.json: using web (priority: 1)
- prompts/coding.md: using web (priority: 1)

Syncing...
✓ Synced to web (3 new files)
✓ Synced to mobile (5 new files, 2 overwritten)
✓ Synced to admin (8 new files, 1 overwritten)

Summary: 15 unique files synced across 3 projects
```

#### 3.2 rm コマンド (cmd/rm.go)
**現在の状態**: ✅ 完全実装

**実装タスク**:
- [x] ファイル検索ロジック
  - [x] 各プロジェクトで指定パスを検索
  - [x] 存在するファイルをリスト化
- [x] 削除確認プロンプト
  - [x] 削除対象ファイルのリスト表示
  - [x] forceフラグ対応
- [x] 削除実行
  - [x] ファイル/ディレクトリ削除
  - [x] 各プロジェクトでの削除結果表示
  - [x] エラーハンドリング
- [x] サマリー表示

**期待される出力例**:
```
This will delete from 'frontend' group:
- ./packages/web/.claude/prompts/old-prompt.md
- ./packages/mobile/.claude/prompts/old-prompt.md

Continue? [y/N]: y

✓ Deleted from web
✓ Deleted from mobile
✗ Not found in admin (skipped)

Summary: 2 files deleted
```

#### 3.3 mv コマンド (cmd/mv.go)
**現在の状態**: ✅ 完全動作

**実装タスク**:
- [x] ファイル検索ロジック
  - [x] 各プロジェクトで移動元パスを検索
- [x] 移動確認プロンプト
  - [x] 移動対象のリスト表示
  - [x] 移動先の衝突チェック
  - [x] forceフラグ対応
- [x] 移動実行
  - [x] ファイル/ディレクトリ移動
  - [x] 移動先ディレクトリの作成
  - [x] 各プロジェクトでの移動結果表示
  - [x] エラーハンドリング
- [x] サマリー表示

**期待される出力例**:
```
This will rename in 'frontend' group:
prompts/old.md → prompts/new.md

Continue? [y/N]: y

✓ Moved in web
✓ Moved in mobile
✗ Source not found in admin (skipped)

Summary: 2 files moved
```

---

### 4. Error Handling & Edge Cases

#### 4.1 エラーハンドリング
- [ ] 設定ファイルの検証
  - [ ] パスの存在チェック
  - [ ] パスの読み込み権限チェック
- [ ] 一部プロジェクトでの失敗時の継続処理
- [ ] 詳細なエラーメッセージ
- [ ] dry-runモードでのシミュレーション

#### 4.2 エッジケース
- [ ] 空ディレクトリの処理
- [ ] シンボリックリンクの処理
- [ ] 隠しファイルの処理
- [ ] 大容量ファイルの処理
- [ ] パス長制限
- [ ] ファイル名の特殊文字

---

### 5. Testing

#### 5.1 ユニットテスト
- [x] config パッケージ ✅
  - [x] YAML読み込み (config/config_test.go - 96.4% coverage)
  - [x] 優先度解決ロジック
  - [x] エラーケース (無効なYAML、存在しないファイル等)
- [ ] syncer パッケージ
  - [ ] ファイル収集
  - [ ] 競合解決
  - [ ] ファイル配布
  - 注: pushのテストで間接的にカバー済み
- [x] utils パッケージ ✅
  - [x] ファイル操作 (utils/file_test.go - 72% coverage)
- [x] cmd パッケージ (部分的)
  - [x] pushコマンド (cmd/push_test.go - 8 test cases)
  - [x] initコマンド (cmd/init_test.go - 6 test cases)
  - [x] rmコマンド (cmd/rm_test.go)
  - [x] mvコマンド (cmd/mv_test.go)

#### 5.2 統合テスト
- [ ] push コマンドのエンドツーエンドテスト
- [ ] rm コマンドのエンドツーエンドテスト
- [ ] mv コマンドのエンドツーエンドテスト
- [ ] 複数グループの処理
- [ ] エラーシナリオ

#### 5.3 テストデータ
- [ ] サンプルプロジェクト構造の作成
- [ ] テスト用の設定ファイル
- [ ] 期待される出力の定義

---

### 6. Documentation

#### 6.1 コードドキュメント
- [ ] パッケージレベルのドキュメント
- [ ] 公開関数のGoDoc
- [ ] 複雑なロジックのコメント

#### 6.2 ユーザードキュメント
- [✅] README.md (基本完成)
- [ ] 詳細な使用例
- [ ] トラブルシューティングガイド
- [ ] FAQセクション

---

### 7. Future Enhancements (優先度低)

- [ ] 除外パターン設定（.gitignoreのような）
- [ ] バックアップ機能
- [ ] 変更履歴の記録
- [ ] ロールバック機能
- [ ] プレビューモードの改善
- [ ] カラー出力対応
- [ ] 進捗バーの表示
- [ ] 並列処理による高速化
- [ ] ファイル内容の差分表示
- [ ] WebUIの提供

---

## 実装優先順位

### Phase 1: Core Functionality (最優先) ✅ 完了
1. ✅ utils/file.go - 基本的なファイル操作
2. ✅ syncer/collector.go - ファイル収集
3. ✅ syncer/resolver.go - 競合解決
4. ✅ syncer/syncer.go - ファイル配布
5. ✅ push コマンドの完成

### Phase 2: Additional Commands
6. utils/prompt.go - 確認プロンプト
7. rm コマンドの完成
8. mv コマンドの完成

### Phase 3: Quality & Testing
9. エラーハンドリングの強化
10. ユニットテスト
11. 統合テスト
12. ドキュメント整備

---

## 進捗トラッキング

### 全体進捗: 75%

```
██████████████████████░░░░░░ 75%
```

### フェーズ別進捗

| フェーズ | 進捗 | 状態 | 完了タスク | 残りタスク |
|---------|------|------|-----------|-----------|
| **Phase 1: Core Functionality** | 100% | ✅ 完了 | 5/5 | なし |
| **Phase 2: Additional Commands** | 0% | 🚧 進行中 | 0/3 | utils/prompt.go, rm実装, mv実装 |
| **Phase 3: Quality & Testing** | 0% | ⏸️ 未着手 | 0/4 | エラーハンドリング, ユニットテスト, 統合テスト, ドキュメント |

### パッケージ別進捗

| パッケージ | ファイル数 | 完成 | 進捗 |
|-----------|-----------|------|------|
| main | 1/1 | ✅ | 100% |
| cmd | 5/6 | 🚧 | 83% |
| config | 1/1 | ✅ | 100% |
| syncer | 3/3 | ✅ | 100% |
| utils | 1/1 | ✅ | 100% |

### 重要マイルストーン

- [x] プロジェクト構造作成 (2025-11-14)
- [x] 設定ファイル管理実装 (2025-11-14)
- [x] initコマンド実装 (2025-11-14)
- [x] listコマンド実装 (2025-11-14)
- [x] pushコマンド実装 (2025-11-14)
- [ ] rm/mvコマンド実装（次のマイルストーン）
- [ ] v0.1.0リリース
- [ ] テスト完備
- [ ] v1.0.0リリース

### 最新の変更履歴

**2025-11-22**
- ✅ .github/workflows/version-update.yml追加（バージョン自動更新ワークフロー実装）
  - git tagのpushでトリガー（例: `v0.2.0`）
  - cmd/root.goのVersionフィールドを自動更新
  - バージョン更新PRを自動作成
  - ビルド検証を含む
  - workflow_dispatch手動トリガーサポート
  - セマンティックバージョニング検証
  - detached HEAD状態の修正（mainブランチ明示的チェックアウト）
- ✅ CLAUDE.md更新（Version Management Workflowセクション追加）
  - リリースプロセスのドキュメント化
  - 手動バージョン更新の手順追加

**2025-11-15**
- ✅ cmd/config.go追加（設定管理コマンド実装）
  - config show - 設定表示
  - config add-group - グループ追加
  - config remove-group - グループ削除
  - config add-project - プロジェクト追加
  - config remove-project - プロジェクト削除
  - config set-priority - 優先度設定
- ✅ config パッケージに変更機能追加
  - Save() - 設定ファイル保存
  - AddGroup() / RemoveGroup()
  - AddProject() / RemoveProject()
  - SetPriority()
- ✅ config_test.go拡張（新機能のテスト追加、88.8% coverage）
- ✅ config/config_test.go追加（設定ファイル読み込み・優先度解決のテスト、96.4% coverage）
- ✅ utils/file_test.go追加（ファイル操作の包括的テスト、72% coverage）
- ✅ cmd/push_test.go追加（pushコマンドの8テストケース）
- ✅ cmd/init_test.go追加（initコマンドの6テストケース）
- ✅ cmd/rm_test.go追加
- ✅ cmd/mv_test.go追加
- ✅ CopyFile関数のバグ修正（同一ファイルコピー時のデータ損失防止）

**2025-11-14**
- ✅ 設定ファイルの場所を`~/.config/claude-sync/config.yaml`に固定
- ✅ `claude-sync init`コマンド追加（インタラクティブ設定作成）
- ✅ `claude-sync list`コマンド完成（グループ一覧・詳細表示）
- ✅ README.md更新（使い方、アンインストール方法追加）
- ✅ spec/todo.md作成（実装TODOリスト）
- ✅ utils/file.go実装（ファイルコピー、削除、移動、ハッシュ計算）
- ✅ syncer/collector.go実装（ファイル収集ロジック）
- ✅ syncer/resolver.go実装（競合解決ロジック）
- ✅ syncer/syncer.go実装（ファイル配布ロジック）
- ✅ `claude-sync push`コマンド完成（Phase 1完了）

### 次の作業予定

1. **utils/prompt.go** - 確認プロンプト実装
2. **cmd/rm.go** - rmコマンド完成
3. **cmd/mv.go** - mvコマンド完成

**推定完了時期**: Phase 2完了まで 0.5-1日

---

**最終更新日**: 2025-11-14

aude-sync 仕様書

## 概要

複数プロジェクト間で`.claude`ディレクトリを同期するCLIツール。
グループ単位で管理し、ファイルの追加・上書き・削除・移動を一括で行う。

---

## コマンド

### `claude-sync push <group>`

指定グループ内の全プロジェクトから`.claude`ディレクトリのファイルを収集し、グループ内の全プロジェクトに配布する。

**動作:**
1. グループ内の全プロジェクトから`.claude`配下のファイルを収集
2. 同名ファイルが複数ある場合、`priority`設定に基づいて優先度の高いファイルを採用
3. 収集したファイル（重複は優先度解決済み）をグループ内の全プロジェクトに配布
4. 各プロジェクトに存在しないファイルは追加、既存ファイルは上書き
5. 配布対象外のファイル（各プロジェクト固有のファイル）は削除しない

**例:**
```bash
claude-sync push frontend
```

**出力例:**
```
Collecting files...
✓ web: 10 files (priority: 1)
✓ mobile: 8 files (priority: 2)
✓ admin: 5 files (priority: 3)

Resolving conflicts...
- config.json: using web (priority: 1)
- prompts/coding.md: using web (priority: 1)
- prompts/mobile-specific.md: no conflict

    Syncing...
    ✓ Synced to web (3 new files)
    ✓ Synced to mobile (5 new files, 2 overwritten)
✓ Synced to admin (8 new files, 1 overwritten)

    Summary: 15 unique files synced across 3 projects
    ```

    ---

### `claude-sync rm <group> <path>`

    指定グループ内の全プロジェクトから、指定したファイルまたはディレクトリを削除する。

    **動作:**
    1. グループ内の全プロジェクトから指定パスを検索
    2. 確認プロンプトを表示（`--force`で省略可能）
    3. 存在するプロジェクトから削除実行
    4. 存在しないプロジェクトはスキップ

    **例:**
    ```bash
    claude-sync rm frontend prompts/old-prompt.md
    claude-sync rm backend prompts/deprecated/  # ディレクトリごと削除
    ```

    **出力例:**
    ```
    This will delete from 'frontend' group:
    - ./packages/web/.claude/prompts/old-prompt.md
    - ./packages/mobile/.claude/prompts/old-prompt.md
    - ./packages/admin/.claude/prompts/old-prompt.md

    Continue? [y/N]: y

    ✓ Deleted from web
    ✓ Deleted from mobile
✗ Not found in admin (skipped)

    Summary: 2 files deleted
    ```

    **オプション:**
    - `--force`: 確認プロンプトをスキップ
    - `--dry-run`: 削除対象を表示するのみ（実際には削除しない）

    ---

### `claude-sync mv <group> <from> <to>`

    指定グループ内の全プロジェクトで、ファイルまたはディレクトリを移動・リネームする。

    **動作:**
    1. グループ内の全プロジェクトから移動元パスを検索
    2. 確認プロンプトを表示（`--force`で省略可能）
    3. 存在するプロジェクトで移動実行
    4. 移動先に同名ファイル/ディレクトリがある場合は警告して上書き確認
    5. 存在しないプロジェクトはスキップ

    **例:**
    ```bash
    claude-sync mv frontend prompts/old.md prompts/new.md
    claude-sync mv backend old-dir/ new-dir/
    ```

    **出力例:**
    ```
    This will rename in 'frontend' group:
    prompts/old.md → prompts/new.md

    Continue? [y/N]: y

    ✓ Moved in web
    ✓ Moved in mobile
✗ Source not found in admin (skipped)

    Summary: 2 files moved
    ```

    **オプション:**
    - `--force`: 確認プロンプトをスキップ
    - `--dry-run`: 移動対象を表示するのみ（実際には移動しない）

    ---

### `claude-sync list [group]`

    グループ一覧、または指定グループの詳細を表示する。

    **例:**
    ```bash
# 全グループ一覧
    claude-sync list

# 出力:
# Groups:
#   frontend (3 projects)
#   backend (3 projects)
#   infra (2 projects)

# 特定グループの詳細
    claude-sync list frontend

# 出力:
# Group: frontend
# Priority order:
#   1. web: ./packages/web/.claude
#   2. mobile: ./packages/mobile/.claude
#   3. admin: ./packages/admin/.claude (default priority)
    ```

    ---

## 設定ファイル

### ファイル名・配置場所

    `.claude-sync.yaml`

    **検索順序:**
    1. カレントディレクトリ
    2. 親ディレクトリを遡って検索
    3. `~/.config/claude-sync/config.yaml`（グローバル設定）

    ---

### 設定形式

    ```yaml
    groups:
    <group-name>:
    paths:
    <alias>: <path-to-.claude-dir>
    ...
    priority:
    - <alias-or-path>
    ...
    ```

    ---

### 設定例

#### 基本形（エイリアス付き）

    ```yaml
    groups:
frontend:
paths:
web: ./packages/web/.claude
mobile: ./packages/mobile/.claude
admin: ./packages/admin/.claude
priority:
- web      # 最優先
- mobile   # 次優先
# admin は指定なし = 最低優先度

backend:
paths:
api: ./services/api/.claude
worker: ./services/worker/.claude
batch: ./services/batch/.claude
priority:
- api
```

#### エイリアスなし（シンプル形）

```yaml
groups:
frontend:
paths:
- ./packages/web/.claude
- ./packages/mobile/.claude
- ./packages/admin/.claude
priority:
- ./packages/web/.claude
- ./packages/mobile/.claude
```

#### priority未指定（paths順序がデフォルト優先度）

```yaml
groups:
infra:
paths:
terraform: ./terraform/.claude
k8s: ./k8s/.claude
# priority未指定 = pathsの記載順が優先度
# 1. terraform (最優先)
# 2. k8s (次優先)
```

---

## 優先度ルール

### 1. priority指定ありの場合

`priority`リストに記載された順序が優先度となる。

- リストの上から順に優先度が高い
- `priority`に記載されていないプロジェクトは最低優先度（同率）

**例:**
```yaml
paths:
web: ./packages/web/.claude
mobile: ./packages/mobile/.claude
admin: ./packages/admin/.claude
priority:
- web
- mobile
```

**優先度:**
1. web（最優先）
2. mobile（次優先）
3. admin（最低優先度）

### 2. priority未指定の場合

`paths`の記載順がそのまま優先度となる。

**例:**
```yaml
paths:
- ./terraform/.claude
- ./k8s/.claude
```

**優先度:**
1. terraform（最優先）
2. k8s（次優先）

---

## 同期の詳細動作

### ファイル収集フェーズ

1. グループ内の全プロジェクトから`.claude/`配下のファイルをリストアップ
2. 各ファイルのパス（`.claude/`からの相対パス）をキーとして収集

### 競合解決フェーズ

同名ファイルが複数のプロジェクトに存在する場合:

1. 各ファイルの由来プロジェクトを確認
2. `priority`設定に基づき、優先度が最も高いプロジェクトのファイルを採用
3. 優先度が同じ場合は、`paths`の記載順で先に記載されている方を採用

### 配布フェーズ

1. 競合解決済みのファイルセットを、グループ内の全プロジェクトに配布
2. 各プロジェクトで:
- 存在しないファイル → 追加
- 既存ファイル → 上書き
- 配布対象外のファイル → 保持（削除しない）

---

## グローバルオプション

全コマンド共通で使用可能なオプション:

- `--config <path>`: 設定ファイルのパスを明示的に指定
- `--dry-run`: 実行内容をシミュレーション（実際には変更しない）
- `--verbose`: 詳細なログを出力
- `--force`: 確認プロンプトをスキップ（`rm`, `mv`コマンド）

**例:**
```bash
claude-sync push frontend --dry-run
claude-sync rm backend old.md --force
claude-sync push frontend --config ./custom-config.yaml
```

---

## エラーハンドリング

### 設定ファイルが見つからない

```
✗ Error: Configuration file not found
Searched locations:
- ./.claude-sync.yaml
- ../.claude-sync.yaml
- ~/.config/claude-sync/config.yaml

Create a configuration file or use --config flag
```

### グループが存在しない

```bash
claude-sync push nonexistent
```

```
✗ Error: Group 'nonexistent' not found in configuration
Available groups: frontend, backend, infra
```

### プロジェクトパスが無効

```
✗ Error: Invalid path in group 'frontend':
./packages/web/.claude does not exist
```

### 一部のプロジェクトで失敗

処理を続行し、最後にサマリーを表示:

    ```
✓ Synced to web (5 files)
    ✗ Failed to sync to mobile: permission denied
✓ Synced to admin (3 files)

    Summary: 2 succeeded, 1 failed
    ```

    ---

## 実装構成（参考）

    ```
    claude-sync/
    ├── main.go
    ├── cmd/
    │   ├── push.go      # pushコマンド
    │   ├── rm.go        # rmコマンド
    │   ├── mv.go        # mvコマンド
    │   └── list.go      # listコマンド
    ├── config/
    │   └── config.go    # YAML読み込み・検証
    ├── syncer/
    │   ├── collector.go # ファイル収集
    │   ├── resolver.go  # 競合解決
    │   └── syncer.go    # 配布ロジック
    └── utils/
    ├── file.go      # ファイル操作
    └── prompt.go    # 確認プロンプト
    ```

    ---

## ユースケース例

### ケース1: 新しいプロンプトをフロントエンド全体に配布

    ```bash
# webプロジェクトで新しいプロンプトを作成
    cd packages/web/.claude/prompts
    vim new-feature.md

# 配布
    claude-sync push frontend
    ```

### ケース2: 古いプロンプトを全プロジェクトから削除

    ```bash
    claude-sync rm backend prompts/deprecated/
    ```

### ケース3: ファイル名を統一

    ```bash
    claude-sync mv frontend old-name.md new-name.md
    ```

### ケース4: 優先度を考慮した設定の統一

    ```yaml
# webをマスターとして、他のプロジェクトに設定を配布
    frontend:
paths:
web: ./packages/web/.claude
mobile: ./packages/mobile/.claude
priority:
- web  # webの設定を優先
```

```bash
claude-sync push frontend
# → 全プロジェクトがwebの設定に統一される
```

---

## 注意事項

1. **バックアップ推奨**: 初回実行前に`.claude`ディレクトリのバックアップを推奨
2. **競合の理解**: 同名ファイルは優先度の高いプロジェクトの内容で上書きされる
3. **削除の不可逆性**: `rm`コマンドは取り消せないため、`--dry-run`での事前確認を推奨
4. **Git管理**: 設定ファイル（`.claude-sync.yaml`）はGit管理推奨、`.claude`ディレクトリ自体も必要に応じてGit管理

---


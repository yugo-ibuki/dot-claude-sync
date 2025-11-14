# CI セットアップ手順

GitHub App の権限制限により、`.github/workflows/` ファイルを自動でプッシュできませんでした。
以下の手順で手動セットアップをお願いします。

## セットアップ方法

### GitHub UI経由で作成（推奨）

1. GitHubでリポジトリを開く: https://github.com/yugo-ibuki/dot-claude-sync
2. "Actions" タブをクリック
3. "set up a workflow yourself" をクリック
4. ファイル名を `ci.yml` に設定
5. 以下の内容を貼り付け
6. "Commit changes" をクリック

### ローカルから手動プッシュ

```bash
cd /home/user/dot-claude-sync
git add .github/
git commit -m "add CI pipeline with GitHub Actions"
git push -u origin claude/setup-ci-pipeline-0175tezcMSk1hyWtp5qvvq1P
```

## ワークフローファイルの内容

`.github/workflows/ci.yml` に以下を作成してください：

```yaml
name: CI

on:
  push:
    branches: [ main, claude/* ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: ['1.23', '1.25']

    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: ${{ matrix.go-version }}

    - name: Verify dependencies
      run: |
        go mod verify
        go mod download

    - name: Build
      run: go build -v ./...

    - name: Run tests
      run: go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

    - name: Check formatting
      run: |
        if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
          echo "The following files are not formatted:"
          gofmt -s -l .
          exit 1
        fi

  lint:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4

    - uses: actions/setup-go@v5
      with:
        go-version: '1.25'

    - name: golangci-lint
      uses: golangci/golangci-lint-action@v4
      with:
        version: latest
```

## CI パイプラインの機能

- **マルチバージョンテスト**: Go 1.23 と 1.25 でビルド・テスト
- **コード品質チェック**: golangci-lint による静的解析
- **フォーマットチェック**: gofmt によるコードスタイル検証
- **依存関係検証**: go mod verify でモジュール整合性確認
- **レースコンディション検出**: go test -race で並行処理の問題を検出

## 注意事項

`.github/` ディレクトリは `.gitignore` に追加済みです。
ワークフローファイルはGitHub UI経由またはご自身のGit認証情報で直接プッシュしてください。

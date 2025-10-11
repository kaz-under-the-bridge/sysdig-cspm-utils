# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

Sysdig CSPMのコンプライアンス結果と違反リソースを管理・分析するためのGolang製CLIツール＆ライブラリ。Sysdig CSPM V1/V2 APIを使用して、コンプライアンス評価結果とインベントリリソースデータを取得・キャッシュし、SQLiteデータベースで管理する。

### 技術スタック
- **言語**: Go 1.23
- **データベース**: SQLite3 (CGO enabled)
- **外部依存**: github.com/kaz-under-the-bridge/sysdig-vuls-utils
- **ビルドツール**: Make, Task (go-task)
- **開発環境**: Dev Container対応

## 主要機能

1. **コンプライアンス結果取得**: CIS、SOC2、PCI-DSS等の各種コンプライアンス基準の評価結果を取得
2. **リソース詳細収集**: 違反が検出されたリソースの詳細情報を再帰的に収集
3. **マルチクラウド対応**: AWS、GCP、Azure、Kubernetesのリソースを統一的に管理
4. **データベース管理**: SQLiteによる構造化データストレージと分析機能
5. **レポート生成**: コンプライアンス違反の分析レポート自動生成

## 推奨実行フロー

### 標準的なコンプライアンスデータ取得

```bash
# scripts/collect-compliance.shを使用（推奨）
./scripts/collect-compliance.sh [targets] [options]

# 全て収集（AWS CIS + GCP CIS + SOC2）
./scripts/collect-compliance.sh all

# 特定の収集対象のみ
./scripts/collect-compliance.sh aws          # AWS CISのみ
./scripts/collect-compliance.sh gcp          # GCP CISのみ
./scripts/collect-compliance.sh soc2         # SOC2のみ
./scripts/collect-compliance.sh aws gcp      # AWSとGCPのみ

# ゾーンを指定
./scripts/collect-compliance.sh all --zone "Production"

# 出力ディレクトリを指定
./scripts/collect-compliance.sh aws --output data/custom-dir

# ヘルプ表示
./scripts/collect-compliance.sh --help
```

**スクリプトの自動処理:**
1. 環境変数の読み込み（`.devcontainer/.env`）
2. バイナリの自動ビルド（未ビルドの場合）
3. タイムスタンプ付きディレクトリの作成
4. 指定した収集対象のコンプライアンスデータ取得
5. 収集結果のサマリー表示

**収集対象:**
- **aws**: CIS Amazon Web Services Foundations Benchmark v3.0.0
- **gcp**: CIS Google Cloud Platform Foundation Benchmark v2.0.0
- **soc2**: SOC 2

### 生成されるファイル

```
data/YYYYMMDD_HHMMSS/
  ├── cis_aws.db   # AWS CIS Benchmark結果（コンプライアンス違反＋リソース詳細）
  ├── cis_gcp.db   # GCP CIS Benchmark結果（コンプライアンス違反＋リソース詳細）
  └── soc2.db      # SOC 2結果（コンプライアンス違反＋リソース詳細）

logs/
  ├── collect_aws_YYYYMMDD_HHMMSS.log   # AWS収集ログ
  ├── collect_gcp_YYYYMMDD_HHMMSS.log   # GCP収集ログ
  └── collect_soc2_YYYYMMDD_HHMMSS.log  # SOC2収集ログ
```

## ビルドとテストコマンド

### Makeを使用（シンプルな操作）

```bash
# ビルド
make build

# テスト実行
make test

# 統合テスト（要API認証）
make integration-test

# クリーンビルド
make clean build

# コードフォーマット
make fmt

# リント実行
make lint

# 依存関係管理
make deps

# ヘルプ表示
make help
```

### Taskを使用（詳細な制御）

```bash
# タスク一覧表示
task --list

# ビルド
task build

# テスト実行（自動的にgo vetを実行）
task test

# カバレッジ付きテスト
task test-coverage

# レース条件検出
task test-race

# コードフォーマット（gofmt + goimports）
task fmt

# リント実行
task lint

# 自動修正可能なリント問題を修正
task lint-fix

# すべての品質チェック実行（fmt, vet, staticcheck, lint, test）
task check

# コミット前チェック
task pre-commit

# 依存関係管理
task deps
task deps-update

# 統合テスト
task integration-test

# テストサーバーをビルド＆実行
task build-test-server
task run-test-server

# リリースビルド（全プラットフォーム）
task release

# 特定パッケージのテスト
task test-pkg PKG=pkg/client
```

**重要**: Goコードを編集した際は必ず以下を実行:
```bash
task fmt    # goimports + gofmt で自動整形
task vet    # go vet で静的解析
```

または統合コマンド:
```bash
task fix    # goimports + go fmt を一括実行
task check  # 全品質チェックを実行
```

## 主要APIエンドポイント

### Compliance Results API
- エンドポイント: `https://us2.app.sysdig.com/api/cspm/v1/compliance/requirements`
- 用途: コンプライアンス要件の評価結果を取得
- 詳細: [Compliance Results API](docs/sysdig-api-ref/Compliance%20Results.md)

### Inventory Resources API
- エンドポイント: `https://us2.app.sysdig.com/api/cspm/v1/inventory/resources`
- 用途: 違反リソースの詳細情報を取得
- 詳細: [Inventory Resources API](docs/sysdig-api-ref/Search%20and%20list%20Inventory%20Resources.md)

## 環境変数

```bash
# 必須
export SYSDIG_API_TOKEN="your-token-here"

# オプション
export SYSDIG_API_URL="https://us2.app.sysdig.com"  # デフォルト
export SYSDIG_API_TIMEOUT="60"                       # APIタイムアウト（秒）
export SYSDIG_CACHE_TTL="900"                        # キャッシュTTL（秒）
```

## ディレクトリ構造

```
.
├── .devcontainer/         # VS Code Dev Container設定
│   ├── devcontainer.json  # コンテナ設定
│   ├── Dockerfile         # コンテナイメージ
│   └── .env               # 環境変数（要作成、.gitignore対象）
├── .vscode/               # VS Code設定
├── cmd/
│   ├── csmp-utils/        # メインCLIアプリケーション
│   └── test-server/       # テスト用モックサーバー
├── pkg/
│   ├── cache/             # キャッシュ管理
│   ├── client/            # CSPM APIクライアント
│   ├── config/            # 設定管理
│   ├── database/          # SQLiteデータベース操作
│   ├── models/            # データモデル定義
│   ├── output/            # 出力フォーマット（テーブル等）
│   └── sysdig/            # Sysdig固有クライアント
├── internal/
│   └── testutil/          # テストユーティリティ
│       ├── fixtures.go    # テストフィクスチャ
│       ├── fixture_loader.go  # フィクスチャローダー
│       └── mock_server.go # モックサーバー実装
├── docs/
│   ├── design/            # 設計ドキュメント
│   └── sysdig-api-ref/    # Sysdig API リファレンス
├── scripts/               # ユーティリティスクリプト
├── data/                  # 出力データディレクトリ
├── logs/                  # ログ出力ディレクトリ
├── response_sample/       # APIレスポンスサンプル
├── Makefile               # Make設定
├── Taskfile.yml           # Task設定（推奨）
└── go.mod                 # Go依存関係定義
```

## 開発環境

### Dev Container

プロジェクトはDev Container対応しており、以下の機能を提供:
- Go 1.23開発環境
- 必要なツール自動インストール（golangci-lint, goimports, Task等）
- VS Code拡張機能自動セットアップ
- ホストのgit設定・Claude設定をマウント

**セットアップ:**
1. `.devcontainer/.env` ファイルを作成し、環境変数を設定
2. VS Codeで「Dev Containerで再度開く」を選択

### ローカル開発

```bash
# 依存関係インストール
task deps

# 開発サイクル
task fmt && task vet && task test

# ビルドして実行
task run

# コミット前チェック
task pre-commit
```

## 開発ガイドライン

### コード規約
- Go 1.23を使用（go.mod参照）
- CGO必須（SQLite3使用のため `CGO_ENABLED=1`）
- `goimports`と`gofmt`でコード整形
- `golangci-lint`でリントチェック
- エラーハンドリングは明示的に
- テストカバレッジ80%以上を維持

### コーディングフロー
1. コード編集
2. `task fmt` でフォーマット
3. `task vet` で静的解析
4. `task test` でテスト実行
5. `task lint` でリントチェック

または統合コマンド: `task check`

### API使用上の注意
- Rate Limitに注意（推奨: 3秒間隔）
- ページネーション処理を適切に実装
- 大量データは並列処理でバッチ取得
- タイムアウト設定を適切に設定

### データベース設計
- マルチクラウド対応スキーマ
- プラットフォーム固有フィールドはJSON保存
- 共通フィールドは個別カラムでインデックス化
- 詳細: [SQLiteスキーマ設計](docs/design/sqlite-schema-design.md)

### テスト戦略
- ユニットテスト: `pkg/` 配下の各パッケージ
- 統合テスト: `pkg/integration_test.go` （要APIトークン）
- モックサーバー: `internal/testutil/mock_server.go`
- フィクスチャ管理: `internal/testutil/fixtures.go`

## トラブルシューティング

### よくある問題と解決方法

1. **API Rate Limit エラー**
   - 解決: `API_DELAY`環境変数で遅延を増やす
   - 推奨値: 3-5秒

2. **メモリ不足エラー**
   - 解決: バッチサイズを小さくする
   - `BATCH_SIZE=10`を設定

3. **認証エラー**
   - 解決: `SYSDIG_API_TOKEN`が正しく設定されているか確認
   - トークンの有効期限を確認

4. **CGOエラー**
   - 解決: `CGO_ENABLED=1`環境変数を設定
   - SQLite3使用のため必須

5. **ビルドエラー (Taskfile.yml:6)**
   - 原因: `MAIN_PATH`のタイポ (`cmd/csmp-utils` → `cmd/cspm-utils`)
   - 解決: Taskfile.ymlのパス修正が必要な場合あり

## 参考ドキュメント

### API仕様
- [Sysdig CSPM API統合ガイド](docs/sysdig-api-ref/CSPM-API-Integration-Guide.md)
- [Compliance Results API](docs/sysdig-api-ref/Compliance%20Results.md)
- [Inventory Resources API](docs/sysdig-api-ref/Search%20and%20list%20Inventory%20Resources.md)

### 設計ドキュメント
- [SQLiteスキーマ設計](docs/design/sqlite-schema-design.md)
- [データ構造定義](docs/design/data-structures.md)
- [ページネーション設計](docs/design/pagination-design.md)
- [Compliance Results API設計](docs/design/compliance-results-api.md)
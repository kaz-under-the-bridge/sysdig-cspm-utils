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

## 最短実行フロー（推奨）

### ワンコマンドでデータ収集＋レポート生成

```bash
# 全て収集＋レポート生成（AWS CIS + GCP CIS + SOC2）
task workflow-all

# 個別に収集＋レポート生成
task workflow-aws     # AWS CISのみ
task workflow-gcp     # GCP CISのみ
task workflow-soc2    # SOC2のみ
```

**自動処理内容:**
1. 環境変数の読み込み（`.devcontainer/.env`）
2. バイナリの自動ビルド
3. タイムスタンプ付きディレクトリの作成
4. コンプライアンスデータ収集
5. **同一ディレクトリにデフォルト設定でレポート生成**（High重要度、詳細モード）
6. 結果サマリー表示

**収集対象:**
- **aws**: CIS Amazon Web Services Foundations Benchmark v3.0.0
- **gcp**: CIS Google Cloud Platform Foundation Benchmark v2.0.0
- **soc2**: SOC 2

### 生成されるファイル

```
data/YYYYMMDD_HHMMSS/
  ├── cis_aws.db      # AWS CIS Benchmark結果（データベース）
  ├── report_aws.md   # AWS CIS Benchmarkレポート（Markdown）
  ├── cis_gcp.db      # GCP CIS Benchmark結果（データベース）
  ├── report_gcp.md   # GCP CIS Benchmarkレポート（Markdown）
  ├── soc2.db         # SOC 2結果（データベース）
  └── report_soc2.md  # SOC 2レポート（Markdown）

logs/
  ├── collect_aws_YYYYMMDD_HHMMSS.log   # AWS収集ログ
  ├── collect_gcp_YYYYMMDD_HHMMSS.log   # GCP収集ログ
  └── collect_soc2_YYYYMMDD_HHMMSS.log  # SOC2収集ログ
```

### 既存データベースからのレポート再生成

既にデータベースがある場合、レポートのみ再生成できます：

```bash
# 既存の最新DBからレポート再生成（別タイムスタンプで生成）
task report-aws
task report-gcp
task report-soc2

# カスタム設定でレポート生成
task report-aws-custom OUTPUT=custom.md SEVERITY=all MODE=full
```

## 開発時の必須タスク

### コード編集後に必ず実行

```bash
# 最低限の品質チェック
task fix    # goimportsで自動整形（go fmtを含む）
task vet    # go vetで静的解析

# または統合コマンド
task check  # 全品質チェック実行（fmt, vet, staticcheck, lint, test）
```

### コミット前に必ず実行

```bash
task pre-commit  # fmt + lint + test-short + git diff確認
```

### 主要タスク一覧

#### ワークフロータスク（データ収集＋レポート生成）
```bash
task workflow-all    # 全て実行（AWS + GCP + SOC2）
task workflow-aws    # AWS CISのみ
task workflow-gcp    # GCP CISのみ
task workflow-soc2   # SOC2のみ
```

#### 開発・ビルドタスク
```bash
task --list          # 全タスク一覧表示
task build           # バイナリビルド
task clean           # ビルド成果物削除
task run             # ビルド＆実行
```

#### テストタスク
```bash
task test            # 全テスト実行（go vet前提）
task test-coverage   # カバレッジ付きテスト
task test-race       # レース条件検出テスト
task test-short      # 短時間テストのみ
task test-pkg PKG=pkg/client  # 特定パッケージのテスト
```

#### コード品質タスク
```bash
task fmt             # コードフォーマット（go fmt + gofmt）
task fix             # 自動整形（goimportsのみ、推奨）
task vet             # go vet実行
task lint            # リント実行
task lint-fix        # 自動修正可能なリント問題を修正
task staticcheck     # staticcheck実行
task check           # 全品質チェック（fmt + vet + staticcheck + lint + test）
```

#### 依存関係管理タスク
```bash
task deps            # 依存関係管理（download + tidy + verify）
task deps-update     # 依存関係更新
task tidy            # go mod tidy
```

#### レポート生成タスク
```bash
task report-aws      # 既存DBから再レポート生成
task report-gcp      # 既存DBから再レポート生成
task report-soc2     # 既存DBから再レポート生成
task report-aws-custom OUTPUT=file.md SEVERITY=all MODE=full  # カスタム設定
```

#### その他のタスク
```bash
task release         # リリースビルド（全プラットフォーム）
task ci              # CI環境タスク
task pre-commit      # コミット前チェック
task docker-build    # Dockerイメージビルド
task docker-run      # Dockerコンテナ実行
task integration-test # 統合テスト（要APIトークン）
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
│   ├── cspm-utils/        # メインCLIアプリケーション
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
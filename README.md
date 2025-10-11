# sysdig-cspm-utils

Sysdig CSPMのコンプライアンス結果と違反リソースを管理・分析するためのGolang製CLIツール＆ライブラリ。

## 概要

Sysdig CSPM V1/V2 APIを使用して、コンプライアンス評価結果とインベントリリソースデータを取得・キャッシュし、SQLiteデータベースで管理します。

### 主要機能

- **コンプライアンス結果取得**: CIS、SOC2、PCI-DSS等の各種コンプライアンス基準の評価結果を取得
- **リソース詳細収集**: 違反が検出されたリソースの詳細情報を再帰的に収集
- **マルチクラウド対応**: AWS、GCP、Azure、Kubernetesのリソースを統一的に管理
- **データベース管理**: SQLiteによる構造化データストレージと分析機能
- **レポート生成**: コンプライアンス違反の分析レポート自動生成

## クイックスタート

### 前提条件

- Go 1.23以上
- CGO有効化（SQLite3使用のため）
- Sysdig APIトークン

### インストール

```bash
# リポジトリをクローン
git clone https://github.com/kaz-under-the-bridge/sysdig-cspm-utils.git
cd sysdig-cspm-utils

# 依存関係をインストール
task deps

# ビルド
task build
```

### 環境変数の設定

```bash
export SYSDIG_API_TOKEN="your-token-here"
export SYSDIG_API_URL="https://us2.app.sysdig.com"  # オプション
```

### 基本的な使い方

#### スクリプトを使用（推奨）

```bash
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

スクリプトは以下を自動的に実行します：
- 環境変数の読み込み（`.devcontainer/.env`）
- バイナリの自動ビルド（未ビルドの場合）
- タイムスタンプ付きディレクトリの作成
- 指定した収集対象のデータ取得
- 収集結果のサマリー表示

#### 手動でコマンドを実行

```bash
# 環境変数を読み込む
source .devcontainer/.env

# ビルド
task build

# コンプライアンスデータ収集
./bin/cspm-utils -command collect \
  -policy "CIS Amazon Web Services Foundations Benchmark v3.0.0" \
  -zone "Entire Infrastructure" \
  -db data/cis_aws.db
```

## 出力ファイル

実行すると以下のファイルが生成されます：

```
data/YYYYMMDD_HHMMSS/
  ├── cis_aws.db   # AWS CIS Benchmark結果
  ├── cis_gcp.db   # GCP CIS Benchmark結果
  └── soc2.db      # SOC 2結果

logs/
  ├── collect_aws_YYYYMMDD_HHMMSS.log   # AWS収集ログ
  ├── collect_gcp_YYYYMMDD_HHMMSS.log   # GCP収集ログ
  └── collect_soc2_YYYYMMDD_HHMMSS.log  # SOC2収集ログ
```

## 開発

### ビルドコマンド

```bash
# Taskを使用
task build        # ビルド
task test         # テスト実行
task fmt          # コードフォーマット
task check        # 全品質チェック
task pre-commit   # コミット前チェック
```

### Dev Container

VS Code Dev Container対応。以下の手順で開発環境を構築：

1. `.devcontainer/.env`ファイルを作成し、環境変数を設定
2. VS Codeで「Dev Containerで再度開く」を選択

## ドキュメント

- [CLAUDE.md](CLAUDE.md) - 詳細な開発ガイド
- [SQLiteスキーマ設計](docs/design/sqlite-schema-design.md)
- [Sysdig CSPM API統合ガイド](docs/sysdig-api-ref/CSPM-API-Integration-Guide.md)
- [Compliance Results API](docs/sysdig-api-ref/Compliance%20Results.md)
- [Inventory Resources API](docs/sysdig-api-ref/Search%20and%20list%20Inventory%20Resources.md)

## ライセンス

MIT License

## 作者

kaz-under-the-bridge

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

#### ワンコマンドでデータ収集＋レポート生成（最短・推奨）

```bash
# 全て収集＋レポート生成（AWS CIS + GCP CIS + SOC2）
task workflow-all

# 個別に収集＋レポート生成
task workflow-aws     # AWS CISのみ
task workflow-gcp     # GCP CISのみ
task workflow-soc2    # SOC2のみ
```

**自動処理内容:**
- 環境変数の読み込み
- バイナリの自動ビルド
- データ収集
- **同一ディレクトリにレポート生成**（High重要度、詳細モード）
- 結果サマリー表示

## 出力ファイル

実行すると以下のファイルが生成されます：

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

## カスタムレポート生成

既存のデータベースから異なる設定でレポートを再生成できます。

### 既存DBからの再生成

```bash
# 既存の最新DBからレポート再生成（別タイムスタンプで生成）
task report-aws
task report-gcp
task report-soc2

# カスタム設定でレポート生成
task report-aws-custom OUTPUT=custom.md SEVERITY=all MODE=full
```

### Python環境での直接実行

Python環境で直接レポート生成スクリプトを実行することも可能です。

#### Python環境のセットアップ

```bash
# 仮想環境を作成
python3 -m venv venv

# 仮想環境を有効化
source venv/bin/activate

# 依存パッケージをインストール
pip install -r requirements.txt
```

#### レポート生成コマンド

```bash
# 基本（High重要度、詳細モード）
python3 scripts/generate_compliance_report.py data/soc2.db report_soc2.md

# 全ての重要度を含む
python3 scripts/generate_compliance_report.py data/soc2.db report.md --severity all

# フルレポート（トップ10 + 詳細 + 統計）
python3 scripts/generate_compliance_report.py data/soc2.db report.md --mode full

# ソート順を変更
python3 scripts/generate_compliance_report.py data/soc2.db report.md --sort-by name
```

#### レポートオプション

| オプション | 値 | デフォルト | 説明 |
|-----------|-----|-----------|------|
| `--severity` | `high`, `all` | `high` | 重要度フィルター |
| `--mode` | `detail`, `full` | `detail` | レポートモード |
| `--sort-by` | `violations`, `name`, `severity` | `violations` | ソート順 |

## 開発

### 主要タスク

```bash
# タスク一覧表示
task --list

# ワークフロー（データ収集＋レポート生成）
task workflow-all     # 全て実行
task workflow-aws     # AWS CISのみ
task workflow-gcp     # GCP CISのみ
task workflow-soc2    # SOC2のみ

# 開発・ビルド
task build            # ビルド
task test             # テスト実行
task fmt              # コードフォーマット
task fix              # 自動整形（推奨）
task check            # 全品質チェック
task pre-commit       # コミット前チェック

# レポート再生成
task report-aws       # 既存DBから再生成
task report-gcp       # 既存DBから再生成
task report-soc2      # 既存DBから再生成
```

### コード編集後の必須タスク

```bash
# 最低限の品質チェック
task fix    # goimportsで自動整形
task vet    # go vetで静的解析

# または統合コマンド
task check  # 全品質チェック（fmt, vet, staticcheck, lint, test）
```

### コミット前の必須タスク

```bash
task pre-commit  # fmt + lint + test-short + git diff確認
```

### Dev Container

VS Code Dev Container対応。以下の手順で開発環境を構築：

1. `.devcontainer/.env`ファイルを作成し、環境変数を設定
2. VS Codeで「Dev Containerで再度開く」を選択

詳細な開発ガイドは[CLAUDE.md](CLAUDE.md)を参照してください。

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

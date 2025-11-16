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

## レポート作成の詳細ガイド

### 実行方法の選択

現在のデータ状況に応じて、以下の2つのアプローチから選択できます：

#### 1. ワークフローコマンド（データ収集 + レポート生成）

既存データがない場合、または最新データを取得したい場合に使用：

```bash
# 推奨: 必要な基準のみ実行
task workflow-aws     # AWS CIS Benchmarkのみ（最短）
task workflow-gcp     # GCP CIS Benchmarkのみ
task workflow-soc2    # SOC 2のみ

# または全て一度に実行
task workflow-all     # AWS + GCP + SOC2（時間がかかる）
```

#### 2. 既存DBからのレポート再生成

既にデータベースがある場合、レポートのみ再生成：

```bash
# デフォルト設定で再生成
task report-aws       # 最新のAWS CIS DBから再生成
task report-gcp       # 最新のGCP CIS DBから再生成
task report-soc2      # 最新のSOC 2 DBから再生成

# カスタム設定で再生成
task report-aws-custom OUTPUT=custom.md SEVERITY=all MODE=full
```

### 処理フローの詳細

ワークフローコマンドを実行すると、以下の処理が自動的に行われます：

1. **バイナリビルド** - `bin/cspm-utils` をビルド（CGO有効）
2. **タイムスタンプディレクトリ作成** - `data/YYYYMMDD_HHMMSS/` を作成
3. **データ収集** - Sysdig CSPM APIからコンプライアンスデータを取得
   - 環境変数: `SYSDIG_API_TOKEN`（必須）
   - 出力: SQLiteデータベース（例: `cis_aws.db`）
   - ログ: `logs/collect_aws_YYYYMMDD_HHMMSS.log`
4. **レポート生成** - 収集したデータからMarkdownレポートを生成
   - デフォルト設定: High重要度のみ、詳細モード
   - 出力: 同一ディレクトリ内にMarkdownファイル（例: `report_aws.md`）
5. **結果サマリー表示** - 生成されたファイル一覧を表示

### レポート生成のカスタマイズ

デフォルト設定を変更したい場合：

```bash
# カスタム設定の例
task report-aws-custom \
  OUTPUT=data/custom_report.md \
  SEVERITY=all \
  MODE=full \
  SORT=severity
```

カスタマイズ可能なオプション：
- **OUTPUT**: 出力ファイルパス（デフォルト: `data/report_aws_custom.md`）
- **SEVERITY**: 重要度フィルター
  - `high`: High重要度のみ（デフォルト）
  - `all`: 全ての重要度を含む
- **MODE**: レポートモード
  - `detail`: 詳細レポートのみ（デフォルト）
  - `full`: トップ10 + 詳細 + 統計情報
- **SORT**: ソート順
  - `violations`: 違反数の多い順（デフォルト）
  - `name`: コントロール名順
  - `severity`: 重要度順

### 事前確認事項

レポート作成を実行する前に確認してください：

1. **環境変数の設定**
   ```bash
   # .devcontainer/.env または環境変数に設定
   export SYSDIG_API_TOKEN="your-token-here"
   ```

2. **Python 3の利用可能性**
   ```bash
   python3 --version  # Python 3.x が必要
   ```

3. **必要なツールの確認**
   ```bash
   task --version     # Task (go-task) が必要
   go version         # Go 1.23+ が必要
   ```

### トラブルシューティング

#### データベースが見つからない

```
エラー: AWS CIS Benchmarkデータが見つかりません。
```

**解決策**: 先にワークフローコマンドを実行してデータを収集してください。
```bash
task workflow-aws  # データ収集 + レポート生成
```

#### API認証エラー

```
Error: authentication failed
```

**解決策**: `SYSDIG_API_TOKEN` 環境変数が正しく設定されているか確認してください。
```bash
echo $SYSDIG_API_TOKEN  # トークンが設定されているか確認
```

#### Python依存パッケージエラー

```
ModuleNotFoundError: No module named 'xxx'
```

**解決策**: 必要な依存パッケージをインストールしてください。
```bash
pip install -r requirements.txt
```

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

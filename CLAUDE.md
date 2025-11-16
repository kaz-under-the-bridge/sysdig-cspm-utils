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

## クイックスタート

### 基本的な使い方

```bash
# データ収集＋レポート生成（ワンコマンド）
task workflow-aws     # AWS CIS Benchmarkのみ（推奨: まずこれから）
task workflow-gcp     # GCP CIS Benchmarkのみ
task workflow-soc2    # SOC 2のみ
task workflow-all     # 全て実行（AWS + GCP + SOC2）

# 既存DBからレポート再生成
task report-aws       # 最新のAWS CIS DBから再生成
task report-gcp       # 最新のGCP CIS DBから再生成
task report-soc2      # 最新のSOC 2 DBから再生成
```

**必須環境変数**: `SYSDIG_API_TOKEN`（`.devcontainer/.env`に設定）

**生成される出力**: `data/YYYYMMDD_HHMMSS/` ディレクトリにSQLite DBとMarkdownレポートが生成されます。

詳細な実行手順、オプション、トラブルシューティングは「**レポート作成の詳細手順（技術ガイド）**」セクションを参照してください。

## レポート作成の詳細手順（技術ガイド）

### 実行方法の概要

レポート作成には以下の2つのアプローチがあります：

#### アプローチ1: ワークフローコマンド（推奨）

**使用タイミング**: 既存データがない場合、または最新データを取得したい場合

```bash
# 個別実行（推奨）
task workflow-aws     # AWS CIS Benchmarkのみ
task workflow-gcp     # GCP CIS Benchmarkのみ
task workflow-soc2    # SOC 2のみ

# 全て一度に実行
task workflow-all     # AWS + GCP + SOC2
```

**実行される処理フロー**:

1. **依存関係チェック & ビルド**
   - `deps: [build]` により `task build` が自動実行される
   - `CGO_ENABLED=1 go build -o bin/cspm-utils cmd/cspm-utils/main.go`
   - SQLite3使用のため、CGO必須

2. **タイムスタンプディレクトリ生成**
   ```bash
   TIMESTAMP=$(date +%Y%m%d_%H%M%S)
   DATA_DIR="data/${TIMESTAMP}"
   ```
   - 全てのデータが同一ディレクトリに集約される
   - 例: `data/20251116_195830/`

3. **データ収集スクリプト実行**
   ```bash
   ./scripts/collect-compliance.sh [aws|gcp|soc2|all] --output "${DATA_DIR}"
   ```
   - Sysdig CSPM V1 APIからコンプライアンス結果を取得
   - SQLiteデータベースに保存（`cis_aws.db`, `cis_gcp.db`, `soc2.db`）
   - ログファイル生成: `logs/collect_[type]_${TIMESTAMP}.log`
   - 環境変数: `SYSDIG_API_TOKEN`（必須）

4. **レポート生成Pythonスクリプト実行**
   ```bash
   python3 scripts/generate_compliance_report.py \
     "${DATA_DIR}/cis_aws.db" \
     "${DATA_DIR}/report_aws.md"
   ```
   - デフォルト設定: `--severity high --mode detail --sort-by violations`
   - 出力: Markdownフォーマットのレポート
   - データ収集と同一ディレクトリに出力

5. **結果サマリー表示**
   ```bash
   ls -lh "${DATA_DIR}/" | grep -v "^total"
   ```
   - 生成されたファイルの一覧とサイズを表示

#### アプローチ2: 既存DBからのレポート再生成

**使用タイミング**: 既にデータベースがあり、異なる設定でレポートを生成したい場合

```bash
# デフォルト設定（High重要度、詳細モード）
task report-aws
task report-gcp
task report-soc2

# カスタム設定
task report-aws-custom OUTPUT=file.md SEVERITY=all MODE=full SORT=severity
```

**内部処理**:

```bash
# 最新のDBファイルを自動検索
LATEST_DB=$(ls -t data/*/cis_aws.db 2>/dev/null | head -n 1)

# レポート生成（別タイムスタンプで生成）
OUTPUT_FILE="data/report_aws_$(date +%Y%m%d_%H%M%S).md"
python3 scripts/generate_compliance_report.py \
  "$LATEST_DB" \
  "$OUTPUT_FILE" \
  --severity high \
  --mode detail \
  --sort-by violations
```

### 生成されるファイル構造

```
プロジェクトルート/
├── data/
│   └── 20251116_195830/          # タイムスタンプディレクトリ
│       ├── cis_aws.db            # AWS CIS Benchmark SQLite DB
│       ├── report_aws.md         # AWS CISレポート（Markdown）
│       ├── cis_gcp.db            # GCP CIS Benchmark SQLite DB
│       ├── report_gcp.md         # GCP CISレポート（Markdown）
│       ├── soc2.db               # SOC 2 SQLite DB
│       └── report_soc2.md        # SOC 2レポート（Markdown）
├── logs/
│   ├── collect_aws_20251116_195830.log    # AWS収集ログ
│   ├── collect_gcp_20251116_195830.log    # GCP収集ログ
│   └── collect_soc2_20251116_195830.log   # SOC2収集ログ
└── bin/
    └── cspm-utils                # ビルドされたバイナリ
```

### レポート生成オプションの詳細

**Pythonスクリプトの引数**:

```bash
python3 scripts/generate_compliance_report.py \
  <db_path> <output_path> \
  [--severity {high,all}] \
  [--mode {detail,full}] \
  [--sort-by {violations,name,severity}]
```

**パラメータ説明**:

| パラメータ | 種別 | 必須 | デフォルト | 説明 |
|----------|------|------|-----------|------|
| `db_path` | 位置引数 | ✅ | - | SQLiteデータベースファイルのパス |
| `output_path` | 位置引数 | ✅ | - | 出力Markdownファイルのパス |
| `--severity` | オプション | ❌ | `high` | 重要度フィルター: `high`（高のみ）/ `all`（全て） |
| `--mode` | オプション | ❌ | `detail` | レポートモード: `detail`（詳細のみ）/ `full`（統計+詳細） |
| `--sort-by` | オプション | ❌ | `violations` | ソート順: `violations`（違反数）/ `name`（名前）/ `severity`（重要度） |

**重要**: `db_path` と `output_path` は位置引数なので、`--db` や `--output` オプションは使用できません。

**レポートモードの違い**:

- **`detail`モード**: 各コントロールの違反詳細をリスト形式で表示
- **`full`モード**: 以下を含む完全なレポート
  - トップ10違反コントロール（棒グラフ風表示）
  - 詳細レポート
  - 統計情報（総違反数、総リソース数、コントロール数）

### 事前確認チェックリスト

#### 1. 環境変数の設定確認

```bash
# 必須
echo $SYSDIG_API_TOKEN
# 出力: ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

# オプション
echo $SYSDIG_API_URL
# デフォルト: https://us2.app.sysdig.com
```

**設定場所**:
- Dev Container: `.devcontainer/.env`
- ローカル環境: `~/.bashrc`, `~/.zshrc`, または実行時の環境変数

#### 2. 必要なツールの確認

```bash
# Task (go-task)
task --version
# 出力例: Task version: v3.x.x

# Go 1.23+
go version
# 出力例: go version go1.23.x darwin/arm64

# Python 3.x
python3 --version
# 出力例: Python 3.11.x

# CGO有効化確認
go env CGO_ENABLED
# 出力: 1（有効）
```

#### 3. ディレクトリ構造の確認

```bash
# dataディレクトリが存在するか
ls -la data/
# .gitkeepが存在すればOK

# logsディレクトリが存在するか
ls -la logs/
# 存在しない場合は自動作成される
```

### トラブルシューティング

#### エラー1: データベースが見つからない

```
エラー: AWS CIS Benchmarkデータが見つかりません。先に 'task workflow-aws' を実行してください。
```

**原因**: `data/` ディレクトリ内に対象のDBファイルが存在しない

**解決策**:
```bash
# 最初にワークフローコマンドを実行してデータを収集
task workflow-aws
```

**確認コマンド**:
```bash
# 既存のDBファイルを検索
find data/ -name "*.db" -type f
```

#### エラー2: API認証失敗

```
Error: authentication failed (401 Unauthorized)
```

**原因**: `SYSDIG_API_TOKEN` が未設定または無効

**解決策**:
```bash
# トークンの設定を確認
echo $SYSDIG_API_TOKEN

# 未設定の場合は設定
export SYSDIG_API_TOKEN="your-valid-token-here"

# Dev Containerの場合は .devcontainer/.env に追記
echo 'SYSDIG_API_TOKEN=your-valid-token-here' >> .devcontainer/.env
```

**トークン取得方法**:
1. Sysdig UI にログイン
2. Settings → User Profile → API Token
3. トークンをコピー

#### エラー3: CGOエラー

```
# runtime/cgo
cgo: C compiler "gcc" not found
```

**原因**: CGOが無効、またはCコンパイラが未インストール

**解決策**:
```bash
# CGOを有効化
export CGO_ENABLED=1

# macOS: Xcode Command Line Toolsをインストール
xcode-select --install

# Linux: gccをインストール
apt-get install build-essential  # Debian/Ubuntu
yum install gcc                   # RHEL/CentOS
```

#### エラー4: Python依存パッケージエラー

```
ModuleNotFoundError: No module named 'xxx'
```

**解決策**:
```bash
# 依存パッケージをインストール
pip install -r requirements.txt

# または個別にインストール
pip install sqlite3  # 通常は標準ライブラリに含まれる
```

#### エラー5: Rate Limitエラー

```
Error: Rate limit exceeded (429 Too Many Requests)
```

**解決策**:
```bash
# API遅延を設定（秒単位）
export API_DELAY=5

# または collect-compliance.sh を編集してsleepを追加
```

### 実装の重要ポイント

#### Taskfile.ymlの設計

**workflow-awsタスクの仕組み**:
- `deps: [build]` により、実行前に自動的にバイナリがビルドされる
- データ収集とレポート生成が同一ディレクトリ（`data/${TIMESTAMP}/`）に出力
- DBファイルの存在確認を行い、エラーハンドリングを実装

**report-awsタスクの仕組み**:
- `ls -t data/*/cis_aws.db | head -n 1` で最新のDBファイルを自動検索
- レポートは新しいタイムスタンプで生成（既存ファイルを上書きしない）
- DBが見つからない場合はエラーメッセージを表示

詳細なコードは `Taskfile.yml` の385-482行目を参照してください。

### 高度な使用例

#### 複数の設定で並行レポート生成

```bash
# 並行実行（バックグラウンド）
task report-aws-custom OUTPUT=data/report_high.md SEVERITY=high MODE=detail &
task report-aws-custom OUTPUT=data/report_all.md SEVERITY=all MODE=full &
wait
echo "全てのレポート生成完了"
```

#### 特定のDBファイルからレポート生成

```bash
# 任意のDBファイルを指定
python3 scripts/generate_compliance_report.py \
  data/20251115_120000/cis_aws.db \
  data/report_20251115.md \
  --severity all \
  --mode full
```

#### デバッグモードでの実行

```bash
# Bashデバッグモードで実行
bash -x scripts/collect-compliance.sh aws --output data/test

# Pythonデバッグモード
python3 -u scripts/generate_compliance_report.py data/test/cis_aws.db test_report.md
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

## リスク受け入れ（Risk Acceptance）の扱い

### 概要

Sysdig CSPMのリスク受け入れ機能により、コンプライアンス違反を意図的に受け入れることができます。このプロジェクトでは以下の3つの機能を提供しています：

```bash
# リスク受け入れデータの収集
./bin/cspm-utils -command risk-collect -db data/risk_acceptances.db

# リスク受け入れ一覧表示
./bin/cspm-utils -command risk-list -db data/risk_acceptances.db
./bin/cspm-utils -command risk-list -db data/risk_acceptances.db -control-id "16027"

# リスク受け入れの削除（APIとDBの両方から削除）
./bin/cspm-utils -command risk-delete -db data/risk_acceptances.db -acceptance-id "abc123..."
```

### リスク受け入れの分類

リスク受け入れは **受け入れ理由 (reason)** によって以下のように分類されます：

#### 1. Sysdig Accepted Risk（Sysdigシステム受け入れ）

**特徴:**
- **Username**: "Sysdig"（システムによる自動生成）
- **Acceptance Date**: 0（システム生成時から存在）
- **目的**: Sysdigプラットフォーム自体の正常な動作に必要なリスクの受け入れ

**主な対象リソース:**
- **Kubernetesシステムリソース**: `kube-system`, `kube-public`, `kube-node-lease` namespace内のワークロード
- **Sysdig自社製品コンポーネント**: `name contains "sysdig"`パターンのリソース
- **クラウドプロバイダー管理リソース**: EKSデフォルトユーザー（`eks:`プレフィックス）など

**具体例:**

```plaintext
Control 36: Kubernetesデフォルトワークロード
  Filter: namespace in ("kube-node-lease", "kube-public", "kube-system")
          and kind in ("DaemonSet", "Deployment", "Service", "Job", "CronJob")
  理由: これらはKubernetesプラットフォーム自体またはクラウドプロバイダーによって
        管理されており、クラスタ運用にとって重要。安定性と互換性を優先するため、
        厳格なセキュリティベンチマークの対象外とする。

Control 2012: Sysdigコンポーネントの特権機能
  Filter: name contains "sysdig"
  理由: 重要なランタイムセキュリティ情報を取得するために、
        Sysdigコンポーネントには特別な機能（Capabilities）が必要。
```

#### 2. Risk Owned（ユーザー組織による受け入れ）

**特徴:**
- **Username**: 実際のユーザー名（例: `john.doe@example.com`）
- **Acceptance Date**: 実際の受け入れ日時
- **目的**: 組織のビジネス要件により意図的に受け入れるリスク

**主な対象:**
- 特定のIAMユーザー（外部サービス連携用など）
- S3 MFA Delete無効化（運用上の理由）
- 特定のAWSリージョン（未使用リージョン）

#### 3. Risk Transferred / Custom（その他）

- **Risk Transferred**: リスクを第三者に移転
- **Custom**: カスタム理由

### 分析・レポート生成時の扱い

**重要**: リスク受け入れデータを分析・レポート生成する際は、以下のルールに従うこと：

#### ✅ 除外対象（Sysdig Accepted Risk）

```sql
-- 分析から除外するリスク受け入れ
SELECT * FROM risk_acceptances
WHERE reason = 'Sysdig Accepted Risk'
```

**除外理由:**
- Sysdigプラットフォームの正常動作に必要
- ユーザー組織の判断や責任によるものではない
- システム自動生成のため、レビュー対象外

#### ✅ 分析・表示対象

```sql
-- 分析対象とするリスク受け入れ
SELECT * FROM risk_acceptances
WHERE reason IN ('Risk Owned', 'Risk Transferred', 'Custom')
```

**これらは:**
- ユーザー組織が意図的に受け入れたリスク
- 定期的なレビューが必要
- コンプライアンス監査の対象

### データ統計（参考）

実際の運用環境での統計例:

```
総数: 691件
├─ Sysdig Accepted Risk: 382件 (55.3%) ← 分析除外
├─ Risk Owned: 303件 (43.8%)           ← 分析対象
├─ Risk Transferred: 4件 (0.6%)        ← 分析対象
└─ Custom: 2件 (0.3%)                   ← 分析対象
```

### フィルタータイプ別分類

リスク受け入れの **filter** フィールドにより、受け入れ範囲を分類できます：

1. **全リソース対象** (filter = "" または NULL)
   - そのコントロールの全リソースが受け入れ対象
   - 例: 特定のコントロールを組織全体で無効化

2. **個別リソース指定** (name in ("リソース名"))
   - 特定の名前付きリソースのみ
   - 例: `name in ("analytics-service-user")`

3. **パターンマッチ** (name contains "パターン")
   - 名前パターンに一致するリソース
   - 例: `name contains "sysdig"`

4. **Kubernetes関連** (namespace in (...) and kind in (...))
   - Kubernetes特有の条件指定
   - 例: `namespace in ("kube-system") and kind in ("DaemonSet")`

### API仕様

- **検索API**: `POST /api/cspm/v1/compliance/violations/acceptances/search`
- **削除API**: `POST /api/cspm/v1/compliance/violations/revoke`
- 詳細: [Risk Acceptance API](docs/sysdig-api-ref/cspm-risk-acceptance.md)

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
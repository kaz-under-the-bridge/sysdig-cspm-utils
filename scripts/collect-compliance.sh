#!/usr/bin/env bash
#
# collect-compliance.sh - Sysdig CSPMコンプライアンスデータ収集スクリプト
#
# 使用方法:
#   ./scripts/collect-compliance.sh [aws|gcp|soc2|all] [options]
#
# 例:
#   ./scripts/collect-compliance.sh all                    # 全て収集
#   ./scripts/collect-compliance.sh aws                    # AWS CISのみ
#   ./scripts/collect-compliance.sh aws gcp                # AWSとGCP
#   ./scripts/collect-compliance.sh --zone "Production"    # ゾーン指定
#

set -euo pipefail

# スクリプトのディレクトリを取得
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${PROJECT_ROOT}"

# カラー出力
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# デフォルト設定
BINARY_PATH="bin/cspm-utils"
ZONE_NAME="Entire Infrastructure"
DATA_DIR=""
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# 収集対象
COLLECT_AWS=0
COLLECT_GCP=0
COLLECT_SOC2=0

# ポリシー定義
POLICY_CIS_AWS="CIS Amazon Web Services Foundations Benchmark v3.0.0"
POLICY_CIS_GCP="CIS Google Cloud Platform Foundation Benchmark v2.0.0"
POLICY_SOC2="SOC 2"

# 関数: ヘルプメッセージ
function show_help() {
    cat <<EOF
使用方法: $(basename "$0") [targets] [options]

収集対象:
  aws         AWS CIS Benchmark
  gcp         GCP CIS Benchmark
  soc2        SOC 2
  all         全て収集（デフォルト）

オプション:
  --zone ZONE      ゾーン名を指定（デフォルト: Entire Infrastructure）
  --output DIR     出力ディレクトリを指定（デフォルト: data/YYYYMMDD_HHMMSS）
  -h, --help       このヘルプを表示

例:
  $(basename "$0") all
  $(basename "$0") aws gcp
  $(basename "$0") soc2 --zone "Production"
  $(basename "$0") aws --output data/custom-dir

環境変数:
  SYSDIG_API_TOKEN    Sysdig APIトークン（必須）
  SYSDIG_API_URL      Sysdig API URL（オプション、デフォルト: https://us2.app.sysdig.com）

EOF
}

# 関数: エラーメッセージ
function error() {
    echo -e "${RED}ERROR: $*${NC}" >&2
    exit 1
}

# 関数: 警告メッセージ
function warn() {
    echo -e "${YELLOW}WARNING: $*${NC}" >&2
}

# 関数: 情報メッセージ
function info() {
    echo -e "${BLUE}INFO: $*${NC}"
}

# 関数: 成功メッセージ
function success() {
    echo -e "${GREEN}SUCCESS: $*${NC}"
}

# 関数: 環境変数の読み込み
function load_env() {
    local env_file="${PROJECT_ROOT}/.devcontainer/.env"

    if [[ -f "${env_file}" ]]; then
        info "環境変数ファイルを読み込みます: ${env_file}"
        set -a
        # shellcheck disable=SC1090
        source "${env_file}"
        set +a
    fi
}

# 関数: 環境チェック
function check_environment() {
    # APIトークンの確認
    if [[ -z "${SYSDIG_API_TOKEN:-}" ]]; then
        error "SYSDIG_API_TOKEN環境変数が設定されていません"
    fi

    # API URLのデフォルト設定
    export SYSDIG_API_URL="${SYSDIG_API_URL:-https://us2.app.sysdig.com}"

    info "API URL: ${SYSDIG_API_URL}"
    info "API Token: ${SYSDIG_API_TOKEN:0:10}..."
}

# 関数: バイナリのビルド
function build_binary() {
    if [[ ! -f "${BINARY_PATH}" ]]; then
        info "バイナリが見つかりません。ビルドを実行します..."
        task build || error "ビルドに失敗しました"
    else
        info "既存のバイナリを使用します: ${BINARY_PATH}"
    fi
}

# 関数: ディレクトリの準備
function prepare_directories() {
    # データディレクトリの設定
    if [[ -z "${DATA_DIR}" ]]; then
        DATA_DIR="data/${TIMESTAMP}"
    fi

    mkdir -p "${DATA_DIR}"
    mkdir -p logs

    info "データディレクトリ: ${DATA_DIR}"
    info "ログディレクトリ: logs"
}

# 関数: コンプライアンスデータの収集
function collect_data() {
    local policy_name="$1"
    local db_file="$2"
    local log_file="$3"

    info "収集開始: ${policy_name}"
    echo "----------------------------------------"

    local db_path="${DATA_DIR}/${db_file}"
    local log_path="logs/${log_file}"

    if "./${BINARY_PATH}" -command collect \
        -policy "${policy_name}" \
        -zone "${ZONE_NAME}" \
        -db "${db_path}" 2>&1 | tee "${log_path}"; then
        success "収集完了: ${db_file}"

        # ファイルサイズを表示
        if [[ -f "${db_path}" ]]; then
            local size=$(du -h "${db_path}" | cut -f1)
            info "データベースサイズ: ${size}"
        fi
    else
        warn "収集に失敗しました: ${policy_name}"
        return 1
    fi

    echo ""
}

# 関数: サマリーの表示
function show_summary() {
    echo ""
    echo "========================================"
    echo "収集結果サマリー"
    echo "========================================"
    echo "出力ディレクトリ: ${DATA_DIR}"
    echo ""

    if [[ -d "${DATA_DIR}" ]]; then
        echo "生成されたファイル:"
        ls -lh "${DATA_DIR}" | grep -v "^total" | awk '{printf "  %-40s %10s\n", $9, $5}'
        echo ""

        echo "合計サイズ:"
        du -sh "${DATA_DIR}" | awk '{printf "  %s\n", $1}'
    fi

    echo ""
    echo "ログファイル:"
    ls -1 logs/*_${TIMESTAMP}.log 2>/dev/null | sed 's/^/  /' || echo "  (なし)"
    echo ""
}

# 引数解析
if [[ $# -eq 0 ]]; then
    show_help
    exit 0
fi

while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)
            show_help
            exit 0
            ;;
        --zone)
            ZONE_NAME="$2"
            shift 2
            ;;
        --output)
            DATA_DIR="$2"
            shift 2
            ;;
        aws)
            COLLECT_AWS=1
            shift
            ;;
        gcp)
            COLLECT_GCP=1
            shift
            ;;
        soc2)
            COLLECT_SOC2=1
            shift
            ;;
        all)
            COLLECT_AWS=1
            COLLECT_GCP=1
            COLLECT_SOC2=1
            shift
            ;;
        *)
            error "不明な引数: $1 (ヘルプ: $(basename "$0") --help)"
            ;;
    esac
done

# 収集対象が指定されていない場合はallとみなす
if [[ $COLLECT_AWS -eq 0 && $COLLECT_GCP -eq 0 && $COLLECT_SOC2 -eq 0 ]]; then
    warn "収集対象が指定されていません。全て収集します。"
    COLLECT_AWS=1
    COLLECT_GCP=1
    COLLECT_SOC2=1
fi

# メイン処理開始
echo ""
info "Sysdig CSPMコンプライアンスデータ収集を開始します"
echo ""

# 環境変数の読み込み
load_env

# 環境チェック
check_environment

# バイナリのビルド
build_binary

# ディレクトリの準備
prepare_directories

# 収集実行
FAILED_COUNT=0

if [[ $COLLECT_AWS -eq 1 ]]; then
    collect_data "${POLICY_CIS_AWS}" "cis_aws.db" "collect_aws_${TIMESTAMP}.log" || ((FAILED_COUNT++))
fi

if [[ $COLLECT_GCP -eq 1 ]]; then
    collect_data "${POLICY_CIS_GCP}" "cis_gcp.db" "collect_gcp_${TIMESTAMP}.log" || ((FAILED_COUNT++))
fi

if [[ $COLLECT_SOC2 -eq 1 ]]; then
    collect_data "${POLICY_SOC2}" "soc2.db" "collect_soc2_${TIMESTAMP}.log" || ((FAILED_COUNT++))
fi

# サマリー表示
show_summary

# 終了
if [[ $FAILED_COUNT -eq 0 ]]; then
    success "全ての収集が完了しました"
    exit 0
else
    warn "一部の収集が失敗しました（失敗数: ${FAILED_COUNT}）"
    exit 1
fi

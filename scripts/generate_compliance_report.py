#!/usr/bin/env python3
"""
コンプライアンス違反レポート生成スクリプト

データベースからコンプライアンス違反とコントロール詳細を取得し、
日本語のMarkdownレポートを生成します。

対応ポリシー: SOC 2, CIS AWS, CIS GCP, PCI-DSS, HIPAA等
"""

import sqlite3
import sys
import argparse
import os
import re
from datetime import datetime

# DeepL翻訳の設定（オプショナル）
DEEPL_TRANSLATOR = None
try:
    import deepl
    DEEPL_API_KEY = os.environ.get('DEEPL_API_KEY')
    if DEEPL_API_KEY:
        DEEPL_TRANSLATOR = deepl.Translator(DEEPL_API_KEY)
        print(f"✓ DeepL翻訳が有効化されました")
    else:
        print("ℹ️  DEEPL_API_KEY環境変数が設定されていません。英語のまま出力します。")
except ImportError:
    print("ℹ️  deeplパッケージがインストールされていません。英語のまま出力します。")
    print("   翻訳を有効にするには: pip install deepl")

def translate_description(text):
    """
    英語テキストを日本語に翻訳（DeepL APIが設定されている場合）

    環境変数 DEEPL_API_KEY が設定されていれば翻訳、なければ英語のまま返す
    """
    if not text or not DEEPL_TRANSLATOR:
        return text

    try:
        result = DEEPL_TRANSLATOR.translate_text(text, target_lang="JA")
        return result.text
    except Exception as e:
        print(f"⚠️  翻訳エラー: {e}")
        return text  # エラー時は元のテキストを返す

def make_anchor_id(text):
    """
    テキストからMarkdownアンカーID を生成
    GitHubスタイル: 小文字化、スペース→ハイフン、特殊文字削除
    """
    # 小文字化
    text = text.lower()
    # 英数字と一部の記号以外を削除（日本語は保持）
    text = re.sub(r'[^\w\s\-]', '', text)
    # スペースをハイフンに
    text = re.sub(r'\s+', '-', text)
    return text

def generate_report(db_path, output_path, severity_filter='high', report_mode='detail', sort_by='violations'):
    """レポートを生成"""
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()

    # データベースからポリシー情報を取得
    cursor.execute("""
        SELECT DISTINCT policy_name, policy_type, platform
        FROM compliance_requirements
        LIMIT 1
    """)
    policy_info = cursor.fetchone()

    if policy_info:
        policy_name, policy_type, platform = policy_info
        # タイトル生成
        if policy_type:
            report_title = f"{policy_type}"
        else:
            report_title = "コンプライアンス"

        if platform and platform != "Multi-Cloud":
            report_title += f" ({platform})"

        report_title += " 違反レポート"
    else:
        report_title = "コンプライアンス違反レポート"
        policy_name = "N/A"

    # Severity filter設定
    severity_where = ""
    severity_label = "全て"
    if severity_filter == 'high':
        severity_where = " AND severity = 'High'"
        severity_label = "High"

    # レポート開始
    report = []
    report.append(f"# {report_title}\n")
    report.append(f"**生成日時**: {datetime.now().strftime('%Y年%m月%d日 %H:%M:%S')}\n")
    if policy_info:
        report.append(f"**ポリシー**: {policy_name}\n")
    report.append(f"**データベース**: `{db_path}`\n")
    report.append(f"**重要度フィルター**: {severity_label}\n\n")

    # 目次
    report.append("## 📑 目次\n\n")
    report.append("- [📊 サマリー](#-サマリー)\n")
    report.append("- [🎯 違反コントロールランキング](#-違反コントロールランキング)\n")
    if report_mode == 'full':
        report.append("- [🔴 トップ10違反要件](#-トップ10違反要件)\n")
    report.append("- [📋 詳細レポート（要件別）](#-詳細レポート要件別)\n")
    if report_mode == 'full':
        report.append("- [📦 影響を受けるリソース統計](#-影響を受けるリソース統計)\n")
        report.append("- [🎯 最も違反の多いコントロール（全体）](#-最も違反の多いコントロール全体)\n")
    report.append("\n")

    # サマリー統計
    cursor.execute("SELECT COUNT(*) FROM compliance_requirements")
    total_requirements = cursor.fetchone()[0]

    # コントロール統計（全体とフィルタ適用後）
    cursor.execute("SELECT COUNT(*) FROM controls")
    total_controls_all = cursor.fetchone()[0]

    cursor.execute(f"SELECT COUNT(*) FROM controls WHERE 1=1 {severity_where}")
    total_controls_filtered = cursor.fetchone()[0]

    cursor.execute(f"SELECT COUNT(*) FROM controls WHERE pass = 0 {severity_where}")
    failed_controls = cursor.fetchone()[0]

    cursor.execute("SELECT COUNT(*) FROM cloud_resources")
    total_resources = cursor.fetchone()[0]

    cursor.execute("SELECT COUNT(*) FROM control_resource_relations")
    total_relations = cursor.fetchone()[0]

    # リソースのpass/failed統計
    cursor.execute("""
        SELECT
            COUNT(CASE WHEN acceptance_status = 'failed' THEN 1 END) as failed,
            COUNT(CASE WHEN acceptance_status = 'passed' THEN 1 END) as passed,
            COUNT(CASE WHEN acceptance_status = 'accepted' THEN 1 END) as accepted
        FROM control_resource_relations
    """)
    res_stats = cursor.fetchone()
    failed_resources, passed_resources, accepted_resources = res_stats

    # 割合計算
    control_violation_rate = (failed_controls / total_controls_filtered * 100) if total_controls_filtered > 0 else 0
    resource_failed_rate = (failed_resources / total_relations * 100) if total_relations > 0 else 0
    resource_passed_rate = (passed_resources / total_relations * 100) if total_relations > 0 else 0

    report.append("## 📊 サマリー\n\n")
    report.append(f"- **コンプライアンス要件**: {total_requirements}件\n")
    report.append(f"- **違反コントロール**: {failed_controls}件 / {total_controls_filtered}件 ({control_violation_rate:.1f}%)\n")
    report.append(f"- **収集リソース**: {total_resources}件\n")
    report.append(f"- **違反リソース**: {failed_resources}件 ({resource_failed_rate:.1f}%)\n")
    report.append(f"- **合格リソース**: {passed_resources}件 ({resource_passed_rate:.1f}%)\n")
    if accepted_resources > 0:
        resource_accepted_rate = (accepted_resources / total_relations * 100)
        report.append(f"- **承認済みリソース**: {accepted_resources}件 ({resource_accepted_rate:.1f}%)\n")
    report.append(f"- **コントロール-リソース関連**: {total_relations}件\n\n")

    # 違反コントロールランキング（1件以上の違反があるコントロール）
    # control_id も取得してリンクを生成
    report.append("## 🎯 違反コントロールランキング\n\n")
    cursor.execute(f"""
        SELECT c.control_id, c.name, c.severity, c.objects_count, c.passing_count, c.accepted_count, c.resource_kind
        FROM controls c
        WHERE c.objects_count > 0 {severity_where}
        ORDER BY c.objects_count DESC
    """)

    ranking_controls = cursor.fetchall()
    if ranking_controls:
        report.append("| コントロール名 | 重要度 | 違反数 | 合格数 | 承認数 | リソース種別 |\n")
        report.append("|--------------|--------|--------|--------|--------|-------------|\n")
        for ctrl_id, name, severity, failed, passed, accepted, kind in ranking_controls:
            name_short = name[:50] + "..." if len(name) > 50 else name
            kind_short = kind[:30] if kind else "N/A"
            # コントロールIDをアンカーリンクに使用
            anchor = f"control-{ctrl_id}"
            report.append(f"| [{name_short}](#{anchor}) | {severity} | {failed} | {passed} | {accepted} | {kind_short} |\n")
    else:
        report.append("違反コントロールはありません。\n")

    report.append("\n")

    # fullモードの場合のみトップ10を表示
    if report_mode == 'full':
        report.append("## 🔴 トップ10違反要件\n\n")
        cursor.execute("""
            SELECT requirement_id, name, failed_controls, high_severity_count, medium_severity_count,
                   low_severity_count, description
            FROM compliance_requirements
            ORDER BY failed_controls DESC
            LIMIT 10
        """)

        for idx, row in enumerate(cursor.fetchall(), 1):
            req_id, name, failed, high, medium, low, desc = row
            report.append(f"### {idx}. {name}\n\n")
            report.append(f"- **違反コントロール数**: {failed}件\n")
            report.append(f"- **重要度**: High: {high}, Medium: {medium}, Low: {low}\n")
            report.append(f"- **説明**: {translate_description(desc)}\n\n")

            # トップ5コントロール（説明付き）を表示
            cursor.execute(f"""
                SELECT control_id, name, description, severity, objects_count
                FROM controls
                WHERE requirement_id = ? {severity_where}
                ORDER BY objects_count DESC
                LIMIT 5
            """, (req_id,))

            top_controls = cursor.fetchall()
            if top_controls:
                report.append(f"**主な違反コントロール（上位5件）**:\n\n")
                for ctrl_id, ctrl_name, ctrl_desc, ctrl_sev, ctrl_count in top_controls:
                    report.append(f"- **{ctrl_name}** ({ctrl_sev}, {ctrl_count}件): {translate_description(ctrl_desc)}\n")
                report.append("\n")

        report.append("---\n\n")

    # 詳細レポート
    # ソート順の決定
    sort_order_map = {
        'violations': 'failed_controls DESC',
        'name': 'name ASC',
        'severity': 'severity DESC, failed_controls DESC'
    }
    sort_order = sort_order_map.get(sort_by, 'failed_controls DESC')

    report.append(f"## 📋 詳細レポート（要件別）\n\n")
    report.append(f"**ソート順**: {sort_by}\n\n")

    cursor.execute(f"""
        SELECT requirement_id, name, failed_controls, description
        FROM compliance_requirements
        WHERE failed_controls > 0
        ORDER BY {sort_order}
    """)

    for req_id, req_name, failed, req_desc in cursor.fetchall():
        # 要件名のアンカーID
        req_anchor = make_anchor_id(req_name)
        report.append(f"### <a id=\"{req_anchor}\"></a>{req_name}\n\n")
        report.append(f"**違反コントロール数**: {failed}件\n\n")
        report.append(f"**要件説明**:\n{translate_description(req_desc)}\n\n")

        # この要件に関連するコントロールを取得（severity filterを適用、全件表示）
        cursor.execute(f"""
            SELECT control_id, name, description, severity, objects_count,
                   passing_count, accepted_count, resource_kind
            FROM controls
            WHERE requirement_id = ? {severity_where}
            ORDER BY objects_count DESC
        """, (req_id,))

        controls = cursor.fetchall()
        if controls:
            report.append(f"#### 違反コントロール（全{len(controls)}件）\n\n")

            for ctrl_id, ctrl_name, ctrl_desc, severity, failed_count, passed, accepted, kind in controls:
                # コントロールのアンカーID
                ctrl_anchor = f"control-{ctrl_id}"
                report.append(f"**<a id=\"{ctrl_anchor}\"></a>{ctrl_name}** (ID: {ctrl_id})\n\n")
                report.append(f"- **重要度**: {severity}\n")
                report.append(f"- **違反リソース数**: {failed_count}件\n")
                if passed > 0:
                    report.append(f"- **合格リソース数**: {passed}件\n")
                if accepted > 0:
                    report.append(f"- **承認済み**: {accepted}件\n")
                if kind:
                    report.append(f"- **リソース種別**: `{kind}`\n")
                
                # detailモードの場合は説明を追加
                if report_mode == 'detail':
                    report.append(f"- **説明**: {translate_description(ctrl_desc)}\n")
                
                report.append("\n")

                # Get failed resources for this control (all resources, no limit)
                cursor.execute("""
                    SELECT cr.name, cr.type, cr.account, cr.location, crr.acceptance_status
                    FROM control_resource_relations crr
                    JOIN cloud_resources cr ON crr.resource_hash = cr.hash
                    WHERE crr.control_id = ? AND crr.acceptance_status = 'failed'
                    ORDER BY cr.name
                """, (ctrl_id,))

                failed_resources = cursor.fetchall()
                if failed_resources:
                    report.append(f"**違反リソース（全{len(failed_resources)}件）**:\n\n")
                    report.append("| リソース名 | タイプ | アカウント | リージョン |\n")
                    report.append("|-----------|--------|----------|----------|\n")
                    for res_name, res_type, res_account, res_location, _ in failed_resources:
                        # Truncate long names
                        res_name_short = res_name[:40] + "..." if len(res_name) > 40 else res_name
                        res_type_short = res_type[:20] if res_type else "N/A"
                        res_account_short = res_account[:15] if res_account else "N/A"
                        res_location_short = res_location[:15] if res_location else "N/A"
                        report.append(f"| {res_name_short} | {res_type_short} | {res_account_short} | {res_location_short} |\n")
                    report.append("\n")

                report.append("---\n\n")

        report.append("\n")

    # fullモードの場合のみ統計セクションを表示
    if report_mode == 'full':
        # リソース統計
        report.append("## 📦 影響を受けるリソース統計\n\n")
        cursor.execute("""
            SELECT type, COUNT(*) as count
            FROM cloud_resources
            GROUP BY type
            ORDER BY count DESC
            LIMIT 20
        """)

        report.append("| リソースタイプ | 件数 |\n")
        report.append("|---------------|------|\n")
        for resource_type, count in cursor.fetchall():
            report.append(f"| {resource_type} | {count} |\n")

        report.append("\n")

        # トップ違反コントロール
        report.append("## 🎯 最も違反の多いコントロール（全体）\n\n")
        cursor.execute(f"""
            SELECT name, severity, objects_count, resource_kind, description
            FROM controls
            WHERE 1=1 {severity_where}
            ORDER BY objects_count DESC
            LIMIT 15
        """)

        report.append("| コントロール名 | 重要度 | 違反数 | リソース種別 |\n")
        report.append("|--------------|--------|--------|-------------|\n")
        for name, severity, count, kind, _ in cursor.fetchall():
            name_short = name[:50] + "..." if len(name) > 50 else name
            kind_short = kind[:30] if kind else "N/A"
            report.append(f"| {name_short} | {severity} | {count} | {kind_short} |\n")

        report.append("\n\n")

    report.append("---\n\n")
    report.append("*このレポートは `sysdig-cspm-utils` により自動生成されました。*\n")

    conn.close()

    # ファイルに書き込み
    with open(output_path, 'w', encoding='utf-8') as f:
        f.write(''.join(report))

    print(f"✅ レポートを生成しました: {output_path}")
    print(f"   - 要件: {total_requirements}件")
    print(f"   - コントロール: {total_controls_filtered}件")
    print(f"   - リソース: {total_resources}件")

if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description='コンプライアンス違反レポート生成 (SOC 2, CIS AWS, CIS GCP等に対応)',
        epilog='例: python3 generate_compliance_report.py data/soc2.db report.md --mode full'
    )
    parser.add_argument('db_path', help='SQLiteデータベースのパス (例: data/soc2.db, data/cis_aws.db)')
    parser.add_argument('output_path', help='出力Markdownファイルのパス')
    parser.add_argument(
        '--severity',
        choices=['high', 'all'],
        default='high',
        help='重要度フィルター: high（デフォルト）またはall'
    )
    parser.add_argument(
        '--mode',
        choices=['detail', 'full'],
        default='detail',
        help='レポートモード: detail（詳細のみ、デフォルト）またはfull（トップ10+詳細+統計）'
    )
    parser.add_argument(
        '--sort-by',
        choices=['violations', 'name', 'severity'],
        default='violations',
        help='詳細レポートのソート順: violations（違反数、デフォルト）、name（名前）、severity（重要度）'
    )

    args = parser.parse_args()

    generate_report(args.db_path, args.output_path, args.severity, args.mode, args.sort_by)

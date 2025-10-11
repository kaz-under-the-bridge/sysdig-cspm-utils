# データベーススキーマ設計（v2.0）

## 概要

Sysdig CSPM のコンプライアンスデータを管理するための SQLite データベース設計。
コントロール単位でのリソース管理と Failed/Passed/Accepted の 3 状態分類に対応。

## アーキテクチャ

### 設計方針

1. **コントロール中心設計**: コンプライアンス要件 → コントロール → リソースの階層構造
2. **3 状態管理**: Failed（未受容の違反）、Passed（合格）、Accepted（リスク受容済み）
3. **多対多関係**: 1 つのリソースが複数のコントロールに紐づく可能性を考慮
4. **将来拡張性**: Risk Acceptance 機能の追加に対応

### データフロー

```
Compliance Requirements API
  ↓
Requirements & Controls
  ↓ (resourceApiEndpoint経由)
Cloud Resources API
  ↓ (passed: true/false, acceptance: object/null)
Database Storage
  ├─ compliance_requirements
  ├─ controls
  ├─ cloud_resources
  └─ control_resource_relations
```

## テーブル設計

### 1. compliance_requirements（コンプライアンス要件）

コンプライアンス要件の基本情報を格納。

```sql
CREATE TABLE compliance_requirements (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    requirement_id TEXT NOT NULL,           -- 要件ID（例: "16023"）
    name TEXT NOT NULL,                     -- 要件名（例: "2.1.2 Ensure MFA Delete..."）
    policy_id TEXT NOT NULL,                -- ポリシーID（例: "15"）
    policy_name TEXT NOT NULL,              -- ポリシー名（例: "CIS AWS v3.0.0"）
    policy_type TEXT,                       -- ポリシータイプ（CIS, SOC2, PCI-DSS等）
    platform TEXT,                          -- プラットフォーム（AWS, GCP, Azure）
    severity TEXT NOT NULL,                 -- 重要度（High, Medium, Low）
    pass BOOLEAN NOT NULL,                  -- 合格/不合格
    zone_id TEXT,                           -- ゾーンID
    zone_name TEXT,                         -- ゾーン名
    failed_controls INTEGER DEFAULT 0,      -- 失敗したコントロール数
    high_severity_count INTEGER DEFAULT 0,  -- 高重要度失敗数
    medium_severity_count INTEGER DEFAULT 0,-- 中重要度失敗数
    low_severity_count INTEGER DEFAULT 0,   -- 低重要度失敗数
    accepted_count INTEGER DEFAULT 0,       -- 受容済み数
    passing_count INTEGER DEFAULT 0,        -- 合格数
    description TEXT,                       -- 要件の説明
    resource_api_endpoint TEXT,             -- リソース取得API（後方互換性用）
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(requirement_id, policy_id, zone_id)
);

CREATE INDEX idx_req_policy_type ON compliance_requirements(policy_type);
CREATE INDEX idx_req_platform ON compliance_requirements(platform);
CREATE INDEX idx_req_policy_id ON compliance_requirements(policy_id);
CREATE INDEX idx_req_pass ON compliance_requirements(pass);
CREATE INDEX idx_req_severity ON compliance_requirements(severity);
```

### 2. controls（コントロール）

各要件配下のコントロール情報を格納。

```sql
CREATE TABLE controls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    control_id TEXT NOT NULL UNIQUE,        -- コントロールID（例: "16027"）
    name TEXT NOT NULL,                     -- コントロール名
    description TEXT,                       -- コントロールの説明
    requirement_id TEXT NOT NULL,           -- 親要件ID
    severity TEXT NOT NULL,                 -- 重要度（High, Medium, Low）
    pass BOOLEAN NOT NULL,                  -- 合格/不合格
    is_manual BOOLEAN DEFAULT 0,            -- 手動評価フラグ

    -- カウント情報
    objects_count INTEGER DEFAULT 0,        -- Failed数（passed: false）
    passing_count INTEGER DEFAULT 0,        -- Passed数（passed: true）
    accepted_count INTEGER DEFAULT 0,       -- Accepted数（acceptance != null）

    -- リソース情報
    resource_kind TEXT,                     -- リソース種別（AWS_S3_BUCKET等）
    resource_api_endpoint TEXT NOT NULL,    -- Cloud Resources API URL

    -- メタデータ
    target TEXT,                            -- ターゲット（AWS, GCP等）
    platform TEXT,                          -- プラットフォーム
    authors TEXT,                           -- 作成者
    remediation_id TEXT,                    -- 修正ID
    last_update TEXT,                       -- 最終更新（Unix timestamp）

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (requirement_id) REFERENCES compliance_requirements(requirement_id) ON DELETE CASCADE
);

CREATE INDEX idx_ctrl_requirement ON controls(requirement_id);
CREATE INDEX idx_ctrl_id ON controls(control_id);
CREATE INDEX idx_ctrl_pass ON controls(pass);
CREATE INDEX idx_ctrl_severity ON controls(severity);
CREATE INDEX idx_ctrl_resource_kind ON controls(resource_kind);
```

### 3. cloud_resources（クラウドリソース）

Cloud Resources API および Cluster Analysis API から取得したリソースの詳細情報。
マルチクラウド・マルチプラットフォーム（AWS, GCP, Azure, Docker, Linux Host）に対応。

```sql
CREATE TABLE cloud_resources (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    hash TEXT NOT NULL UNIQUE,              -- リソース一意ハッシュ
    name TEXT NOT NULL,                     -- リソース名
    type TEXT NOT NULL,                     -- リソースタイプ（表示用）

    -- Cloud Resources API用フィールド（AWS/GCP/Azure）
    platform TEXT,                          -- プラットフォーム（AWS, GCP, Azure）
    account TEXT,                           -- AWS: account名, GCP: project ID, Azure: subscription
    location TEXT,                          -- AWS: region, GCP: region, Azure: location
    organization TEXT,                      -- AWS: o-xxxxx, GCP: 組織ID, Azure: tenant ID

    -- Cluster Analysis API用フィールド（Docker/Linux/Kubernetes）
    os_name TEXT,                           -- OS名（linux, windows等）
    os_image TEXT,                          -- OSイメージ（Amazon Linux 2023等）
    cluster_name TEXT,                      -- クラスタ名（K8s等）
    distribution_name TEXT,                 -- ディストリビューション名（Linux等）
    distribution_version TEXT,              -- ディストリビューションバージョン

    -- 共通メタデータ（JSON形式）
    zones_json TEXT,                        -- ゾーン配列
    label_values_json TEXT,                 -- ラベル配列
    additional_metadata_json TEXT,          -- プラットフォーム固有の追加情報
    last_seen_date TEXT,                    -- 最終確認日時（Unix timestamp）
    global_id TEXT,                         -- グローバルID

    -- Cluster Analysis API追加フィールド
    platform_account_id TEXT,               -- プラットフォームアカウントID
    cloud_resource_id TEXT,                 -- クラウドリソースID
    cloud_region TEXT,                      -- クラウドリージョン
    agent_tags_json TEXT,                   -- エージェントタグ配列

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_cloud_res_hash ON cloud_resources(hash);
CREATE INDEX idx_cloud_res_type ON cloud_resources(type);
CREATE INDEX idx_cloud_res_platform ON cloud_resources(platform) WHERE platform IS NOT NULL;
CREATE INDEX idx_cloud_res_account ON cloud_resources(account) WHERE account IS NOT NULL;
CREATE INDEX idx_cloud_res_location ON cloud_resources(location) WHERE location IS NOT NULL;
CREATE INDEX idx_cloud_res_os_name ON cloud_resources(os_name) WHERE os_name IS NOT NULL;
CREATE INDEX idx_cloud_res_cluster_name ON cloud_resources(cluster_name) WHERE cluster_name IS NOT NULL;
```

**フィールドマッピング**:

| API | platform | account | location | os_name | cluster_name |
|-----|----------|---------|----------|---------|--------------|
| Cloud Resources (AWS) | "AWS" | account名 | region | NULL | NULL |
| Cloud Resources (GCP) | "GCP" | project ID | region | NULL | NULL |
| Cloud Resources (Azure) | "Azure" | subscription | location | NULL | NULL |
| Cluster Analysis (Docker) | NULL | NULL | NULL | "linux" | "N/A" |
| Cluster Analysis (Linux) | NULL | NULL | NULL | "linux" | "N/A" |
| Cluster Analysis (K8s) | NULL | NULL | NULL | "linux" | cluster名 |

### 4. control_resource_relations（コントロールとリソースの関連）

コントロールとリソースの多対多関係を管理。各コントロールにおけるリソースの状態を保持。

```sql
CREATE TABLE control_resource_relations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    control_id TEXT NOT NULL,               -- コントロールID
    resource_hash TEXT NOT NULL,            -- リソースハッシュ

    -- 状態情報
    passed BOOLEAN NOT NULL,                -- そのコントロールでのpass/fail
    acceptance_status TEXT NOT NULL,        -- 'failed', 'passed', 'accepted'

    -- Risk Acceptance情報（将来実装用）
    acceptance_justification TEXT,          -- 受容理由
    acceptance_expiration_date TEXT,        -- 受容期限
    acceptance_approved_by TEXT,            -- 承認者
    acceptance_approved_at TEXT,            -- 承認日時

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (control_id) REFERENCES controls(control_id) ON DELETE CASCADE,
    FOREIGN KEY (resource_hash) REFERENCES cloud_resources(hash) ON DELETE CASCADE,
    UNIQUE(control_id, resource_hash)
);

CREATE INDEX idx_rel_control ON control_resource_relations(control_id);
CREATE INDEX idx_rel_resource ON control_resource_relations(resource_hash);
CREATE INDEX idx_rel_status ON control_resource_relations(acceptance_status);
CREATE INDEX idx_rel_passed ON control_resource_relations(passed);
```

## acceptance_status の定義

| 値 | 説明 | 条件 |
|----|------|------|
| `failed` | 違反（未受容） | `passed = false` AND `acceptance_justification IS NULL` |
| `passed` | 合格 | `passed = true` |
| `accepted` | リスク受容済み | `passed = false` AND `acceptance_justification IS NOT NULL` |

## データ投入フロー

### 1. Compliance Requirements の保存

```sql
INSERT INTO compliance_requirements (
    requirement_id, name, policy_id, policy_name,
    platform, severity, pass, failed_controls, description
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
```

### 2. Controls の保存

```sql
INSERT INTO controls (
    control_id, name, description, requirement_id, severity, pass,
    objects_count, passing_count, accepted_count,
    resource_kind, resource_api_endpoint
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
```

### 3. Cloud Resources の保存

```sql
INSERT INTO cloud_resources (
    hash, name, type, platform, account, location, organization,
    zones_json, last_seen_date
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(hash) DO UPDATE SET
    name = excluded.name,
    updated_at = CURRENT_TIMESTAMP;
```

### 4. Relations の保存

```sql
INSERT INTO control_resource_relations (
    control_id, resource_hash, passed, acceptance_status,
    acceptance_justification, acceptance_expiration_date
) VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(control_id, resource_hash) DO UPDATE SET
    passed = excluded.passed,
    acceptance_status = excluded.acceptance_status,
    updated_at = CURRENT_TIMESTAMP;
```

## 統計クエリ例

### コントロール別の統計

```sql
SELECT
    c.control_id,
    c.name,
    c.severity,
    COUNT(CASE WHEN r.acceptance_status = 'failed' THEN 1 END) as failed_count,
    COUNT(CASE WHEN r.acceptance_status = 'passed' THEN 1 END) as passed_count,
    COUNT(CASE WHEN r.acceptance_status = 'accepted' THEN 1 END) as accepted_count,
    COUNT(*) as total_resources
FROM controls c
LEFT JOIN control_resource_relations r ON c.control_id = r.control_id
GROUP BY c.control_id, c.name, c.severity
ORDER BY failed_count DESC;
```

### 要件別の統計（コントロールを集約）

```sql
SELECT
    req.requirement_id,
    req.name,
    req.severity,
    COUNT(DISTINCT c.control_id) as control_count,
    SUM(CASE WHEN rel.acceptance_status = 'failed' THEN 1 ELSE 0 END) as total_failed,
    SUM(CASE WHEN rel.acceptance_status = 'passed' THEN 1 ELSE 0 END) as total_passed,
    SUM(CASE WHEN rel.acceptance_status = 'accepted' THEN 1 ELSE 0 END) as total_accepted
FROM compliance_requirements req
JOIN controls c ON req.requirement_id = c.requirement_id
LEFT JOIN control_resource_relations rel ON c.control_id = rel.control_id
GROUP BY req.requirement_id, req.name, req.severity
ORDER BY total_failed DESC;
```

### アカウント別の違反統計

```sql
SELECT
    res.account,
    res.platform,
    COUNT(CASE WHEN rel.acceptance_status = 'failed' THEN 1 END) as failed_count,
    COUNT(CASE WHEN rel.acceptance_status = 'accepted' THEN 1 END) as accepted_count,
    COUNT(CASE WHEN rel.acceptance_status = 'passed' THEN 1 END) as passed_count
FROM cloud_resources res
JOIN control_resource_relations rel ON res.hash = rel.resource_hash
GROUP BY res.account, res.platform
ORDER BY failed_count DESC;
```

### 特定コントロールのリソース一覧

```sql
-- Failed リソース
SELECT
    res.name,
    res.type,
    res.account,
    res.location
FROM cloud_resources res
JOIN control_resource_relations rel ON res.hash = rel.resource_hash
WHERE rel.control_id = ?
  AND rel.acceptance_status = 'failed'
ORDER BY res.account, res.name;

-- Accepted リソース
SELECT
    res.name,
    res.type,
    res.account,
    res.location,
    rel.acceptance_justification,
    rel.acceptance_expiration_date,
    rel.acceptance_approved_by
FROM cloud_resources res
JOIN control_resource_relations rel ON res.hash = rel.resource_hash
WHERE rel.control_id = ?
  AND rel.acceptance_status = 'accepted'
ORDER BY rel.acceptance_expiration_date;

-- Passed リソース
SELECT
    res.name,
    res.type,
    res.account,
    res.location
FROM cloud_resources res
JOIN control_resource_relations rel ON res.hash = rel.resource_hash
WHERE rel.control_id = ?
  AND rel.acceptance_status = 'passed'
ORDER BY res.account, res.name;
```

## マイグレーション戦略

### 旧スキーマからの移行

現在の `inventory_resources` テーブルから新しいスキーマへの移行：

1. **データバックアップ**: 既存 DB を保存
2. **新スキーマ作成**: 上記 SQL で新テーブル作成
3. **データ再収集**: Cloud Resources API から全データ取得
4. **検証**: 旧データとの比較検証

### 新規実装

新規実装の場合は上記スキーマをそのまま使用。

## パフォーマンス最適化

### インデックス戦略

1. **主キー**: すべてのテーブルで AUTO INCREMENT を使用
2. **外部キー**: JOIN 対象カラムにインデックス
3. **検索頻度の高いカラム**: status, severity, platform 等

### クエリ最適化

1. **EXPLAIN QUERY PLAN** で実行計画を確認
2. **大量データ処理**: バッチ処理と TRANSACTION 使用
3. **集計クエリ**: 必要に応じてマテリアライズドビュー検討

## データ保持ポリシー

### タイムスタンプベースの管理

- `created_at`: レコード作成日時（不変）
- `updated_at`: 最終更新日時（更新時に自動更新）

### 履歴管理（将来実装）

時系列分析のため、過去データをアーカイブテーブルに保存：

```sql
CREATE TABLE compliance_requirements_history (
    -- 同じカラム構成 + snapshot_date
    snapshot_date DATE NOT NULL,
    ...
);
```

## セキュリティ考慮事項

### センシティブデータ

- リソース名に機密情報が含まれる可能性を考慮
- データベースファイルの暗号化検討
- アクセス制御の実装

### データ整合性

- FOREIGN KEY 制約で参照整合性を保証
- UNIQUE 制約で重複データを防止
- NOT NULL 制約で必須フィールドを強制

## 関連ドキュメント

- [CSPM API Integration Guide](./CSPM-API-Integration-Guide.md)
- [Resource API Endpoint Analysis](./resource-api-endpoint-analysis.md)
- [Data Structures](./data-structures.md)

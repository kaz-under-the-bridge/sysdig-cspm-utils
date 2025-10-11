# Sysdig CSPM データ構造定義（v2.0）

## 概要

Sysdig CSPM API と内部データモデルの構造定義。
Cloud Resources API の導入により、コントロール単位でのリソース管理と Failed/Passed/Accepted の状態分類に対応。

## 1. API レスポンス構造

### 1.1 ComplianceRequirement（コンプライアンス要件）

Compliance Results API のレスポンス構造。

```typescript
interface ComplianceRequirement {
  // 識別子
  requirementId: string;          // 要件ID（例: "16023"）
  name: string;                   // 要件名（例: "2.1.2 Ensure MFA Delete..."）

  // ステータス
  pass: boolean;                  // 合格/不合格
  severity: Severity;             // 重要度

  // ポリシー関連
  policyId: string;              // ポリシーID（例: "15"）
  policyName: string;            // ポリシー名（例: "CIS AWS v3.0.0"）

  // カウント情報
  failedControls: number;         // 失敗したコントロール数
  highSeverityCount: number;      // 高重要度の失敗数
  mediumSeverityCount: number;    // 中重要度の失敗数
  lowSeverityCount: number;       // 低重要度の失敗数
  acceptedCount: number;          // リスク承認済み数
  passingCount: number;           // 合格数

  // 関連情報
  description: string;            // 詳細説明
  zone: Zone;                     // 対象ゾーン
  controls: Control[];            // 詳細コントロール配列
}
```

### 1.2 Control（コントロール）

コンプライアンス要件配下のコントロール。

```typescript
interface Control {
  // 識別子
  id: string;                     // コントロールID（例: "16027"）
  name: string;                   // コントロール名

  // 詳細情報
  description: string;            // 説明
  target: string;                 // ターゲット環境（"AWS", "Azure", "GCP"）
  type: ControlType;              // コントロールタイプ
  platform: string;               // プラットフォーム
  authors: string;                // 作成者

  // ステータス
  pass: boolean;                  // 合格/不合格
  severity: Severity;             // 重要度
  isManual: boolean;              // 手動評価かどうか

  // カウント情報（重要）
  objectsCount: number;           // Failed数（passed: false）
  acceptedCount: number;          // Accepted数（acceptance != null）
  passingCount: number;           // Passed数（passed: true）

  // メタデータ
  lastUpdate: string;             // 最終更新（Unix timestamp）
  remediationId: string;          // 修正ID
  resourceKind: ResourceKind;     // リソース種別

  // API情報（重要）
  resourceApiEndpoint: string;    // Cloud Resources API URL
  supportedDistributions: Distribution[];  // サポート環境
}
```

**重要**: `resourceApiEndpoint` を使用して Cloud Resources API からリソース詳細を取得。

### 1.3 CloudResource（クラウドリソース）

Cloud Resources API のレスポンス構造。

```typescript
interface CloudResource {
  // 識別子
  hash: string;                   // リソース一意ハッシュ
  name: string;                   // リソース名
  type: string;                   // リソースタイプ（表示用）

  // 状態（重要）
  passed: boolean;                // そのコントロールでのpass/fail状態
  acceptance: Acceptance | null;  // リスク受容情報

  // クラウドアカウント情報
  account: string;                // アカウント名
  location: string;               // リージョン/ロケーション
  organization: string;           // 組織ID
  platform: string;               // プラットフォーム（AWS, GCP, Azure）

  // メタデータ
  zones: Zone[];                  // 所属ゾーン配列
  labelValues: string[];          // ラベル配列
  lastSeenDate: string;           // 最終確認日時（Unix timestamp）
  globalId: string;               // グローバルID
}
```

**リソース状態の判定ロジック**:
```typescript
function getAcceptanceStatus(resource: CloudResource): AcceptanceStatus {
  if (resource.acceptance !== null) {
    return 'accepted';  // リスク受容済み
  } else if (!resource.passed) {
    return 'failed';    // 違反（未受容）
  } else {
    return 'passed';    // 合格
  }
}
```

### 1.4 Acceptance（リスク受容情報）

```typescript
interface Acceptance {
  justification: string;          // 受容理由
  expirationDate: string;         // 受容期限（ISO 8601形式）
  approvedBy?: string;            // 承認者（将来実装）
  approvedAt?: string;            // 承認日時（将来実装）
}
```

### 1.5 Zone（ゾーン）

```typescript
interface Zone {
  id: string;        // ゾーンID（例: "13010272"）
  name: string;      // ゾーン名（例: "Entire Infrastructure"）
}
```

### 1.6 Distribution（配布環境）

```typescript
interface Distribution {
  name: string;        // 環境名（例: "aws", "Vanilla"）
  minVersion: number;  // 最小サポートバージョン
  maxVersion: number;  // 最大サポートバージョン
}
```

## 2. 列挙型定義

### 2.1 Severity（重要度）

```typescript
enum Severity {
  HIGH = "High",
  MEDIUM = "Medium",
  LOW = "Low"
}

// 数値マッピング（フィルタ・ソート用）
const SeverityValue = {
  [Severity.LOW]: 1,
  [Severity.MEDIUM]: 2,
  [Severity.HIGH]: 3
} as const;
```

### 2.2 ControlType（コントロールタイプ）

```typescript
enum ControlType {
  RESOURCE_EVALUATION = 8,  // リソース評価
  ACCOUNT_EVALUATION = 9    // アカウント評価
}
```

### 2.3 AcceptanceStatus（受容状態）

```typescript
enum AcceptanceStatus {
  FAILED = "failed",      // 違反（未受容）
  PASSED = "passed",      // 合格
  ACCEPTED = "accepted"   // リスク受容済み
}
```

### 2.4 ResourceKind（リソース種別）

```typescript
enum ResourceKind {
  // AWS リソース
  AWS_ACCOUNT = "AWS_ACCOUNT",
  AWS_USER = "AWS_USER",
  AWS_ROOT_USER = "AWS_ROOT_USER",
  AWS_POLICY = "AWS_POLICY",
  AWS_S3_BUCKET = "AWS_S3_BUCKET",
  AWS_S3_VERSIONING_CONFIGURATION = "AWS_S3_VERSIONING_CONFIGURATION",
  AWS_EBS_VOLUME = "AWS_EBS_VOLUME",
  AWS_DATABASE = "AWS_DATABASE",
  AWS_VPC = "AWS_VPC",
  AWS_SECURITY_GROUP = "AWS_SECURITY_GROUP",
  AWS_NETWORK_ACL = "AWS_NETWORK_ACL",
  AWS_ROUTE_TABLE = "AWS_ROUTE_TABLE",
  AWS_CLOUD_TRAIL = "AWS_CLOUD_TRAIL",
  AWS_CONFIGURATION_RECORDER = "AWS_CONFIGURATION_RECORDER",
  AWS_KEY = "AWS_KEY",
  AWS_REGION = "AWS_REGION",

  // GCP リソース
  GCP_PROJECT = "GCP_PROJECT",
  GCP_IAM_SERVICE_ACCOUNT = "GCP_IAM_SERVICE_ACCOUNT",

  // Azure リソース
  AZURE_RESOURCE_GROUP = "AZURE_RESOURCE_GROUP",

  // Kubernetes リソース
  KUBE_WORKLOAD = "workload",
  KUBE_HOST = "host"
}
```

### 2.5 Platform（プラットフォーム）

```typescript
enum Platform {
  AWS = "AWS",
  GCP = "GCP",
  AZURE = "Azure",
  KUBERNETES = "Kubernetes"
}
```

## 3. 内部データモデル

### 3.1 データベースモデル

```typescript
// compliance_requirements テーブル
interface ComplianceRequirementRecord {
  id: number;
  requirement_id: string;
  name: string;
  policy_id: string;
  policy_name: string;
  policy_type: string;
  platform: string;
  severity: string;
  pass: boolean;
  zone_id: string;
  zone_name: string;
  failed_controls: number;
  high_severity_count: number;
  medium_severity_count: number;
  low_severity_count: number;
  accepted_count: number;
  passing_count: number;
  description: string;
  created_at: string;
  updated_at: string;
}

// controls テーブル
interface ControlRecord {
  id: number;
  control_id: string;
  name: string;
  description: string;
  requirement_id: string;
  severity: string;
  pass: boolean;
  is_manual: boolean;
  objects_count: number;
  passing_count: number;
  accepted_count: number;
  resource_kind: string;
  resource_api_endpoint: string;
  target: string;
  platform: string;
  authors: string;
  remediation_id: string;
  last_update: string;
  created_at: string;
  updated_at: string;
}

// cloud_resources テーブル
interface CloudResourceRecord {
  id: number;
  hash: string;
  name: string;
  type: string;
  platform: string;
  account: string;
  location: string;
  organization: string;
  zones_json: string;  // JSON配列
  label_values_json: string;  // JSON配列
  last_seen_date: string;
  global_id: string;
  created_at: string;
  updated_at: string;
}

// control_resource_relations テーブル
interface ControlResourceRelationRecord {
  id: number;
  control_id: string;
  resource_hash: string;
  passed: boolean;
  acceptance_status: AcceptanceStatus;
  acceptance_justification: string | null;
  acceptance_expiration_date: string | null;
  acceptance_approved_by: string | null;
  acceptance_approved_at: string | null;
  created_at: string;
  updated_at: string;
}
```

## 4. API レスポンス構造

### 4.1 Compliance Results API

```typescript
interface ComplianceResponse {
  data: ComplianceRequirement[];
  totalCount: number;
}
```

### 4.2 Cloud Resources API

```typescript
interface CloudResourcesResponse {
  data: CloudResource[];
  totalCount: number;
}
```

### 4.3 ページネーションレスポンス（共通）

```typescript
interface PaginatedResponse<T> {
  data: T[];           // データ配列
  totalCount: number;  // 全体件数
  pageNumber?: number; // 現在のページ番号（省略可）
  pageSize?: number;   // ページサイズ（省略可）
}
```

### 4.4 エラーレスポンス

```typescript
interface ErrorResponse {
  message: string;     // エラーメッセージ
  code?: string;       // エラーコード（省略可）
  details?: any;       // 詳細情報（省略可）
}
```

## 5. 集計データ構造

### 5.1 コントロール統計

```typescript
interface ControlStatistics {
  controlId: string;
  controlName: string;
  severity: Severity;
  failedCount: number;     // acceptance_status = 'failed'
  passedCount: number;     // acceptance_status = 'passed'
  acceptedCount: number;   // acceptance_status = 'accepted'
  totalResources: number;  // 全リソース数
}
```

### 5.2 要件統計

```typescript
interface RequirementStatistics {
  requirementId: string;
  requirementName: string;
  severity: Severity;
  controlCount: number;    // コントロール数
  failedCount: number;     // 全コントロールのfailed合計
  passedCount: number;     // 全コントロールのpassed合計
  acceptedCount: number;   // 全コントロールのaccepted合計
  totalResources: number;  // 全リソース数
}
```

### 5.3 アカウント統計

```typescript
interface AccountStatistics {
  account: string;
  platform: Platform;
  failedCount: number;
  passedCount: number;
  acceptedCount: number;
  totalResources: number;
}
```

### 5.4 全体サマリ

```typescript
interface ComplianceSummary {
  totalRequirements: number;
  totalControls: number;
  totalResources: number;

  failed: {
    requirements: number;
    controls: number;
    resources: number;
  };

  passed: {
    requirements: number;
    controls: number;
    resources: number;
  };

  accepted: {
    requirements: number;
    controls: number;
    resources: number;
  };

  bySeverity: {
    high: SeverityStatistics;
    medium: SeverityStatistics;
    low: SeverityStatistics;
  };
}

interface SeverityStatistics {
  controls: number;
  failed: number;
  passed: number;
  accepted: number;
}
```

## 6. フィルタークエリ構造

### 6.1 フィルタ条件

```typescript
interface FilterCondition {
  field: string;           // フィールド名
  operator: FilterOperator; // 演算子
  value: any;              // 値
}

enum FilterOperator {
  EQUALS = "=",
  NOT_EQUALS = "!=",
  IN = "in",
  NOT_IN = "not in",
  CONTAINS = "contains",
  STARTS_WITH = "startsWith"
}
```

### 6.2 複合フィルタ

```typescript
interface CompositeFilter {
  conditions: (FilterCondition | CompositeFilter)[];
  operator: LogicalOperator;
}

enum LogicalOperator {
  AND = "and",
  OR = "or",
  NOT = "not"
}
```

## 7. 実装ガイドライン

### 7.1 データ型の変換

- **Unix timestamp**: 数値で管理、表示時に Date 型に変換
- **重要度**: API では文字列、内部では enum
- **ブール値**: JSON では `true`/`false`、表示では「合格」/「不合格」
- **JSON フィールド**: `zones_json`, `label_values_json` は文字列として格納、パース時に配列に変換

### 7.2 Null 値の扱い

- オプショナルフィールドは `null` を使用（`undefined` ではない）
- 数値フィールドのデフォルト値は `0`
- 文字列フィールドのデフォルト値は空文字列 `""`

### 7.3 データ整合性

- `resource_hash` はグローバルに一意
- `control_id` と `resource_hash` の組み合わせは一意
- 外部キー制約で参照整合性を保証

### 7.4 パフォーマンス考慮事項

- 大量データ処理時はストリーミング処理を検討
- 頻繁にアクセスされるフィールドはインデックス化
- 集計処理は SQL で実行（アプリケーション層での集計は避ける）

## 8. データフロー

### 8.1 データ収集フロー

```
1. Compliance Results API
   ↓ (Requirements + Controls)
2. 各 Control の resourceApiEndpoint 抽出
   ↓
3. Cloud Resources API（並列実行）
   ↓ (Resources with passed/acceptance)
4. acceptance_status 判定
   ↓
5. Database 保存
   - compliance_requirements
   - controls
   - cloud_resources
   - control_resource_relations
```

### 8.2 データ分析フロー

```
1. Database クエリ
   ↓
2. JOIN 処理（controls + relations + resources）
   ↓
3. 集計処理（GROUP BY, COUNT）
   ↓
4. レポート生成
```

## 9. バリデーション

### 9.1 必須フィールド

- `ComplianceRequirement`: requirementId, name, policyId, severity, pass
- `Control`: controlId, name, requirementId, severity, resourceApiEndpoint
- `CloudResource`: hash, name, type, platform, passed
- `ControlResourceRelation`: controlId, resourceHash, passed, acceptanceStatus

### 9.2 データ範囲

- `severity`: "High", "Medium", "Low" のいずれか
- `acceptanceStatus`: "failed", "passed", "accepted" のいずれか
- `platform`: "AWS", "GCP", "Azure", "Kubernetes" のいずれか
- カウント系フィールド: 0 以上の整数

## 10. 関連ドキュメント

- [Database Schema Design](./database-schema-design.md)
- [CSPM API Integration Guide](./CSPM-API-Integration-Guide.md)
- [Resource API Endpoint Analysis](./resource-api-endpoint-analysis.md)

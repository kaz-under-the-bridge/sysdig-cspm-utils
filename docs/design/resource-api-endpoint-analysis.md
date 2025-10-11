# resourceApiEndpoint から取得できる情報の調査結果

## 調査したコントロール

### 1. Network ACL - Control 16071
**Endpoint**: `/api/cspm/v1/cloud/resources?controlId=16071&providerType=AWS&resourceKind=AWS_NETWORK_ACL&filter=policyId=15 and zones.id=13010272`

**統計**:
- Objects Count: 433
- Accepted Count: 0
- Passing Count: 0
- Total Count (API): 433

### 2. S3 MFA Delete - Control 16027
**Endpoint**: `/api/cspm/v1/cloud/resources?controlId=16027&providerType=AWS&resourceKind=AWS_S3_VERSIONING_CONFIGURATION&filter=policyId=15 and zones.id=13010272`

**統計**:
- Objects Count: 138
- Accepted Count: 118
- Passing Count: 0
- Total Count (API): 256

### 3. IAM Policy - Control 16018
**Endpoint**: `/api/cspm/v1/cloud/resources?controlId=16018&providerType=AWS&resourceKind=AWS_POLICY&filter=policyId=15 and zones.id=13010272`

**統計**:
- Objects Count: 40
- Accepted Count: 0
- Passing Count: 1608
- Total Count (API): 1648

## 取得できるリソースデータ構造

```json
{
  "data": [
    {
      "name": "リソース名",
      "passed": true/false,        // ★このコントロールでのpass/fail状態
      "hash": "ハッシュ値",
      "type": "リソースタイプ",
      "account": "AWSアカウント名",
      "location": "リージョン",
      "organization": "組織ID",
      "acceptance": null,           // acceptされた場合に情報が入る
      "zones": [
        {
          "id": "ゾーンID",
          "name": "ゾーン名"
        }
      ],
      "lastSeenDate": "最終確認日時（Unix timestamp）",
      "labelValues": [],
      "globalId": ""
    }
  ],
  "totalCount": 全リソース数（passed + failed）
}
```

## 重要な発見

### 1. **passed フィールド**
- 各リソースに `passed: true/false` フィールドがある
- **そのコントロールに対してリソースがpass/failしているかが明示的にわかる**
- これにより、failed/passedリソースを正確に分類可能

### 2. **totalCount の意味**
- `totalCount` = passed + failed の全リソース数
- Compliance Requirements APIの `objectsCount` はfailedのみ
- Compliance Requirements APIの `passingCount` はpassedのみ
- 例: Control 16018
  - objectsCount: 40 (failed)
  - passingCount: 1608 (passed)
  - totalCount: 1648 (40 + 1608)

### 3. **acceptance フィールド**
- リスクが受容（accept）された場合に情報が入る
- acceptされたリソースは `acceptedCount` にカウントされる

## 現在の実装との違い

### 現在の実装 (`/api/cspm/v1/inventory/resources`)
- ポリシーレベルでの一括取得
- フィルター: `policy.failed in ("CIS AWS v3.0.0")`
- コントロールとの紐付けが不明確
- リソースに `passed` フィールドなし

### resourceApiEndpoint (`/api/cspm/v1/cloud/resources`)
- **コントロール別に個別取得可能**
- controlId、resourceKindでフィルタリング
- **各リソースに `passed: true/false` が明示される**
- passed/failedの正確な分類が可能

## 推奨される使い方

1. Compliance Requirements APIで違反要件とコントロールを取得
2. 各コントロールの `resourceApiEndpoint` を使用してリソース取得
3. `passed: false` でfailedリソース、`passed: true` でpassedリソースを分類
4. コントロール単位での詳細な分析が可能

## メリット

- ✅ コントロール単位でリソースを正確に特定できる
- ✅ passed/failedの明確な分類
- ✅ より細かい粒度での分析が可能
- ✅ 各コントロールの remediation に必要なリソースリストが取得できる

## API仕様の詳細

### エンドポイント形式
```
/api/cspm/v1/cloud/resources?controlId={controlId}&providerType={providerType}&resourceKind={resourceKind}&filter=policyId={policyId} and zones.id={zoneId}
```

### パラメータ
- `controlId`: コントロールID（例: 16027）
- `providerType`: プロバイダータイプ（AWS, GCP, Azure等）
- `resourceKind`: リソース種別（AWS_S3_BUCKET, AWS_POLICY等）
- `filter`: フィルター条件（policyId, zones.id等）
- `pageSize`: ページサイズ（デフォルト: 50）
- `pageNumber`: ページ番号（1から開始）

### ページネーション
- `/api/cspm/v1/inventory/resources` と同様にページネーション対応
- `totalCount` フィールドで全件数を取得
- `pageSize` と `pageNumber` パラメータで分割取得

## 実装への影響

### 現在の実装の課題
1. ポリシーレベルでの一括取得のため、コントロールとリソースの紐付けが不明
2. データベースに保存されたリソースから、特定のコントロールでfailedしているリソースを特定できない
3. passed/failedの判定ができない

### 改善案
1. コントロール別のリソース取得を追加実装
2. データベースにコントロールとリソースの関連テーブルを追加
3. リソースごとのpass/fail状態を保存
4. より詳細なレポート生成が可能になる

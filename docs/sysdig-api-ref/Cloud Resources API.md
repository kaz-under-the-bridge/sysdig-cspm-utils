# Cloud Resources API

## 概要

Sysdig CSPM Cloud Resources APIは、特定のコントロールに紐づくクラウドリソースの詳細情報を取得するためのAPIです。Compliance Requirements APIの各コントロールに含まれる `resourceApiEndpoint` フィールドを使用してアクセスします。

## エンドポイント

```
GET /api/cspm/v1/cloud/resources
```

## 認証

Bearer Token認証を使用します。

```
Authorization: Bearer {SYSDIG_API_TOKEN}
```

## リクエストパラメータ

| パラメータ | 型 | 必須 | 説明 | 例 |
|-----------|-----|------|------|-----|
| `controlId` | string | Yes | コントロールID | `16027` |
| `providerType` | string | Yes | クラウドプロバイダータイプ | `AWS`, `GCP`, `Azure`, `Kubernetes` |
| `resourceKind` | string | Yes | リソース種別 | `AWS_S3_BUCKET`, `AWS_POLICY`, `AWS_NETWORK_ACL` |
| `filter` | string | No | フィルター条件 | `policyId=15 and zones.id=13010272` |
| `pageSize` | integer | No | ページサイズ（デフォルト: 50） | `50` |
| `pageNumber` | integer | No | ページ番号（1から開始） | `1` |

### フィルター構文

`filter` パラメータでは以下の条件を指定できます：

- `policyId={id}` - ポリシーIDでフィルタ
- `zones.id={id}` - ゾーンIDでフィルタ
- 複数条件は `and` で連結

## レスポンス形式

### 成功レスポンス (200 OK)

```json
{
  "data": [
    {
      "name": "リソース名",
      "passed": true,
      "hash": "ハッシュ値",
      "type": "リソースタイプ",
      "account": "アカウント名",
      "location": "リージョン",
      "organization": "組織ID",
      "acceptance": null,
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
  "totalCount": 256
}
```

### レスポンスフィールド

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `data` | array | リソース配列 |
| `data[].name` | string | リソース名 |
| `data[].passed` | boolean | **そのコントロールでのpass/fail状態**（false=failed, true=passed） |
| `data[].hash` | string | リソースの一意なハッシュ値 |
| `data[].type` | string | リソースタイプ（表示用） |
| `data[].account` | string | クラウドアカウント名 |
| `data[].location` | string | リージョンまたはロケーション |
| `data[].organization` | string | 組織ID |
| `data[].acceptance` | object/null | リスク受容情報（acceptされた場合に値が入る） |
| `data[].zones` | array | 所属ゾーン情報 |
| `data[].lastSeenDate` | string | 最終確認日時（Unix timestampのミリ秒） |
| `data[].labelValues` | array | ラベル配列 |
| `data[].globalId` | string | グローバルID |
| `totalCount` | integer | 全リソース数（passed + failed） |

## リクエスト例

### 例1: S3 MFA Delete コントロールのリソース取得

```bash
curl -X GET "https://us2.app.sysdig.com/api/cspm/v1/cloud/resources?controlId=16027&providerType=AWS&resourceKind=AWS_S3_VERSIONING_CONFIGURATION&filter=policyId=15%20and%20zones.id=13010272&pageSize=50&pageNumber=1" \
  -H "Authorization: Bearer ${SYSDIG_API_TOKEN}" \
  -H "Content-Type: application/json"
```

### 例2: Network ACL コントロールのリソース取得

```bash
curl -X GET "https://us2.app.sysdig.com/api/cspm/v1/cloud/resources?controlId=16071&providerType=AWS&resourceKind=AWS_NETWORK_ACL&filter=policyId=15%20and%20zones.id=13010272&pageSize=50&pageNumber=1" \
  -H "Authorization: Bearer ${SYSDIG_API_TOKEN}" \
  -H "Content-Type: application/json"
```

### 例3: IAM Policy コントロールのリソース取得

```bash
curl -X GET "https://us2.app.sysdig.com/api/cspm/v1/cloud/resources?controlId=16018&providerType=AWS&resourceKind=AWS_POLICY&filter=policyId=15%20and%20zones.id=13010272&pageSize=50&pageNumber=1" \
  -H "Authorization: Bearer ${SYSDIG_API_TOKEN}" \
  -H "Content-Type: application/json"
```

## レスポンス例

### 例1: S3 MFA Delete (Control 16027)

```json
{
  "data": [
    {
      "name": "p4t-k7w9m3x-artifacts-download - Versioning",
      "passed": false,
      "hash": "912d0a04113fe91d",
      "type": "S3 Versioning Configuration",
      "account": "p4t-k7w9m3x",
      "location": "ap-northeast-1",
      "organization": "o-r8x5pd2mqh",
      "acceptance": null,
      "zones": [
        {
          "id": "13010272",
          "name": "Entire Infrastructure"
        }
      ],
      "lastSeenDate": "1759577199912",
      "labelValues": [],
      "globalId": ""
    },
    {
      "name": "p4t-k7w9m3x-release-tfstate - Versioning",
      "passed": false,
      "hash": "dd60c6c7dfe6780e",
      "type": "S3 Versioning Configuration",
      "account": "p4t-k7w9m3x",
      "location": "ap-northeast-1",
      "organization": "o-r8x5pd2mqh",
      "acceptance": null,
      "zones": [
        {
          "id": "13010272",
          "name": "Entire Infrastructure"
        }
      ],
      "lastSeenDate": "1759577199912",
      "labelValues": [],
      "globalId": ""
    }
  ],
  "totalCount": 256
}
```

### 例2: Network ACL (Control 16071)

```json
{
  "data": [
    {
      "name": "acl-0c2cdec879be495af",
      "passed": false,
      "hash": "7d07f3276988135",
      "type": "Network ACL",
      "account": "s6q-k7w9m3x",
      "location": "ap-northeast-1",
      "organization": "o-r8x5pd2mqh",
      "acceptance": null,
      "zones": [
        {
          "id": "13010272",
          "name": "Entire Infrastructure"
        }
      ],
      "lastSeenDate": "1759580454899",
      "labelValues": [],
      "globalId": ""
    }
  ],
  "totalCount": 433
}
```

### 例3: IAM Policy (Control 16018)

```json
{
  "data": [
    {
      "name": "AdministratorAccess (ANPAXJ9M2KFPR8TQ3VWHN)",
      "passed": false,
      "hash": "1dab1f244ac14409",
      "type": "IAM Policy",
      "account": "s6q-k7w9m3x",
      "location": "global",
      "organization": "o-r8x5pd2mqh",
      "acceptance": null,
      "zones": [
        {
          "id": "13010272",
          "name": "Entire Infrastructure"
        }
      ],
      "lastSeenDate": "1759580709402",
      "labelValues": [],
      "globalId": ""
    }
  ],
  "totalCount": 1648
}
```

## ページネーション

このAPIはページネーションをサポートしています。

### ページネーションの実装例

```bash
# 最初のページを取得してtotalCountを確認
curl -X GET "https://us2.app.sysdig.com/api/cspm/v1/cloud/resources?controlId=16027&providerType=AWS&resourceKind=AWS_S3_VERSIONING_CONFIGURATION&filter=policyId=15%20and%20zones.id=13010272&pageSize=50&pageNumber=1" \
  -H "Authorization: Bearer ${SYSDIG_API_TOKEN}" \
  -H "Content-Type: application/json"

# totalCountから総ページ数を計算
# totalPages = (totalCount + pageSize - 1) / pageSize

# 2ページ目以降を取得
curl -X GET "https://us2.app.sysdig.com/api/cspm/v1/cloud/resources?controlId=16027&providerType=AWS&resourceKind=AWS_S3_VERSIONING_CONFIGURATION&filter=policyId=15%20and%20zones.id=13010272&pageSize=50&pageNumber=2" \
  -H "Authorization: Bearer ${SYSDIG_API_TOKEN}" \
  -H "Content-Type: application/json"
```

## エラーレスポンス

### 401 Unauthorized

認証エラー

```json
{
  "message": "no credentials in request"
}
```

### 400 Bad Request

不正なリクエストパラメータ

```json
{
  "message": "Invalid parameter"
}
```

## 使用上の注意

### Rate Limiting

APIには Rate Limit が設定されています。大量のリソースを取得する場合は、適切な遅延を設定してください。

推奨設定：
- バッチサイズ: 3-5リクエスト/バッチ
- バッチ間遅延: 1-3秒

### totalCount の意味

- `totalCount` は **passed + failed の全リソース数** を表します
- Compliance Requirements APIの統計値との対応：
  - `objectsCount` = failedのリソース数（passed=false）
  - `passingCount` = passedのリソース数（passed=true）
  - Cloud Resources APIの `totalCount` ≈ `objectsCount` + `passingCount`

### passed フィールドの重要性

- **`passed: false`**: そのコントロールで失敗しているリソース
- **`passed: true`**: そのコントロールで合格しているリソース

このフィールドを使用することで、コントロール単位での正確なpass/fail判定が可能です。

## Compliance Requirements API との関係

### 取得フロー

1. **Compliance Requirements APIでコントロール情報を取得**
   ```bash
   GET /api/cspm/v1/compliance/requirements?filter=policy.name contains "CIS Amazon" and pass = "false"
   ```

2. **レスポンスから `resourceApiEndpoint` を抽出**
   ```json
   {
     "controls": [
       {
         "id": "16027",
         "name": "S3 - Enabled MFA Delete",
         "objectsCount": 138,
         "passingCount": 0,
         "resourceApiEndpoint": "/api/cspm/v1/cloud/resources?controlId=16027&providerType=AWS&resourceKind=AWS_S3_VERSIONING_CONFIGURATION&filter=policyId=15 and zones.id=13010272"
       }
     ]
   }
   ```

3. **Cloud Resources APIでリソース詳細を取得**
   ```bash
   GET {resourceApiEndpoint}
   ```

### データの対応関係

| Compliance Requirements API | Cloud Resources API |
|-----------------------------|---------------------|
| `controls[].objectsCount` | `passed: false` のリソース数 |
| `controls[].passingCount` | `passed: true` のリソース数 |
| `controls[].acceptedCount` | `acceptance != null` のリソース数 |
| - | `totalCount` = objectsCount + passingCount |

## 実装例

### Go言語での実装例

```go
// GetCloudResources retrieves resources for a specific control
func (c *CSPMClient) GetCloudResources(controlID, providerType, resourceKind, filter string, pageSize, pageNumber int) (*CloudResourcesResponse, error) {
    endpoint := "/api/cspm/v1/cloud/resources"

    params := url.Values{}
    params.Set("controlId", controlID)
    params.Set("providerType", providerType)
    params.Set("resourceKind", resourceKind)
    if filter != "" {
        params.Set("filter", filter)
    }
    if pageNumber > 0 {
        params.Set("pageNumber", strconv.Itoa(pageNumber))
    }
    if pageSize > 0 {
        params.Set("pageSize", strconv.Itoa(pageSize))
    }

    fullURL := endpoint + "?" + params.Encode()

    resp, err := c.Client.MakeRequest("GET", fullURL, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to get cloud resources: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
    }

    var response CloudResourcesResponse
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }

    return &response, nil
}

// CloudResourcesResponse represents the response from Cloud Resources API
type CloudResourcesResponse struct {
    Data       []CloudResource `json:"data"`
    TotalCount int             `json:"totalCount"`
}

// CloudResource represents a single cloud resource
type CloudResource struct {
    Name         string   `json:"name"`
    Passed       bool     `json:"passed"`
    Hash         string   `json:"hash"`
    Type         string   `json:"type"`
    Account      string   `json:"account"`
    Location     string   `json:"location"`
    Organization string   `json:"organization"`
    Acceptance   *Acceptance `json:"acceptance"`
    Zones        []Zone   `json:"zones"`
    LastSeenDate string   `json:"lastSeenDate"`
    LabelValues  []string `json:"labelValues"`
    GlobalID     string   `json:"globalId"`
}

type Acceptance struct {
    // Acceptance structure details
}

type Zone struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}
```

## 関連APIドキュメント

- [Compliance Results API](./Compliance%20Results.md)
- [Inventory Resources API](./Search%20and%20list%20Inventory%20Resources.md)
- [CSPM API Integration Guide](./CSPM-API-Integration-Guide.md)

## 参考資料

- [Resource API Endpoint Analysis](../design/resource-api-endpoint-analysis.md) - 詳細な調査結果

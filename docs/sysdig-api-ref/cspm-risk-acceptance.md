# CSPM Risk Acceptance API

## Overview

Sysdig CSPMのリスク受容（Risk Acceptance）APIを使用して、検出された違反に対してリスクを受容することができます。

## API Endpoint

```
POST https://us2.app.sysdig.com/api/cspm/v1/compliance/violations/acceptances
```

## Authentication

```
Authorization: Bearer {API_TOKEN}
```

## Request

### Request Body Schema

**Content-Type:** `application/json`

### Parameters

| パラメータ | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| `controlId` | integer | ✅ | リスクを受容するPosture Control ID |
| `description` | string | - | 受容に関する追加説明 |
| `expiresAt` | string | - | リスク受容の有効期限（Unix timestamp, milliseconds）。指定しない場合は期限なし |
| `filter` | string | - | 結果をフィルタリングするためのクエリ言語式 |
| `reason` | string | ✅ | 受容理由 |
| `sourceId` | string | - | リスクを受容するAccount ID/Cluster/Host |
| `zoneId` | integer | - | リスクを受容するZone ID |

### Filter Query Language

**サポートされるオペレーター:** `in`

**フィルタ可能なフィールド:**

| フィールド | 型 | 説明 | 例 |
|-----------|-----|------|-----|
| `name` | string | リソース名 | `name in ("cf-templates-1s951ca3qbh1-us-west-2")` |
| `namespace` | string | Namespace（K8s等） | `namespace in ("my-namespace")` |
| `kind` | string | リソース種別 | `kind in ("AWS_S3_BUCKET")` |
| `location` | string | クラウドのロケーション/リージョン | `location in ("ap-southeast-2")` |
| `providerType` | string | クラウドプロバイダー | `providerType in ("AWS")` |

### Request Example

```json
{
  "controlId": 1,
  "description": "Jane - will take care of it",
  "expiresAt": "1660742030427",
  "filter": "location in (\"us-west-2\") and name in (\"cf-templates-1s951ca3qbh1-us-west-2\")",
  "reason": "Risk Owned",
  "sourceId": "012345678901",
  "zoneId": 7
}
```

## Response

### Response Schema

**Content-Type:** `application/json`

### Response Fields

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `id` | string | 受容レコードのID |
| `acceptPeriod` | string | 受容期間（日数） |
| `acceptanceDate` | string | 受容日時（Unix timestamp, milliseconds） |
| `controlId` | integer | Control ID |
| `description` | string | 追加説明 |
| `expiresAt` | string | 有効期限（Unix timestamp, milliseconds） |
| `isExpired` | boolean | 期限切れかどうか |
| `reason` | string | 受容理由 |
| `sourceId` | string | Account ID/Cluster/Host |
| `userDisplayName` | string | 受容を実行したユーザーの表示名 |
| `username` | string | 受容を実行したユーザーのメールアドレス |
| `zoneId` | integer | Zone ID |

### Response Example

```json
{
  "acceptPeriod": "30",
  "acceptanceDate": "1660742030427",
  "controlId": 1,
  "description": "Jane - will take care of it",
  "expiresAt": "1663361999999",
  "id": "62fce98ebc19e98141f04f1f",
  "isExpired": false,
  "reason": "Risk Owned",
  "sourceId": "string",
  "userDisplayName": "Jane Doe",
  "username": "jane.doe@myorg.com",
  "zoneId": 7
}
```

## 関連API

### Control ID検索API

Control名からControl IDを取得するAPI:

```
GET https://us2.app.sysdig.com/api/cspm/v1/policy/controls/search
```

**Query Parameters:**
- `filter`: `name="<control_name>"`

**Response:**
```json
{
  "data": [
    {
      "id": 1,
      "name": "Control Name"
    }
  ]
}
```

## Python実装サンプル

### 一括リスク受容スクリプト

CSVファイルからリソース一覧を読み込み、一括でリスク受容を設定するスクリプト。

```python
import requests
import argparse
import os
import logging

API_TOKEN = os.environ.get('SYSDIG_API_TOKEN')

def parse_args():
    parser = argparse.ArgumentParser(
        description="Bulk set risk acceptance for Sysdig CSPM",
        usage="""
        cspm.py --name <control_name> --file <path_to_file>

        bulk set risk acceptance for sysdig cspm

        arguments:
            -h, --help  ヘルプの表示
            --name      CSPMのコントロール名（検知名）
            --file      Risk Acceptanceを指定するリスト（report出力機能でdownloadしたcsvファイルを加工したもの）
        """
    )

    parser.add_argument(
        "--name",
        type=str,
        required=True,
        help="CSPMのコントロール名"
    )
    parser.add_argument(
        "--file",
        type=str,
        required=True,
        help="Risk Acceptanceの対象ファイル名"
    )

    return parser

def set_logging_config(level=logging.INFO):
    LOG_FORMAT = '%(asctime)s@ %(name)s [%(levelname)s] %(funcName)s: %(message)s'
    stream_handler = logging.StreamHandler()
    stream_handler.setLevel(level)
    stream_handler.setFormatter(logging.Formatter(LOG_FORMAT))
    logging.basicConfig(level=level, handlers=[stream_handler])

class HttpRequestHelper:
    @staticmethod
    def request(method, url, headers=None, params=None, json=None):
        default_headers = {
            "Authorization": f"Bearer {API_TOKEN}"
        }
        headers = {**default_headers, **(headers or {})}

        try:
            response = requests.request(method, url, headers=headers, params=params, json=json)
            print(response.status_code)
            response.raise_for_status()
            return response.json()
        except requests.exceptions.HTTPError as http_err:
            print(f"HTTP error occurred: {http_err}")
        except Exception as err:
            print(f"An error occurred: {err}")
        return None

class Cspm:
    def get_control_id_by_name(self, control_name):
        url = "https://us2.app.sysdig.com/api/cspm/v1/policy/controls/search"

        params = {
            "filter": f'name="{control_name}"'
        }

        data = HttpRequestHelper.request("GET", url, None, params=params)
        if data and 'data' in data and data['data']:
            return data['data'][0].get('id')
        return "Id Not Found"

class CSPMRiskAcceptance:
    def set_acceptance(self, control_id, name, resource_id):
        url = "https://us2.app.sysdig.com/api/cspm/v1/compliance/violations/acceptances"

        json = {
            "controlId": control_id,
            "filter": f'name in ("{name}")',
            "expiresAt": None,
            "reason": "Risk Owned",
            "description": "CSPM棚卸しでRiskAcceptanceを設定",
            "sourceId": resource_id,  # optional: Specific GCP Project-ID can be added to acceptance using SourceID parameter.
        }

        return HttpRequestHelper.request("POST", url, None, None, json)

if __name__ == '__main__':
    parser = parse_args()

    name = parser.parse_args().name
    file = parser.parse_args().file

    cspm = Cspm()
    control_id = cspm.get_control_id_by_name(name)

    if control_id == "Id Not Found":
        print(f"Control Name: {name} が見つかりませんでした")
        exit(1)

    cspm_ra = CSPMRiskAcceptance()

    with open(file, 'r') as file:
        for line in file:
            line = line.strip()
            parts = line.split()

            if len(parts) == 1:
                cspm_ra.set_acceptance(control_id, parts[0], "")
            elif len(parts) == 2:
                cspm_ra.set_acceptance(control_id, parts[0], parts[1])
```

## 使用例

### 基本的な使用方法

1. **Control IDを取得**
   ```bash
   curl -X GET "https://us2.app.sysdig.com/api/cspm/v1/policy/controls/search?filter=name%3D%22IAM%20-%20Defined%20Users%20MFA%22" \
     -H "Authorization: Bearer ${SYSDIG_API_TOKEN}"
   ```

2. **リスク受容を設定**
   ```bash
   curl -X POST "https://us2.app.sysdig.com/api/cspm/v1/compliance/violations/acceptances" \
     -H "Authorization: Bearer ${SYSDIG_API_TOKEN}" \
     -H "Content-Type: application/json" \
     -d '{
       "controlId": 15001,
       "filter": "name in (\"test-bucket\")",
       "reason": "Risk Owned",
       "description": "Approved by security team"
     }'
   ```

### 特定リージョンのリソースに対する受容

```bash
curl -X POST "https://us2.app.sysdig.com/api/cspm/v1/compliance/violations/acceptances" \
  -H "Authorization: Bearer ${SYSDIG_API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "controlId": 16031,
    "filter": "location in (\"us-west-2\") and providerType in (\"AWS\")",
    "reason": "Risk Owned",
    "description": "US West 2 region is approved",
    "sourceId": "012345678901"
  }'
```

### 期限付きリスク受容

30日間の期限付きでリスクを受容する例:

```python
import time

# 現在時刻から30日後のUnix timestamp（ミリ秒）
expires_at = str(int((time.time() + 30 * 24 * 60 * 60) * 1000))

acceptance_data = {
    "controlId": 16027,
    "filter": "name in (\"my-s3-bucket\")",
    "reason": "Temporary exception",
    "description": "Will be fixed in next sprint",
    "expiresAt": expires_at
}
```

## 注意事項

1. **Control ID**: Control名ではなくControl IDを使用する必要があります。Control ID検索APIで事前に取得してください。

2. **Filter式**: 複数条件を組み合わせる場合は`and`演算子を使用します。
   ```
   location in ("us-west-2") and name in ("bucket-name")
   ```

3. **sourceId**: AWSの場合はAccount ID、GCPの場合はProject ID、K8sの場合はCluster名を指定します。

4. **有効期限**: `expiresAt`を指定しない場合、リスク受容は無期限となります。

5. **既存の受容**: 同じcontrolId + filterの組み合わせで既に受容が存在する場合、更新されます。

## トラブルシューティング

### エラー: 400 Bad Request

- `controlId`が存在しない、またはfilter式の構文が間違っている可能性があります
- Filter式の引用符のエスケープを確認してください

### エラー: 401 Unauthorized

- APIトークンが無効または期限切れです
- `Authorization`ヘッダーの形式を確認してください: `Bearer {token}`

### エラー: 404 Not Found

- エンドポイントURLが間違っている可能性があります
- リージョン（us2, eu1など）を確認してください

# Sysdig CSPM フィルター設計書

## 概要

Sysdig CSPM APIのフィルター機能の設計仕様。CLIからのフィルター指定方法とAPI呼び出し時のエンコード処理を定義する。

## フィルター処理フロー

```
Shell Script → Go CLI (デコード文字列) → Go内部でURLエンコード → Sysdig API
```

## CLI引数仕様

### `-policy` (ポリシー名フィルター)

- **形式**: 文字列（カンマ区切りで複数指定可能）
- **マッチング**: 部分一致（`contains`演算子）
- **デフォルト**: 空（全ポリシー対象）

#### 使用例

```bash
# 単一ポリシー指定
./cspm-utils -policy "CIS Amazon Web Services Foundations Benchmark v3.0.0"

# 複数ポリシー指定（カンマ区切り）
./cspm-utils -policy "CIS AWS,SOC 2,CIS GCP"

# 部分一致
./cspm-utils -policy "CIS"  # CIS AWS, CIS GCP, CIS Azure などにマッチ
```

#### 実際のポリシー名例

- `CIS Amazon Web Services Foundations Benchmark v3.0.0`
- `CIS Google Cloud Platform Foundation Benchmark v2.0.0`
- `SOC 2`
- `CIS Docker Benchmark`
- `CIS Distribution Independent Linux Benchmark`

### `-platform` (プラットフォームフィルター)

- **形式**: 文字列（単一指定）
- **マッチング**: 完全一致（`=`演算子）
- **デフォルト**: 空（全プラットフォーム対象）
- **有効値**: `AWS`, `GCP`, `Azure`, `Kubernetes`

#### 使用例

```bash
./cspm-utils -platform "AWS"
./cspm-utils -platform "GCP"
```

### `-zone` (ゾーン名フィルター)

- **形式**: 文字列
- **マッチング**: 完全一致（`in`演算子）
- **デフォルト**: `"Entire Infrastructure"`

#### 使用例

```bash
./cspm-utils -zone "Entire Infrastructure"
./cspm-utils -zone "Production Environment"
```

## API フィルター構文

### 基本構文

```
pass = "false" and <conditions>
```

### 演算子

| 演算子 | 説明 | 使用例 |
|--------|------|--------|
| `=` | 完全一致 | `pass = "false"` |
| `!=` | 不一致 | `pass != "true"` |
| `in` | リスト内一致 | `policy.name in ("CIS AWS", "SOC 2")` |
| `contains` | 部分一致 | `policy.name contains "CIS"` |
| `startsWith` | 前方一致 | `name startsWith "1."` |
| `and` | 論理積 | `pass = "false" and severity = 3` |
| `or` | 論理和 | `severity = 3 or severity = 2` |
| `not` | 論理否定 | `not pass = "true"` |

### フィルター対象フィールド

| フィールド | 型 | 説明 | 例 |
|-----------|-----|------|-----|
| `pass` | boolean | 合格/不合格 | `pass = "false"` |
| `policy.name` | string | ポリシー名 | `policy.name contains "CIS AWS"` |
| `zone.name` | string | ゾーン名 | `zone.name in ("Entire Infrastructure")` |
| `platform` | string | プラットフォーム | `platform = "AWS"` |
| `severity` | integer | 重要度 | `severity in (3, 2)` (1=Low, 2=Medium, 3=High) |
| `name` | string | 要件名 | `name contains "1.5"` |

## URLエンコード仕様

### エンコード対象文字

| 文字 | エンコード後 | 備考 |
|------|-------------|------|
| スペース | `+` または `%20` | APIは両方受け入れ可能、`+`推奨 |
| `"` | `%22` | クォート文字 |
| `(` | `%28` | 左括弧 |
| `)` | `%29` | 右括弧 |
| `,` | `%2C` | カンマ |
| `=` | `%3D` | 等号（パラメータ値内のみ） |

### エンコード例

#### 入力（デコード）
```
pass = "false" and policy.name contains "CIS Amazon Web Services Foundations Benchmark v3.0.0" and zone.name in ("Entire Infrastructure")
```

#### 出力（エンコード）
```
pass+%3D+%22false%22+and+policy.name+contains+%22CIS+Amazon+Web+Services+Foundations+Benchmark+v3.0.0%22+and+zone.name+in+%28%22Entire+Infrastructure%22%29
```

## Go実装仕様

### buildFilter関数

複数条件を受け取り、Sysdig APIフィルター文字列を構築する。

#### 関数シグネチャ

```go
func buildFilter(policies, platform, zoneName string) string
```

#### パラメータ

- `policies`: カンマ区切りポリシー名（部分一致）
- `platform`: プラットフォーム名（完全一致）
- `zoneName`: ゾーン名（完全一致）

#### 処理フロー

1. 基本条件: `pass = "false"` を設定
2. ポリシーフィルター構築:
   - カンマで分割
   - 各ポリシー名に対して`contains`条件を生成
   - `or`で結合
3. プラットフォームフィルター追加（指定時）
4. ゾーンフィルター追加（指定時）
5. 条件を`and`で結合

#### 実装例

```go
func buildFilter(policies, platform, zoneName string) string {
    conditions := []string{`pass = "false"`}

    // ポリシーフィルター（複数対応、部分一致）
    if policies != "" {
        policyList := strings.Split(policies, ",")
        policyConditions := []string{}
        for _, p := range policyList {
            p = strings.TrimSpace(p)
            if p != "" {
                policyConditions = append(policyConditions,
                    fmt.Sprintf(`policy.name contains "%s"`, p))
            }
        }
        if len(policyConditions) > 0 {
            conditions = append(conditions,
                fmt.Sprintf("(%s)", strings.Join(policyConditions, " or ")))
        }
    }

    // プラットフォームフィルター（完全一致）
    if platform != "" {
        conditions = append(conditions,
            fmt.Sprintf(`platform = "%s"`, platform))
    }

    // ゾーンフィルター（完全一致、in演算子）
    if zoneName != "" {
        conditions = append(conditions,
            fmt.Sprintf(`zone.name in ("%s")`, zoneName))
    }

    return strings.Join(conditions, " and ")
}
```

#### 出力例

```go
// 入力
buildFilter("CIS AWS,SOC 2", "AWS", "Entire Infrastructure")

// 出力（エンコード前）
pass = "false" and (policy.name contains "CIS AWS" or policy.name contains "SOC 2") and platform = "AWS" and zone.name in ("Entire Infrastructure")
```

### URLエンコード処理

Go標準の`net/url`パッケージを使用してエンコード。

```go
import "net/url"

params := url.Values{}
params.Set("filter", filterString)
encodedURL := endpoint + "?" + params.Encode()
```

`url.Values.Encode()`は自動的に以下を実施:
- スペース → `+`
- 特殊文字 → `%XX`形式

## シェルスクリプト使用例

### 単一ポリシー

```bash
#!/bin/bash

# CIS AWS v3.0.0のみ
./bin/cspm-utils \
  -token "${SYSDIG_API_TOKEN}" \
  -command collect \
  -policy "CIS Amazon Web Services Foundations Benchmark v3.0.0" \
  -zone "Entire Infrastructure" \
  -db "data/cis_aws.db"
```

### 複数ポリシー（カンマ区切り）

```bash
#!/bin/bash

# AWS系CISとSOC2を対象
./bin/cspm-utils \
  -token "${SYSDIG_API_TOKEN}" \
  -command collect \
  -policy "CIS AWS,SOC 2" \
  -platform "AWS" \
  -zone "Entire Infrastructure" \
  -db "data/aws_compliance.db"
```

### 部分一致を活用

```bash
#!/bin/bash

# すべてのCIS関連（AWS, GCP, Azure, Docker等）
./bin/cspm-utils \
  -token "${SYSDIG_API_TOKEN}" \
  -command collect \
  -policy "CIS" \
  -zone "Entire Infrastructure" \
  -db "data/all_cis.db"
```

### プラットフォーム別

```bash
#!/bin/bash

# GCPのみ
./bin/cspm-utils \
  -token "${SYSDIG_API_TOKEN}" \
  -command collect \
  -platform "GCP" \
  -zone "Entire Infrastructure" \
  -db "data/gcp_compliance.db"
```

## テストケース

### 基本ケース

| # | 入力 (policies, platform, zone) | 期待される出力 |
|---|--------------------------------|---------------|
| 1 | `"", "", ""` | `pass = "false"` |
| 2 | `"CIS AWS", "", ""` | `pass = "false" and policy.name contains "CIS AWS"` |
| 3 | `"", "AWS", ""` | `pass = "false" and platform = "AWS"` |
| 4 | `"", "", "Entire Infrastructure"` | `pass = "false" and zone.name in ("Entire Infrastructure")` |

### 複合ケース

| # | 入力 | 期待される出力 |
|---|------|---------------|
| 5 | `"CIS AWS", "AWS", "Entire Infrastructure"` | `pass = "false" and policy.name contains "CIS AWS" and platform = "AWS" and zone.name in ("Entire Infrastructure")` |
| 6 | `"CIS AWS,SOC 2", "", ""` | `pass = "false" and (policy.name contains "CIS AWS" or policy.name contains "SOC 2")` |

### エッジケース

| # | 入力 | 説明 | 期待される出力 |
|---|------|------|---------------|
| 7 | `"CIS AWS, SOC 2"` (スペース含む) | トリム処理 | `pass = "false" and (policy.name contains "CIS AWS" or policy.name contains "SOC 2")` |
| 8 | `"CIS AWS,,SOC 2"` (空要素) | 空要素スキップ | `pass = "false" and (policy.name contains "CIS AWS" or policy.name contains "SOC 2")` |

## 注意事項

### 特殊文字のエスケープ

ポリシー名に`"`や`\`が含まれる場合は、Sysdig APIドキュメントに従いエスケープが必要。

```
Example: policy.name contains "CIS \"Special\" Benchmark"
→ エスケープ: policy.name contains "CIS \\"Special\\" Benchmark"
→ URLエンコード: policy.name+contains+%22CIS+%5C%22Special%5C%22+Benchmark%22
```

### ページネーション

フィルター条件に関わらず、ページネーション処理が必要。
- デフォルト: pageSize=10
- 推奨: pageSize=50（最大）
- 大量データ: `GetAllComplianceRequirements()`使用

### パフォーマンス

複数ポリシー指定時は、`or`条件が増えるため応答時間が増加する可能性あり。
- 3ポリシー以下推奨
- 大量ポリシー指定時は個別実行を検討

# Inception Phase

**目的**: 計画、要件収集、アーキテクチャ決定

**焦点**: **何を**作り、**なぜ**作るかを決定する

---

## ステージ一覧

| ステージ | 実行条件 | 目的 |
|---------|---------|------|
| Workspace Detection | 常に実行 | 既存コードの有無判定 |
| Reverse Engineering | Brownfieldのみ | 既存コード分析 |
| Requirements Analysis | 常に実行（深度は適応的） | 要件収集・明確化 |
| User Stories | 条件付き | ユーザーストーリー作成 |
| Workflow Planning | 常に実行 | 実行計画策定 |
| Application Design | 条件付き | コンポーネント設計 |
| Units Generation | 条件付き | 作業単位への分解 |

---

## 1. Workspace Detection（常に実行）

### 目的
ワークスペースの状態を判定し、Greenfield（新規）かBrownfield（既存）かを識別する。

### 実行内容
1. `aidlc-docs/aidlc-state.md`の存在確認（存在すれば再開）
2. ソースコードファイルのスキャン（.java, .py, .ts, .go等）
3. ビルドファイルの確認（pom.xml, package.json, go.mod等）
4. プロジェクト構造の識別

### 出力
```markdown
## Workspace State
- **Existing Code**: [Yes/No]
- **Programming Languages**: [List if found]
- **Build System**: [Maven/Gradle/npm/etc.]
- **Project Structure**: [Monolith/Microservices/Library/Empty]
- **Workspace Root**: [Absolute path]
```

### 次フェーズ判定
- **空のワークスペース**: Requirements Analysis へ
- **既存コードあり + リバエン成果物なし**: Reverse Engineering へ
- **既存コードあり + リバエン成果物あり**: Requirements Analysis へ

---

## 2. Reverse Engineering（Brownfieldのみ）

### 目的
既存コードベースを分析し、文書化する。

### 実行条件
- 既存コードが検出された
- リバースエンジニアリング成果物が存在しない

### 生成成果物
| ファイル | 内容 |
|---------|------|
| `architecture.md` | システムアーキテクチャ概要 |
| `component-inventory.md` | コンポーネント一覧 |
| `code-structure.md` | コード構造 |
| `api-documentation.md` | API仕様 |
| `technology-stack.md` | 技術スタック |
| `dependencies.md` | 依存関係 |
| `interaction-diagrams.md` | ビジネストランザクションの実装フロー |

### 承認
**明示的な承認が必要** - 完了メッセージ表示後、ユーザー確認まで待機

---

## 3. Requirements Analysis（常に実行）

### 目的
ユーザーリクエストを分析し、要件を明確化する。

### 適応的深度

| 深度 | 条件 | 内容 |
|-----|------|------|
| **Minimal** | シンプル・明確なリクエスト | 基本的な理解の文書化 |
| **Standard** | 通常の複雑さ | 機能・非機能要件の収集 |
| **Comprehensive** | 複雑・高リスク | 詳細要件＋トレーサビリティ |

### 分析項目

#### リクエスト明確度
- **Clear**: 具体的、明確、実行可能
- **Vague**: 一般的、曖昧、要明確化
- **Incomplete**: 重要情報が欠落

#### リクエストタイプ
- New Feature / Bug Fix / Refactoring / Upgrade / Migration / Enhancement / New Project

#### スコープ推定
- Single File → Single Component → Multiple Components → System-wide → Cross-system

### 実行ステップ
1. リバースエンジニアリング成果物を読み込み（Brownfieldの場合）
2. ユーザーリクエスト分析（Intent Analysis）
3. 要件深度の決定
4. 完全性分析（機能・非機能・ユーザーシナリオ・ビジネスコンテキスト）
5. 明確化質問の生成（`requirement-verification-questions.md`）
6. 回答収集・曖昧さ解消
7. 要件ドキュメント生成（`requirements.md`）

### 生成成果物
- `aidlc-docs/inception/requirements/requirements.md`
- `aidlc-docs/inception/requirements/requirement-verification-questions.md`

---

## 4. User Stories（条件付き）

### 目的
要件をユーザー中心のストーリーに変換し、受け入れ基準を定義する。

### 実行判定

#### 常に実行（高優先度）
- 新しいユーザー向け機能
- ユーザーワークフローへの影響
- 複数のユーザーペルソナ
- 複雑なビジネス要件
- チーム間コラボレーション必要
- 顧客向けAPI/サービス変更

#### 評価して判断（中優先度）
- 既存ユーザー向け機能の修正
- ユーザー体験に間接的影響を与えるバックエンド変更
- ユーザーワークフローに影響するインテグレーション
- ユーザーに見えるパフォーマンス改善

#### スキップ（低優先度のみ）
- 純粋な内部リファクタリング（ユーザー影響ゼロ）
- 明確なスコープの単純バグ修正
- インフラ変更のみ
- 技術的負債の解消
- 開発ツール改善

### 2部構成

#### Part 1: Planning
1. ストーリー計画作成（チェックボックス付き）
2. コンテキスト適切な質問生成
3. 回答収集・曖昧さ分析
4. 計画の承認取得

#### Part 2: Generation
1. 承認された計画に従ってストーリー生成
2. ペルソナ生成
3. INVEST基準でストーリー検証
4. 承認取得

### 生成成果物
- `aidlc-docs/inception/user-stories/stories.md`
- `aidlc-docs/inception/user-stories/personas.md`
- `aidlc-docs/inception/plans/story-generation-plan.md`

### INVEST基準
- **I**ndependent: 独立している
- **N**egotiable: 交渉可能
- **V**aluable: 価値がある
- **E**stimable: 見積もり可能
- **S**mall: 小さい
- **T**estable: テスト可能

---

## 5. Workflow Planning（常に実行）

### 目的
どのフェーズ/ステージを実行するか決定し、実行計画を作成する。

### 分析内容

#### スコープ・影響分析
1. **変換スコープ検出**（Brownfieldのみ）
   - 単一コンポーネント変更 vs アーキテクチャ変換
   - インフラ変更 vs アプリケーション変更
   - デプロイモデル変更（Lambda→Container等）

2. **変更影響評価**
   - ユーザー向け変更
   - 構造変更
   - データモデル変更
   - API変更
   - NFR影響

3. **リスク評価**
   - Low / Medium / High / Critical

#### フェーズ決定

| ステージ | 実行条件 |
|---------|---------|
| User Stories | 複数ペルソナ、UX影響、受け入れ基準必要 |
| Application Design | 新コンポーネント/サービス、メソッド定義必要 |
| Units Generation | システム分解必要、複数サービス/モジュール |

### 生成成果物
- `aidlc-docs/inception/plans/execution-plan.md`
  - Mermaidフローチャート
  - 各フェーズのEXECUTE/SKIP判定と理由
  - パッケージ更新順序（Brownfieldのみ）
  - 成功基準

---

## 6. Application Design（条件付き）

### 目的
高レベルのコンポーネント識別とサービス層設計。

### 実行条件
- 新コンポーネント/サービスが必要
- コンポーネントメソッドとビジネスルールの定義が必要
- サービス層設計が必要
- コンポーネント依存関係の明確化が必要

### スキップ条件
- 既存コンポーネント境界内の変更
- 新コンポーネント/メソッドなし
- 純粋な実装変更

### 生成成果物
- `aidlc-docs/inception/application-design/components.md`
- `aidlc-docs/inception/application-design/component-methods.md`
- `aidlc-docs/inception/application-design/services.md`
- `aidlc-docs/inception/application-design/component-dependency.md`

**注意**: 詳細なビジネスロジック設計はConstruction PhaseのFunctional Designで行う

---

## 7. Units Generation（条件付き）

### 目的
システムを管理可能な作業単位（Unit of Work）に分解する。

### 実行条件
- 複数ユニットへの分解が必要
- 複数サービス/モジュールが必要
- 構造化された分解が必要な複雑システム

### スキップ条件
- 単一のシンプルなユニット
- 分解不要
- 単一コンポーネント実装

### 用語定義
- **Service**: 独立デプロイ可能なコンポーネント
- **Module**: サービス内の論理グループ
- **Unit of Work**: 計画コンテキストでの作業単位

### 2部構成

#### Part 1: Planning
1. 分解計画作成
2. コンテキスト適切な質問生成
3. 回答収集・曖昧さ分析
4. 計画の承認取得

#### Part 2: Generation
1. ユニット定義生成
2. 依存関係マトリクス作成
3. ストーリーマッピング
4. 承認取得

### 生成成果物
- `aidlc-docs/inception/application-design/unit-of-work.md`
- `aidlc-docs/inception/application-design/unit-of-work-dependency.md`
- `aidlc-docs/inception/application-design/unit-of-work-story-map.md`
- `aidlc-docs/inception/plans/unit-of-work-plan.md`

---

## 状態管理

### aidlc-state.md

```markdown
# AI-DLC State Tracking

## Project Information
- **Project Type**: [Greenfield/Brownfield]
- **Start Date**: [ISO timestamp]
- **Current Stage**: INCEPTION - [Stage Name]

## Workspace State
- **Existing Code**: [Yes/No]
- **Reverse Engineering Needed**: [Yes/No]
- **Workspace Root**: [Absolute path]

## Stage Progress
### INCEPTION PHASE
- [x] Workspace Detection
- [x] Reverse Engineering (if applicable)
- [x] Requirements Analysis
- [ ] User Stories - [EXECUTE/SKIP]
- [ ] Workflow Planning
- [ ] Application Design - [EXECUTE/SKIP]
- [ ] Units Generation - [EXECUTE/SKIP]

## Current Status
- **Lifecycle Phase**: INCEPTION
- **Current Stage**: [Stage Name]
- **Next Stage**: [Next Stage Name]
- **Status**: [Status]
```

---

## 承認フォーマット

各ステージ完了時の標準フォーマット：

```markdown
# [Emoji] [Stage Name] Complete

[AI-generated summary in bullet points]

> **REVIEW REQUIRED:**
> Please examine the artifacts at: `aidlc-docs/inception/[directory]/`

> **WHAT'S NEXT?**
>
> **You may:**
>
> **Request Changes** - Ask for modifications
> **Approve & Continue** - Approve and proceed to **[Next Stage]**
```

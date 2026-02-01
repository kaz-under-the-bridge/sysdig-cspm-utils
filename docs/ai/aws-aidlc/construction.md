# Construction Phase

**目的**: 詳細設計、NFR実装、コード生成

**焦点**: **どのように**作るかを決定する

---

## 概要

Construction Phaseは**ユニット単位でループ**して実行される。
各ユニットは完全に完了（設計＋コード）してから次のユニットへ進む。

```
┌─────────────────────────────────────────────────────────┐
│              Per-Unit Loop（各ユニットで繰り返し）         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ Functional   │→│ NFR          │→│ NFR          │  │
│  │ Design       │  │ Requirements │  │ Design       │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
│          ↓                                   ↓          │
│  ┌──────────────┐  ┌──────────────┐                    │
│  │Infrastructure│→│ Code         │                    │
│  │ Design       │  │ Generation   │                    │
│  └──────────────┘  └──────────────┘                    │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│  ┌──────────────────────────────────────────────────┐  │
│  │         Build and Test（全ユニット完了後）         │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

---

## ステージ一覧

| ステージ | 実行条件 | 目的 |
|---------|---------|------|
| Functional Design | 条件付き（per-unit） | 詳細ビジネスロジック設計 |
| NFR Requirements | 条件付き（per-unit） | 非機能要件の評価 |
| NFR Design | 条件付き（per-unit） | NFRパターンの組み込み |
| Infrastructure Design | 条件付き（per-unit） | インフラサービスマッピング |
| Code Generation | 常に実行（per-unit） | コード生成・実装 |
| Build and Test | 常に実行（全ユニット後） | ビルド・テスト実行 |

---

## 1. Functional Design（条件付き、per-unit）

### 目的
ユニットごとの詳細ビジネスロジック設計。技術非依存で純粋にビジネス機能に焦点。

### 実行条件
- 新しいデータモデル/スキーマ
- 複雑なビジネスロジック
- ビジネスルールの詳細設計が必要

### スキップ条件
- 単純なロジック変更
- 新しいビジネスロジックなし

### 焦点領域
- 詳細なビジネスロジック・アルゴリズム
- ドメインモデル（エンティティと関係）
- ビジネスルール・検証ロジック・制約
- **技術非依存**（インフラ懸念なし）

### 質問カテゴリ（必要に応じて）
- **Business Logic Modeling**: コアエンティティ、ワークフロー、データ変換
- **Domain Model**: ドメイン概念、エンティティ関係、データ構造
- **Business Rules**: 決定ルール、検証ロジック、制約
- **Data Flow**: データ入出力、変換、永続化要件
- **Integration Points**: 外部システム連携、API、データ交換
- **Error Handling**: エラーシナリオ、検証失敗、例外処理
- **Business Scenarios**: エッジケース、代替フロー

### 生成成果物
- `aidlc-docs/construction/{unit-name}/functional-design/business-logic-model.md`
- `aidlc-docs/construction/{unit-name}/functional-design/business-rules.md`
- `aidlc-docs/construction/{unit-name}/functional-design/domain-entities.md`

---

## 2. NFR Requirements（条件付き、per-unit）

### 目的
ユニットの非機能要件を決定し、技術スタック選択を行う。

### 実行条件
- パフォーマンス要件あり
- セキュリティ考慮が必要
- スケーラビリティ懸念あり
- 技術スタック選択が必要

### スキップ条件
- NFR要件なし
- 技術スタック決定済み

### 質問カテゴリ（必要に応じて）
- **Scalability Requirements**: 予想負荷、成長パターン、スケーリングトリガー
- **Performance Requirements**: 応答時間、スループット、レイテンシ
- **Availability Requirements**: 稼働時間、災害復旧、フェイルオーバー
- **Security Requirements**: データ保護、コンプライアンス、認証認可
- **Tech Stack Selection**: 技術選好、制約、既存システム連携
- **Reliability Requirements**: エラー処理、耐障害性、監視
- **Maintainability Requirements**: コード品質、ドキュメント、テスト

### 生成成果物
- `aidlc-docs/construction/{unit-name}/nfr-requirements/nfr-requirements.md`
- `aidlc-docs/construction/{unit-name}/nfr-requirements/tech-stack-decisions.md`

---

## 3. NFR Design（条件付き、per-unit）

### 目的
NFR要件をパターンと論理コンポーネントを使ってユニット設計に組み込む。

### 実行条件
- NFR Requirementsが実行された
- NFRパターンの組み込みが必要

### スキップ条件
- NFR要件なし
- NFR Requirements Assessmentがスキップされた

### 質問カテゴリ（必要に応じて）
- **Resilience Patterns**: 耐障害性アプローチ
- **Scalability Patterns**: スケーリングメカニズム
- **Performance Patterns**: パフォーマンス最適化戦略
- **Security Patterns**: セキュリティ実装アプローチ
- **Logical Components**: インフラコンポーネント（キュー、キャッシュ等）

### 生成成果物
- `aidlc-docs/construction/{unit-name}/nfr-design/nfr-design-patterns.md`
- `aidlc-docs/construction/{unit-name}/nfr-design/logical-components.md`

---

## 4. Infrastructure Design（条件付き、per-unit）

### 目的
論理的なソフトウェアコンポーネントを実際のインフラ選択にマッピング。

### 実行条件
- インフラサービスのマッピングが必要
- デプロイアーキテクチャが必要
- クラウドリソースの指定が必要

### スキップ条件
- インフラ変更なし
- インフラ定義済み

### 質問カテゴリ（必要に応じて）
- **Deployment Environment**: クラウドプロバイダー、環境セットアップ
- **Compute Infrastructure**: コンピュートサービス選択
- **Storage Infrastructure**: データベース、ストレージ選択
- **Messaging Infrastructure**: メッセージング/キューイングサービス
- **Networking Infrastructure**: ロードバランシング、APIゲートウェイ
- **Monitoring Infrastructure**: 可観測性ツール
- **Shared Infrastructure**: インフラ共有戦略

### 生成成果物
- `aidlc-docs/construction/{unit-name}/infrastructure-design/infrastructure-design.md`
- `aidlc-docs/construction/{unit-name}/infrastructure-design/deployment-architecture.md`
- `aidlc-docs/construction/shared-infrastructure.md`（共有インフラの場合）

---

## 5. Code Generation（常に実行、per-unit）

### 目的
各ユニットのコードを生成する。

### 2部構成

#### Part 1: Planning
1. ユニットコンテキスト分析
2. 詳細なコード生成計画作成
3. 承認取得

#### Part 2: Generation
1. 計画に従ってコード生成
2. テスト生成
3. ドキュメント生成
4. 承認取得

### コード生成ステップ（計画に含める）
1. Project Structure Setup（Greenfieldのみ）
2. Business Logic Generation
3. Business Logic Unit Testing
4. Business Logic Summary
5. API Layer Generation
6. API Layer Unit Testing
7. API Layer Summary
8. Repository Layer Generation
9. Repository Layer Unit Testing
10. Repository Layer Summary
11. Database Migration Scripts（データモデルがある場合）
12. Documentation Generation
13. Deployment Artifacts Generation

### コード配置ルール

| 対象 | 配置場所 |
|-----|---------|
| アプリケーションコード | ワークスペースルート（**決して**aidlc-docs/には置かない） |
| ドキュメント | `aidlc-docs/construction/{unit-name}/code/`（Markdownのみ） |
| ビルド/設定ファイル | ワークスペースルート |

### プロジェクト構造パターン

| プロジェクトタイプ | 構造 |
|------------------|------|
| Brownfield | 既存構造を使用（`src/main/java/`, `lib/`, `pkg/`等） |
| Greenfield単一ユニット | `src/`, `tests/`, `config/` |
| Greenfieldマルチユニット（マイクロサービス） | `{unit-name}/src/`, `{unit-name}/tests/` |
| Greenfieldマルチユニット（モノリス） | `src/{unit-name}/`, `tests/{unit-name}/` |

### Brownfieldファイル修正ルール
- ファイル存在確認後に生成
- **存在する場合**: インプレースで修正（**決して**`ClassName_modified.java`等のコピーを作成しない）
- **存在しない場合**: 新規ファイル作成
- 生成後に重複ファイルがないか確認

### 生成成果物
- 実際のソースコード（ワークスペースルート）
- `aidlc-docs/construction/plans/{unit-name}-code-generation-plan.md`
- `aidlc-docs/construction/{unit-name}/code/`（Markdownサマリー）

---

## 6. Build and Test（常に実行、全ユニット完了後）

### 目的
全ユニットをビルドし、包括的なテスト戦略を実行。

### テスト種類

| テスト | 目的 |
|-------|------|
| Unit tests | 各ユニットの単体テスト（Code Generationで生成済み） |
| Integration tests | ユニット/サービス間の連携テスト |
| Performance tests | 負荷、ストレス、スケーラビリティテスト |
| End-to-end tests | 完全なユーザーワークフローテスト |
| Contract tests | サービス間APIコントラクト検証 |
| Security tests | 脆弱性スキャン、ペネトレーションテスト |

### 生成成果物

```
aidlc-docs/construction/build-and-test/
├── build-instructions.md           # ビルド手順
├── unit-test-instructions.md       # 単体テスト実行手順
├── integration-test-instructions.md # 統合テスト手順
├── performance-test-instructions.md # パフォーマンステスト手順（該当時）
├── contract-test-instructions.md   # コントラクトテスト（マイクロサービス）
├── security-test-instructions.md   # セキュリティテスト
├── e2e-test-instructions.md        # E2Eテスト
└── build-and-test-summary.md       # 結果サマリー
```

### build-instructions.md の内容
- 前提条件（ビルドツール、依存関係、環境変数）
- ビルドステップ
- 成功時の期待出力
- トラブルシューティング

### build-and-test-summary.md の内容
```markdown
# Build and Test Summary

## Build Status
- **Build Tool**: [Tool name]
- **Build Status**: [Success/Failed]
- **Build Artifacts**: [List]

## Test Execution Summary

### Unit Tests
- **Total Tests**: [X]
- **Passed**: [X]
- **Failed**: [X]
- **Coverage**: [X]%

### Integration Tests
- **Test Scenarios**: [X]
- **Passed**: [X]
- **Failed**: [X]

### Performance Tests
- **Response Time**: [Actual] (Target: [Expected])
- **Throughput**: [Actual] (Target: [Expected])
- **Error Rate**: [Actual] (Target: [Expected])

## Overall Status
- **Build**: [Success/Failed]
- **All Tests**: [Pass/Fail]
- **Ready for Operations**: [Yes/No]
```

---

## 重要ルール

### Planning Phase Rules
- 明示的な番号付きステップを作成
- ストーリートレーサビリティを含める
- ユニットコンテキストと依存関係を文書化
- 生成前にユーザー承認を取得

### Generation Phase Rules
- **NO HARDCODED LOGIC**: 計画に書かれたことのみ実行
- **FOLLOW PLAN EXACTLY**: ステップ順序から逸脱しない
- **UPDATE CHECKBOXES**: 各ステップ完了後すぐに[x]マーク
- **STORY TRACEABILITY**: 機能実装時にユニットストーリーを[x]マーク
- **RESPECT DEPENDENCIES**: 依存関係が満たされた時のみ実装

### 標準完了メッセージ
Construction Phaseでは**2オプション**の完了メッセージを使用：

```markdown
> **WHAT'S NEXT?**
>
> **You may:**
>
> **Request Changes** - Ask for modifications based on your review
> **Continue to Next Stage** - Approve and proceed to **[next-stage-name]**
```

**重要**: 3オプションメニューや他の新規パターンを作成しない

---

## 完了基準

### Per-Unit完了
- 該当する全設計ステージが完了
- コード生成計画が作成・承認済み
- 計画内の全ステップが[x]マーク済み
- 全ユニットストーリーが計画通り実装済み
- 全コード・テストが生成済み

### Phase完了
- 全ユニットのCode Generationが完了
- Build and Testが成功
- 全テストがパス
- Operations Phaseへの準備完了

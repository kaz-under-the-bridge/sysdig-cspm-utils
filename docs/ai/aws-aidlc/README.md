# AWS AI-DLC (AI-Driven Development Life Cycle)

AI-DLCは、AIと人間が協調してソフトウェア開発を進めるための構造化されたワークフローです。

## 概要

AI-DLCは3つのフェーズで構成されます：

| フェーズ | 目的 | 焦点 |
|---------|------|------|
| **INCEPTION** | 計画・要件収集・設計 | **何を**作り、**なぜ**作るか |
| **CONSTRUCTION** | 詳細設計・実装・テスト | **どのように**作るか |
| **OPERATIONS** | デプロイ・監視（将来拡張） | どのように**運用**するか |

## 基本原則

### 1. Human in the Loop
- AIが提案し、人間が承認する
- 各ステージ完了時に**明示的な承認**が必要
- ユーザーはステージの追加・スキップを制御可能

### 2. 適応的ワークフロー
- ワークフローは問題に適応する（逆ではない）
- 複雑さに応じて詳細度が変化
- 単純な変更は効率的に、複雑な変更は包括的に

### 3. 質問駆動アプローチ
- チャットではなく**専用ファイル**で質問
- 選択式＋Otherオプションで明確化
- 曖昧さは必ず解消してから次へ

### 4. 監査証跡
- 全てのユーザー入力をaudit.mdに記録
- 完全な生入力を保存（要約しない）
- ISO 8601タイムスタンプ使用

## ワークフロー図

```
┌──────────────────────────────────────────────────────────────┐
│                     INCEPTION PHASE                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐     │
│  │Workspace │→│Reverse   │→│Require-  │→│User      │      │
│  │Detection │  │Engineer  │  │ments     │  │Stories   │      │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘     │
│        ↓                                        ↓            │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                   │
│  │Workflow  │→│App       │→│Units     │                    │
│  │Planning  │  │Design    │  │Generation│                    │
│  └──────────┘  └──────────┘  └──────────┘                   │
└──────────────────────────────────────────────────────────────┘
                              ↓
┌──────────────────────────────────────────────────────────────┐
│                   CONSTRUCTION PHASE                         │
│              [各ユニットごとにループ]                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐     │
│  │Functional│→│NFR       │→│NFR       │→│Infra     │      │
│  │Design    │  │Require.  │  │Design    │  │Design    │      │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘     │
│                                                  ↓            │
│  ┌──────────┐  ┌──────────┐                                  │
│  │Code      │→│Build &   │                                  │
│  │Generation│  │Test      │                                  │
│  └──────────┘  └──────────┘                                  │
└──────────────────────────────────────────────────────────────┘
```

## 成果物ディレクトリ構造

```
<WORKSPACE-ROOT>/                   # アプリケーションコード
├── [project-specific structure]
│
├── aidlc-docs/                     # ドキュメントのみ
│   ├── inception/                  # INCEPTION PHASE
│   │   ├── plans/
│   │   ├── reverse-engineering/    # Brownfieldのみ
│   │   ├── requirements/
│   │   ├── user-stories/
│   │   └── application-design/
│   ├── construction/               # CONSTRUCTION PHASE
│   │   ├── plans/
│   │   ├── {unit-name}/
│   │   │   ├── functional-design/
│   │   │   ├── nfr-requirements/
│   │   │   ├── nfr-design/
│   │   │   ├── infrastructure-design/
│   │   │   └── code/               # Markdownサマリーのみ
│   │   └── build-and-test/
│   ├── operations/                 # OPERATIONS PHASE（プレースホルダー）
│   ├── aidlc-state.md              # 進捗状態
│   └── audit.md                    # 監査ログ
```

**重要ルール**:
- アプリケーションコード: ワークスペースルート（**決して** aidlc-docs/には置かない）
- ドキュメント: aidlc-docs/のみ

## 関連ドキュメント

- [Inception Phase](./inception.md) - 計画・要件収集フェーズ
- [Construction Phase](./construction.md) - 設計・実装フェーズ
- [Question Format](./question-format.md) - 質問フォーマットガイド
- [Adaptive Depth](./adaptive-depth.md) - 適応的深度の説明

## 参照元

- [AWS Labs AIDLC Workflows](https://github.com/awslabs/aidlc-workflows)
- [AWS Blog: AI-Driven Development Life Cycle](https://aws.amazon.com/blogs/devops/ai-driven-development-life-cycle/)

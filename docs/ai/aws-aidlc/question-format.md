# Question Format Guide

AI-DLCでは、チャットではなく**専用の質問ファイル**を使用して構造化された質問を行う。

---

## 基本ルール

1. **チャットで直接質問しない** - 全ての質問は専用ファイルに配置
2. **[Answer]:タグを使用** - ユーザーの回答場所を明確化
3. **選択肢 + Other** - 常に「Other」オプションを最後に含める
4. **曖昧さは解消してから進む** - フォローアップ質問で明確化

---

## ファイル命名規則

```
{phase-name}-questions.md
```

例:
- `classification-questions.md`
- `requirements-questions.md`
- `story-planning-questions.md`
- `design-questions.md`

---

## 質問構造

```markdown
## Question [Number]
[明確で具体的な質問文]

A) [選択肢1]
B) [選択肢2]
C) [選択肢3]（必要に応じて追加）
D) [選択肢4]（必要に応じて追加）
X) Other (please describe after [Answer]: tag below)

[Answer]:
```

### 選択肢の数
- **最小**: 2つの意味のある選択肢 + Other (A, B, C)
- **典型**: 3-4つの意味のある選択肢 + Other (A, B, C, D, E)
- **最大**: 5つの意味のある選択肢 + Other (A, B, C, D, E, F)

**重要**: 選択肢を埋めるために無理に作らない。意味のある選択肢のみ含める。

---

## 完全な例

```markdown
# Requirements Clarification Questions

Please answer the following questions to help clarify the requirements.

## Question 1
What is the primary user authentication method?

A) Username and password
B) Social media login (Google, Facebook)
C) Single Sign-On (SSO)
D) Multi-factor authentication
E) Other (please describe after [Answer]: tag below)

[Answer]:

## Question 2
Will this be a web or mobile application?

A) Web application
B) Mobile application
C) Both web and mobile
D) Other (please describe after [Answer]: tag below)

[Answer]:

## Question 3
Is this a new project or existing codebase?

A) New project (greenfield)
B) Existing codebase (brownfield)
C) Other (please describe after [Answer]: tag below)

[Answer]:
```

---

## ユーザー回答フォーマット

ユーザーは[Answer]:タグの後に選択肢の文字を記入：

```markdown
## Question 1
What is the primary user authentication method?

A) Username and password
B) Social media login (Google, Facebook)
C) Single Sign-On (SSO)
D) Multi-factor authentication

[Answer]: C
```

「Other」を選んだ場合は説明を追記：

```markdown
[Answer]: X - OAuth 2.0 with custom identity provider
```

---

## 選択肢の品質ガイドライン

### 良い例
```markdown
## Question 5
What database technology will be used?

A) Relational (PostgreSQL, MySQL)
B) NoSQL Document (MongoDB, DynamoDB)
C) NoSQL Key-Value (Redis, Memcached)
D) Graph Database (Neo4j, Neptune)
E) Other (please describe after [Answer]: tag below)

[Answer]:
```

### 悪い例（避ける）
```markdown
## Question 5
What database will you use?

A) Yes
B) No
C) Maybe

[Answer]:
```

### 選択肢作成のポイント
- 相互排他的にする
- 最も一般的なシナリオをカバー
- 意味のある現実的な選択肢のみ
- **常に「Other」を最後に**（必須）
- 具体的で明確に

---

## 矛盾・曖昧さの検出

### 矛盾の例
- スコープ不一致: 「バグ修正」なのに「コードベース全体に影響」
- リスク不一致: 「低リスク」なのに「破壊的変更あり」
- タイムライン不一致: 「クイックフィックス」なのに「複数サブシステム」
- 影響不一致: 「単一コンポーネント」なのに「大幅なアーキテクチャ変更」

### 曖昧な回答の例
- "mix of A and B"
- "somewhere between"
- "not sure"
- "depends"
- "maybe"
- "probably"
- "standard" (未定義)
- "typical"

---

## 明確化質問ファイル

矛盾・曖昧さが検出された場合、明確化質問ファイルを作成：

```markdown
# [Phase Name] Clarification Questions

I detected contradictions in your responses that need clarification:

## Contradiction 1: [Brief Description]
You indicated "[Answer A]" (Q[X]:[Letter]) but also "[Answer B]" (Q[Y]:[Letter]).
These responses are contradictory because [explanation].

### Clarification Question 1
[Specific question to resolve contradiction]

A) [Option that resolves toward first answer]
B) [Option that resolves toward second answer]
C) [Option that provides middle ground]
D) Other (please describe after [Answer]: tag below)

[Answer]:

## Ambiguity 1: [Brief Description]
Your response to Q[X] ("[Answer]") is ambiguous because [explanation].

### Clarification Question 2
[Specific question to clarify ambiguity]

A) [Clear option 1]
B) [Clear option 2]
C) [Clear option 3]
D) Other (please describe after [Answer]: tag below)

[Answer]:
```

---

## ワークフロー統合

### Step 1: 質問ファイル作成
```
Create aidlc-docs/{phase-name}-questions.md with all questions
```

### Step 2: ユーザーに通知
```
"I've created {phase-name}-questions.md with [X] questions.
Please answer each question by filling in the letter choice after the [Answer]: tag.
If none of the options match your needs, choose the last option (Other) and describe your preference.
Let me know when you're done."
```

### Step 3: 完了確認を待つ
ユーザーが "done", "completed", "finished" などと言うまで待機。

### Step 4: 読み取り・分析
```
Read aidlc-docs/{phase-name}-questions.md
Extract all answers
Validate completeness
Check for contradictions/ambiguities
Proceed with analysis (or create clarification file)
```

---

## エラー処理

### 未回答
```
"I noticed Question [X] is not answered. Please provide an answer using one of the letter choices
for all questions before proceeding."
```

### 無効な回答
```
"Question [X] has an invalid answer '[answer]'.
Please use only the letter choices provided in the question."
```

### 曖昧な回答
```
"For Question [X], please provide the letter choice that best matches your answer.
If none match, choose 'Other' and add your description after the [Answer]: tag."
```

---

## サマリー

### DO（するべきこと）
- 常に質問ファイルを作成
- 常に選択式フォーマットを使用
- **常に「Other」を最後に含める**（必須）
- 常に[Answer]:タグを使用
- 常にユーザー完了を待つ
- 常に矛盾を検証
- 常に必要なら明確化ファイルを作成
- 常に矛盾を解消してから進む

### DON'T（してはいけないこと）
- チャットで質問しない
- A, B, C, Dを埋めるためだけに選択肢を作らない
- 回答なしで進まない
- 未解決の矛盾のまま進まない
- 曖昧な回答について仮定しない

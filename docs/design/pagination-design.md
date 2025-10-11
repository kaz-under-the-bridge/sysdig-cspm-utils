# Sysdig Compliance API ページネーション設計書

## 1. ページネーション方式

### 1.1 採用方式
**オフセットベースページネーション**を採用

#### 選定理由
- UIでのページジャンプが容易
- 総件数の把握が可能
- 実装がシンプル
- Sysdig APIの標準仕様に準拠

### 1.2 パラメータ仕様

| パラメータ | 型 | デフォルト値 | 最大値 | 説明 |
|-----------|-----|-------------|---------|------|
| `pageNumber` | integer | 1 | - | ページ番号（1始まり） |
| `pageSize` | integer | 10 | 50 | 1ページあたりの件数 |

## 2. レスポンス構造

### 2.1 基本構造
```json
{
  "data": [...],           // 実データ配列
  "totalCount": 732,       // 全体件数
  "pageNumber": 1,         // 現在のページ番号（拡張提案）
  "pageSize": 50,          // ページサイズ（拡張提案）
  "totalPages": 15,        // 総ページ数（拡張提案）
  "hasNext": true,         // 次ページ存在フラグ（拡張提案）
  "hasPrevious": false     // 前ページ存在フラグ（拡張提案）
}
```

### 2.2 メタデータの計算
```typescript
interface PaginationMeta {
  totalCount: number;
  pageNumber: number;
  pageSize: number;
  totalPages: number;
  hasNext: boolean;
  hasPrevious: boolean;
  startIndex: number;  // 開始インデックス（0始まり）
  endIndex: number;    // 終了インデックス
}

function calculatePaginationMeta(
  totalCount: number,
  pageNumber: number,
  pageSize: number
): PaginationMeta {
  const totalPages = Math.ceil(totalCount / pageSize);
  const startIndex = (pageNumber - 1) * pageSize;
  const endIndex = Math.min(startIndex + pageSize - 1, totalCount - 1);

  return {
    totalCount,
    pageNumber,
    pageSize,
    totalPages,
    hasNext: pageNumber < totalPages,
    hasPrevious: pageNumber > 1,
    startIndex,
    endIndex
  };
}
```

## 3. 実装パターン

### 3.1 基本的な取得パターン

```typescript
class ComplianceApiClient {
  private baseUrl: string;
  private token: string;

  async getComplianceResults(
    filter: string,
    pageNumber: number = 1,
    pageSize: number = 50
  ): Promise<PaginatedResponse<ComplianceRequirement>> {
    const params = new URLSearchParams({
      filter,
      pageNumber: pageNumber.toString(),
      pageSize: pageSize.toString()
    });

    const response = await fetch(
      `${this.baseUrl}?${params}`,
      {
        headers: {
          'Authorization': `Bearer ${this.token}`,
          'Content-Type': 'application/json'
        }
      }
    );

    return response.json();
  }
}
```

### 3.2 全件取得パターン

```typescript
class ComplianceFetcher {
  async fetchAllResults(
    client: ComplianceApiClient,
    filter: string,
    pageSize: number = 50
  ): Promise<ComplianceRequirement[]> {
    const allResults: ComplianceRequirement[] = [];
    let pageNumber = 1;
    let hasMore = true;

    while (hasMore) {
      const response = await client.getComplianceResults(
        filter,
        pageNumber,
        pageSize
      );

      allResults.push(...response.data);

      // 終了条件の判定
      const fetchedCount = pageNumber * pageSize;
      hasMore = fetchedCount < response.totalCount;
      pageNumber++;

      // レート制限対策
      await this.delay(100);
    }

    return allResults;
  }

  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}
```

### 3.3 並列取得パターン

```typescript
class ParallelFetcher {
  async fetchAllParallel(
    client: ComplianceApiClient,
    filter: string,
    pageSize: number = 50,
    maxConcurrency: number = 5
  ): Promise<ComplianceRequirement[]> {
    // 最初のリクエストで全体件数を取得
    const firstPage = await client.getComplianceResults(filter, 1, pageSize);
    const totalCount = firstPage.totalCount;
    const totalPages = Math.ceil(totalCount / pageSize);

    // ページ番号の配列を生成
    const pageNumbers = Array.from(
      { length: totalPages },
      (_, i) => i + 1
    );

    // バッチ処理
    const results: ComplianceRequirement[] = [...firstPage.data];

    for (let i = 1; i < pageNumbers.length; i += maxConcurrency) {
      const batch = pageNumbers.slice(i, i + maxConcurrency);
      const batchPromises = batch.map(pageNum =>
        client.getComplianceResults(filter, pageNum, pageSize)
      );

      const batchResults = await Promise.all(batchPromises);
      batchResults.forEach(response => {
        results.push(...response.data);
      });
    }

    return results;
  }
}
```

## 4. キャッシュ戦略

### 4.1 ページ単位のキャッシュ

```typescript
class PageCache {
  private cache: Map<string, CachedPage>;
  private ttlMs: number;

  constructor(ttlMinutes: number = 15) {
    this.cache = new Map();
    this.ttlMs = ttlMinutes * 60 * 1000;
  }

  generateKey(filter: string, pageNumber: number, pageSize: number): string {
    const filterHash = this.hashString(filter);
    return `${filterHash}:${pageNumber}:${pageSize}`;
  }

  get(key: string): PaginatedResponse<ComplianceRequirement> | null {
    const cached = this.cache.get(key);

    if (!cached) return null;

    if (Date.now() > cached.expiresAt) {
      this.cache.delete(key);
      return null;
    }

    return cached.data;
  }

  set(key: string, data: PaginatedResponse<ComplianceRequirement>): void {
    this.cache.set(key, {
      data,
      expiresAt: Date.now() + this.ttlMs
    });
  }

  private hashString(str: string): string {
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
      const char = str.charCodeAt(i);
      hash = ((hash << 5) - hash) + char;
      hash = hash & hash;
    }
    return hash.toString(36);
  }
}

interface CachedPage {
  data: PaginatedResponse<ComplianceRequirement>;
  expiresAt: number;
}
```

### 4.2 インテリジェントプリフェッチ

```typescript
class SmartPaginator {
  private cache: PageCache;
  private prefetchQueue: Set<string>;

  async getPageWithPrefetch(
    client: ComplianceApiClient,
    filter: string,
    pageNumber: number,
    pageSize: number
  ): Promise<PaginatedResponse<ComplianceRequirement>> {
    const key = this.cache.generateKey(filter, pageNumber, pageSize);

    // キャッシュチェック
    let result = this.cache.get(key);

    if (!result) {
      result = await client.getComplianceResults(filter, pageNumber, pageSize);
      this.cache.set(key, result);
    }

    // 次ページのプリフェッチ
    this.prefetchNextPages(client, filter, pageNumber, pageSize, result.totalCount);

    return result;
  }

  private async prefetchNextPages(
    client: ComplianceApiClient,
    filter: string,
    currentPage: number,
    pageSize: number,
    totalCount: number
  ): Promise<void> {
    const totalPages = Math.ceil(totalCount / pageSize);
    const pagesToPrefetch: number[] = [];

    // 次の2ページをプリフェッチ
    if (currentPage < totalPages) {
      pagesToPrefetch.push(currentPage + 1);
    }
    if (currentPage + 1 < totalPages) {
      pagesToPrefetch.push(currentPage + 2);
    }

    // 非同期でプリフェッチ
    pagesToPrefetch.forEach(pageNum => {
      const key = this.cache.generateKey(filter, pageNum, pageSize);
      if (!this.prefetchQueue.has(key)) {
        this.prefetchQueue.add(key);
        this.doPrefetch(client, filter, pageNum, pageSize, key);
      }
    });
  }

  private async doPrefetch(
    client: ComplianceApiClient,
    filter: string,
    pageNumber: number,
    pageSize: number,
    key: string
  ): Promise<void> {
    try {
      const result = await client.getComplianceResults(filter, pageNumber, pageSize);
      this.cache.set(key, result);
    } finally {
      this.prefetchQueue.delete(key);
    }
  }
}
```

## 5. UI実装ガイドライン

### 5.1 ページネーションコンポーネント

```typescript
interface PaginationProps {
  currentPage: number;
  totalPages: number;
  pageSize: number;
  totalCount: number;
  onPageChange: (page: number) => void;
  onPageSizeChange: (size: number) => void;
}

const PaginationComponent: React.FC<PaginationProps> = ({
  currentPage,
  totalPages,
  pageSize,
  totalCount,
  onPageChange,
  onPageSizeChange
}) => {
  const startItem = (currentPage - 1) * pageSize + 1;
  const endItem = Math.min(currentPage * pageSize, totalCount);

  return (
    <div className="pagination">
      <div className="pagination-info">
        表示中: {startItem}-{endItem} / 全{totalCount}件
      </div>

      <div className="pagination-controls">
        <button
          disabled={currentPage === 1}
          onClick={() => onPageChange(1)}
        >
          最初
        </button>

        <button
          disabled={currentPage === 1}
          onClick={() => onPageChange(currentPage - 1)}
        >
          前へ
        </button>

        <span>
          ページ {currentPage} / {totalPages}
        </span>

        <button
          disabled={currentPage === totalPages}
          onClick={() => onPageChange(currentPage + 1)}
        >
          次へ
        </button>

        <button
          disabled={currentPage === totalPages}
          onClick={() => onPageChange(totalPages)}
        >
          最後
        </button>
      </div>

      <div className="page-size-selector">
        <label>表示件数:</label>
        <select
          value={pageSize}
          onChange={(e) => onPageSizeChange(Number(e.target.value))}
        >
          <option value="10">10件</option>
          <option value="25">25件</option>
          <option value="50">50件</option>
        </select>
      </div>
    </div>
  );
};
```

### 5.2 無限スクロール実装

```typescript
class InfiniteScrollLoader {
  private observer: IntersectionObserver;
  private isLoading: boolean = false;
  private hasMore: boolean = true;

  constructor(
    private container: HTMLElement,
    private loadMore: () => Promise<void>
  ) {
    this.setupObserver();
  }

  private setupObserver(): void {
    const options = {
      root: this.container,
      rootMargin: '100px',
      threshold: 0.1
    };

    this.observer = new IntersectionObserver(
      this.handleIntersection.bind(this),
      options
    );
  }

  private async handleIntersection(
    entries: IntersectionObserverEntry[]
  ): Promise<void> {
    const [entry] = entries;

    if (entry.isIntersecting && !this.isLoading && this.hasMore) {
      this.isLoading = true;
      try {
        await this.loadMore();
      } finally {
        this.isLoading = false;
      }
    }
  }

  observe(element: HTMLElement): void {
    this.observer.observe(element);
  }

  unobserve(element: HTMLElement): void {
    this.observer.unobserve(element);
  }

  setHasMore(hasMore: boolean): void {
    this.hasMore = hasMore;
  }

  disconnect(): void {
    this.observer.disconnect();
  }
}
```

## 6. パフォーマンス最適化

### 6.1 推奨設定

| 用途 | 推奨pageSize | 説明 |
|------|-------------|------|
| ダッシュボード表示 | 10-25 | レスポンス速度重視 |
| データテーブル | 25-50 | バランス重視 |
| バッチ処理 | 50 | スループット重視 |
| エクスポート | 50 | 最大効率 |

### 6.2 レート制限対策

```typescript
class RateLimitManager {
  private requestQueue: (() => Promise<any>)[] = [];
  private activeRequests: number = 0;
  private readonly maxConcurrent: number = 5;
  private readonly minDelay: number = 100;

  async execute<T>(fn: () => Promise<T>): Promise<T> {
    return new Promise((resolve, reject) => {
      this.requestQueue.push(async () => {
        try {
          const result = await fn();
          resolve(result);
        } catch (error) {
          reject(error);
        }
      });

      this.processQueue();
    });
  }

  private async processQueue(): Promise<void> {
    if (this.activeRequests >= this.maxConcurrent || this.requestQueue.length === 0) {
      return;
    }

    this.activeRequests++;
    const request = this.requestQueue.shift();

    if (request) {
      await request();
      await this.delay(this.minDelay);
      this.activeRequests--;
      this.processQueue();
    }
  }

  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}
```

## 7. エラーハンドリング

### 7.1 ページネーション固有のエラー

```typescript
enum PaginationError {
  PAGE_OUT_OF_RANGE = 'PAGE_OUT_OF_RANGE',
  INVALID_PAGE_SIZE = 'INVALID_PAGE_SIZE',
  CONCURRENT_MODIFICATION = 'CONCURRENT_MODIFICATION'
}

class PaginationValidator {
  static validatePageNumber(
    pageNumber: number,
    totalPages: number
  ): void {
    if (pageNumber < 1 || pageNumber > totalPages) {
      throw new Error(
        `Page number ${pageNumber} is out of range [1, ${totalPages}]`
      );
    }
  }

  static validatePageSize(pageSize: number): void {
    if (pageSize < 1 || pageSize > 50) {
      throw new Error(
        `Page size ${pageSize} is invalid. Must be between 1 and 50`
      );
    }
  }
}
```

### 7.2 リトライ戦略

```typescript
class PaginationRetryStrategy {
  async fetchWithRetry(
    fetchFn: () => Promise<any>,
    maxRetries: number = 3
  ): Promise<any> {
    let lastError: Error | null = null;

    for (let attempt = 0; attempt < maxRetries; attempt++) {
      try {
        return await fetchFn();
      } catch (error) {
        lastError = error as Error;

        // ページ範囲外エラーの場合はリトライしない
        if (this.isPageOutOfRangeError(error)) {
          throw error;
        }

        // 指数バックオフ
        const delay = Math.min(1000 * Math.pow(2, attempt), 10000);
        await this.delay(delay);
      }
    }

    throw lastError;
  }

  private isPageOutOfRangeError(error: any): boolean {
    return error?.message?.includes('out of range') ||
           error?.code === 'PAGE_OUT_OF_RANGE';
  }

  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}
```

## 8. テスト戦略

### 8.1 ユニットテスト例

```typescript
describe('Pagination', () => {
  describe('calculatePaginationMeta', () => {
    it('should calculate correct metadata for first page', () => {
      const meta = calculatePaginationMeta(100, 1, 10);
      expect(meta.hasNext).toBe(true);
      expect(meta.hasPrevious).toBe(false);
      expect(meta.totalPages).toBe(10);
      expect(meta.startIndex).toBe(0);
      expect(meta.endIndex).toBe(9);
    });

    it('should handle last page correctly', () => {
      const meta = calculatePaginationMeta(95, 10, 10);
      expect(meta.hasNext).toBe(false);
      expect(meta.hasPrevious).toBe(true);
      expect(meta.endIndex).toBe(94);
    });
  });
});
```
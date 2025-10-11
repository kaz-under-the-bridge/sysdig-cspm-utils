package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/kaz-under-the-bridge/sysdig-cspm-utils/pkg/models"
	"github.com/kaz-under-the-bridge/sysdig-cspm-utils/pkg/sysdig"
)

// CSPMClient wraps the base Sysdig client for CSPM-specific operations
type CSPMClient struct {
	*sysdig.Client
}

// NewCSPMClient creates a new CSPM client
func NewCSPMClient(apiURL, apiToken string) *CSPMClient {
	return &CSPMClient{
		Client: sysdig.NewClient(apiURL, apiToken),
	}
}

// GetComplianceRequirements retrieves compliance requirements with violations
func (c *CSPMClient) GetComplianceRequirements(filter string) (*models.ComplianceResponse, error) {
	return c.GetComplianceRequirementsPaginated(filter, 0, 0)
}

// GetComplianceRequirementsPaginated retrieves compliance requirements with pagination support
func (c *CSPMClient) GetComplianceRequirementsPaginated(filter string, pageNumber, pageSize int) (*models.ComplianceResponse, error) {
	endpoint := "/api/cspm/v1/compliance/requirements"

	// Build query parameters
	params := url.Values{}
	if filter != "" {
		params.Set("filter", filter)
	}
	if pageNumber > 0 {
		params.Set("pageNumber", strconv.Itoa(pageNumber))
	}
	if pageSize > 0 {
		params.Set("pageSize", strconv.Itoa(pageSize))
	}

	fullURL := endpoint
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

	resp, err := c.Client.MakeRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get compliance requirements: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var response models.ComplianceResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to parse compliance response: %w", err)
	}

	return &response, nil
}

// GetAllComplianceRequirements retrieves all compliance requirements by iterating through all pages with parallel processing
func (c *CSPMClient) GetAllComplianceRequirements(filter string, pageSize, batchSize, apiDelay int) (*models.ComplianceResponse, error) {
	if pageSize <= 0 {
		pageSize = 50 // デフォルト値
	}
	if batchSize <= 0 {
		batchSize = 3 // デフォルト値
	}
	if apiDelay < 0 {
		apiDelay = 1 // デフォルト値
	}

	// 最初のページを取得してtotalCountを確認
	firstResponse, err := c.GetComplianceRequirementsPaginated(filter, 1, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get first page: %w", err)
	}

	totalCount := firstResponse.TotalCount
	totalPages := (totalCount.Int() + pageSize - 1) / pageSize

	fmt.Printf("  Total pages: %d (pageSize: %d)\n", totalPages, pageSize)

	// 全データを格納するスライス
	allData := make([]models.ComplianceRequirement, 0, totalCount.Int())
	allData = append(allData, firstResponse.Data...)

	if totalPages <= 1 {
		fmt.Println()
		return &models.ComplianceResponse{
			Data:       allData,
			TotalCount: totalCount,
		}, nil
	}

	// 並列処理用の変数
	var mu sync.Mutex
	var wg sync.WaitGroup
	errors := make([]error, 0)
	pageResults := make(map[int][]models.ComplianceRequirement)

	// ページ2以降をバッチで並列処理
	for i := 2; i <= totalPages; i += batchSize {
		end := i + batchSize
		if end > totalPages {
			end = totalPages + 1 // +1 because j < end (exclusive upper bound)
		}

		for j := i; j < end && j <= totalPages; j++ {
			wg.Add(1)
			go func(pageNum int) {
				defer wg.Done()

				response, err := c.GetComplianceRequirementsPaginated(filter, pageNum, pageSize)
				if err != nil {
					mu.Lock()
					errors = append(errors, fmt.Errorf("failed to get page %d: %w", pageNum, err))
					mu.Unlock()
					return
				}

				mu.Lock()
				pageResults[pageNum] = response.Data
				currentTotal := len(allData)
				for _, data := range pageResults {
					currentTotal += len(data)
				}
				pagesProcessed := len(pageResults) + 1 // +1 for first page
				fmt.Printf("\r  Progress: %d/%d pages processed, %d compliance requirements collected", pagesProcessed, totalPages, currentTotal)
				mu.Unlock()
			}(j)
		}
		wg.Wait()

		// エラーチェック
		if len(errors) > 0 {
			fmt.Printf("\n[ERROR] Encountered %d errors during compliance requirements batch processing\n", len(errors))
			for i, err := range errors {
				fmt.Printf("[ERROR %d] %v\n", i+1, err)
			}
			return nil, errors[0]
		}

		// バッチ間の遅延
		if i+batchSize <= totalPages {
			time.Sleep(time.Duration(apiDelay) * time.Second)
		}
	}

	// ページ順にデータを結合
	for page := 2; page <= totalPages; page++ {
		if data, ok := pageResults[page]; ok {
			allData = append(allData, data...)
		}
	}

	fmt.Println() // 改行

	return &models.ComplianceResponse{
		Data:       allData,
		TotalCount: totalCount,
	}, nil
}

// GetInventoryResources retrieves inventory resources using specified filter (backward compatible)
func (c *CSPMClient) GetInventoryResources(filter string, limit int) (*models.InventoryResponse, error) {
	return c.GetInventoryResourcesPaginated(filter, 0, limit)
}

// GetInventoryResourcesPaginated retrieves inventory resources with pagination support
func (c *CSPMClient) GetInventoryResourcesPaginated(filter string, pageNumber, pageSize int) (*models.InventoryResponse, error) {
	endpoint := "/api/cspm/v1/inventory/resources"

	// Build query parameters
	params := url.Values{}
	if filter != "" {
		params.Set("filter", filter)
	}
	if pageNumber > 0 {
		params.Set("pageNumber", strconv.Itoa(pageNumber))
	}
	if pageSize > 0 {
		params.Set("limit", strconv.Itoa(pageSize))
	}

	fullURL := endpoint
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

	resp, err := c.Client.MakeRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory resources: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// エラーレスポンスボディを読み取る
		var errResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			if msg, ok := errResp["message"]; ok {
				return nil, fmt.Errorf("API request failed with status %d: %v (filter: %s)", resp.StatusCode, msg, filter)
			}
		}
		return nil, fmt.Errorf("API request failed with status %d (filter: %s)", resp.StatusCode, filter)
	}

	var response models.InventoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to parse inventory response: %w", err)
	}

	return &response, nil
}

// GetAllInventoryResources retrieves all inventory resources by iterating through all pages with parallel processing
func (c *CSPMClient) GetAllInventoryResources(filter string, pageSize, batchSize, apiDelay int) (*models.InventoryResponse, error) {
	if pageSize <= 0 {
		pageSize = 50 // デフォルト値
	}
	if batchSize <= 0 {
		batchSize = 3 // デフォルト値
	}
	if apiDelay < 0 {
		apiDelay = 1 // デフォルト値
	}

	// 最初のページを取得してtotalCountを確認
	firstResponse, err := c.GetInventoryResourcesPaginated(filter, 1, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get first page: %w", err)
	}

	totalCount := firstResponse.TotalCount
	totalPages := (totalCount.Int() + pageSize - 1) / pageSize

	fmt.Printf("  Total pages: %d (pageSize: %d)\n", totalPages, pageSize)

	// 全データを格納するスライス
	allData := make([]models.InventoryResource, 0, totalCount.Int())
	allData = append(allData, firstResponse.Data...)

	if totalPages <= 1 {
		fmt.Println()
		return &models.InventoryResponse{
			Data:       allData,
			TotalCount: totalCount,
		}, nil
	}

	// 並列処理用の変数
	var mu sync.Mutex
	var wg sync.WaitGroup
	errors := make([]error, 0)
	pageResults := make(map[int][]models.InventoryResource)

	// ページ2以降をバッチで並列処理
	for i := 2; i <= totalPages; i += batchSize {
		end := i + batchSize
		if end > totalPages {
			end = totalPages + 1 // +1 because j < end (exclusive upper bound)
		}

		for j := i; j < end && j <= totalPages; j++ {
			wg.Add(1)
			go func(pageNum int) {
				defer wg.Done()

				response, err := c.GetInventoryResourcesPaginated(filter, pageNum, pageSize)
				if err != nil {
					mu.Lock()
					errors = append(errors, fmt.Errorf("failed to get page %d: %w", pageNum, err))
					mu.Unlock()
					return
				}

				mu.Lock()
				pageResults[pageNum] = response.Data
				if len(response.Data) == 0 {
					fmt.Printf("\n[WARN] Page %d returned 0 resources\n", pageNum)
				}
				currentTotal := len(allData)
				for _, data := range pageResults {
					currentTotal += len(data)
				}
				pagesProcessed := len(pageResults) + 1 // +1 for first page
				fmt.Printf("\r  Progress: %d/%d pages processed, %d resources collected (page %d: %d items)", pagesProcessed, totalPages, currentTotal, pageNum, len(response.Data))
				mu.Unlock()
			}(j)
		}
		wg.Wait()

		// エラーチェック
		if len(errors) > 0 {
			fmt.Printf("\n[ERROR] Encountered %d errors during inventory resource batch processing\n", len(errors))
			for i, err := range errors {
				fmt.Printf("[ERROR %d] %v\n", i+1, err)
			}
			return nil, errors[0]
		}

		// バッチ間の遅延
		if i+batchSize <= totalPages {
			time.Sleep(time.Duration(apiDelay) * time.Second)
		}
	}

	// ページ順にデータを結合
	for page := 2; page <= totalPages; page++ {
		if data, ok := pageResults[page]; ok {
			allData = append(allData, data...)
		}
	}

	fmt.Println() // 改行

	return &models.InventoryResponse{
		Data:       allData,
		TotalCount: totalCount,
	}, nil
}

// GetComplianceViolations retrieves compliance violations for a specific policy and zone
func (c *CSPMClient) GetComplianceViolations(policyName, zoneName string) (*models.ComplianceResponse, error) {
	filter := fmt.Sprintf(`pass = "false" and policy.name in ("%s") and zone.name in ("%s")`,
		policyName, zoneName)
	return c.GetComplianceRequirements(filter)
}

// GetResourcesForControl retrieves resources associated with a specific failed control
func (c *CSPMClient) GetResourcesForControl(controlName string, limit int) (*models.InventoryResponse, error) {
	filter := fmt.Sprintf(`control.failed in ("%s")`, controlName)
	return c.GetInventoryResources(filter, limit)
}

// GetComplianceRequirementsWithControls retrieves compliance requirements with controls included
func (c *CSPMClient) GetComplianceRequirementsWithControls(filter string, pageNumber, pageSize int) (*models.ComplianceResponseWithControls, error) {
	endpoint := "/api/cspm/v1/compliance/requirements"

	// Build query parameters
	params := url.Values{}
	if filter != "" {
		params.Set("filter", filter)
	}
	if pageNumber > 0 {
		params.Set("pageNumber", strconv.Itoa(pageNumber))
	}
	if pageSize > 0 {
		params.Set("pageSize", strconv.Itoa(pageSize))
	}

	fullURL := endpoint
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

	resp, err := c.Client.MakeRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get compliance requirements: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var response models.ComplianceResponseWithControls
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to parse compliance response: %w", err)
	}

	return &response, nil
}

// GetAllComplianceRequirementsWithControls retrieves all compliance requirements with controls by iterating through all pages
func (c *CSPMClient) GetAllComplianceRequirementsWithControls(filter string, pageSize, batchSize, apiDelay int) (*models.ComplianceResponseWithControls, error) {
	if pageSize <= 0 {
		pageSize = 50 // デフォルト値
	}
	if batchSize <= 0 {
		batchSize = 3 // デフォルト値
	}
	if apiDelay < 0 {
		apiDelay = 1 // デフォルト値
	}

	// 最初のページを取得してtotalCountを確認
	firstResponse, err := c.GetComplianceRequirementsWithControls(filter, 1, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get first page: %w", err)
	}

	totalCount := firstResponse.TotalCount
	totalPages := (totalCount.Int() + pageSize - 1) / pageSize

	fmt.Printf("  Total pages: %d (pageSize: %d)\n", totalPages, pageSize)

	// 全データを格納するスライス
	allData := make([]models.ComplianceRequirementWithControls, 0, totalCount.Int())
	allData = append(allData, firstResponse.Data...)

	if totalPages <= 1 {
		fmt.Println()
		return &models.ComplianceResponseWithControls{
			Data:       allData,
			TotalCount: totalCount,
		}, nil
	}

	// 並列処理用の変数
	var mu sync.Mutex
	var wg sync.WaitGroup
	errors := make([]error, 0)
	pageResults := make(map[int][]models.ComplianceRequirementWithControls)

	// ページ2以降をバッチで並列処理
	for i := 2; i <= totalPages; i += batchSize {
		end := i + batchSize
		if end > totalPages {
			end = totalPages + 1
		}

		for j := i; j < end && j <= totalPages; j++ {
			wg.Add(1)
			go func(pageNum int) {
				defer wg.Done()

				response, err := c.GetComplianceRequirementsWithControls(filter, pageNum, pageSize)
				if err != nil {
					mu.Lock()
					errors = append(errors, fmt.Errorf("failed to get page %d: %w", pageNum, err))
					mu.Unlock()
					return
				}

				mu.Lock()
				pageResults[pageNum] = response.Data
				currentTotal := len(allData)
				for _, data := range pageResults {
					currentTotal += len(data)
				}
				pagesProcessed := len(pageResults) + 1
				fmt.Printf("\r  Progress: %d/%d pages processed, %d compliance requirements collected", pagesProcessed, totalPages, currentTotal)
				mu.Unlock()
			}(j)
		}
		wg.Wait()

		// エラーチェック
		if len(errors) > 0 {
			fmt.Printf("\n[ERROR] Encountered %d errors during compliance requirements batch processing\n", len(errors))
			for i, err := range errors {
				fmt.Printf("[ERROR %d] %v\n", i+1, err)
			}
			return nil, errors[0]
		}

		// バッチ間の遅延
		if i+batchSize <= totalPages {
			time.Sleep(time.Duration(apiDelay) * time.Second)
		}
	}

	// ページ順にデータを結合
	for page := 2; page <= totalPages; page++ {
		if data, ok := pageResults[page]; ok {
			allData = append(allData, data...)
		}
	}

	fmt.Println() // 改行

	return &models.ComplianceResponseWithControls{
		Data:       allData,
		TotalCount: totalCount,
	}, nil
}

// GetCloudResources retrieves cloud resources using the resource API endpoint
func (c *CSPMClient) GetCloudResources(endpoint string, pageNumber, pageSize int) (*models.CloudResourceResponse, error) {
	// Parse the endpoint URL to properly handle query parameters
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse endpoint URL: %w", err)
	}

	// Get existing query parameters
	queryParams := parsedURL.Query()

	// Add pagination parameters
	if pageNumber > 0 {
		queryParams.Set("pageNumber", strconv.Itoa(pageNumber))
	}
	if pageSize > 0 {
		queryParams.Set("pageSize", strconv.Itoa(pageSize))
	}

	// Rebuild the URL with properly encoded query parameters
	parsedURL.RawQuery = queryParams.Encode()
	fullURL := parsedURL.String()

	resp, err := c.Client.MakeRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get cloud resources: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		var errResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			if msg, ok := errResp["message"]; ok {
				return nil, fmt.Errorf("API request failed with status %d: %v", resp.StatusCode, msg)
			}
		}
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var response models.CloudResourceResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to parse cloud resource response: %w", err)
	}

	return &response, nil
}

// GetAllCloudResources retrieves all cloud resources by iterating through all pages
func (c *CSPMClient) GetAllCloudResources(endpoint string, pageSize, batchSize, apiDelay int) (*models.CloudResourceResponse, error) {
	if pageSize <= 0 {
		pageSize = 50
	}
	if batchSize <= 0 {
		batchSize = 3
	}
	if apiDelay < 0 {
		apiDelay = 1
	}

	// 最初のページを取得してtotalCountを確認
	firstResponse, err := c.GetCloudResources(endpoint, 1, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get first page: %w", err)
	}

	totalCount := firstResponse.TotalCount
	totalPages := (totalCount.Int() + pageSize - 1) / pageSize

	// 全データを格納するスライス
	allData := make([]models.CloudResource, 0, totalCount.Int())
	allData = append(allData, firstResponse.Data...)

	if totalPages <= 1 {
		return &models.CloudResourceResponse{
			Data:       allData,
			TotalCount: totalCount,
		}, nil
	}

	// 並列処理用の変数
	var mu sync.Mutex
	var wg sync.WaitGroup
	errors := make([]error, 0)
	pageResults := make(map[int][]models.CloudResource)

	// ページ2以降をバッチで並列処理
	for i := 2; i <= totalPages; i += batchSize {
		end := i + batchSize
		if end > totalPages {
			end = totalPages + 1
		}

		for j := i; j < end && j <= totalPages; j++ {
			wg.Add(1)
			go func(pageNum int) {
				defer wg.Done()

				response, err := c.GetCloudResources(endpoint, pageNum, pageSize)
				if err != nil {
					mu.Lock()
					errors = append(errors, fmt.Errorf("failed to get page %d: %w", pageNum, err))
					mu.Unlock()
					return
				}

				mu.Lock()
				pageResults[pageNum] = response.Data
				mu.Unlock()
			}(j)
		}
		wg.Wait()

		// エラーチェック
		if len(errors) > 0 {
			return nil, errors[0]
		}

		// バッチ間の遅延
		if i+batchSize <= totalPages {
			time.Sleep(time.Duration(apiDelay) * time.Second)
		}
	}

	// ページ順にデータを結合
	for page := 2; page <= totalPages; page++ {
		if data, ok := pageResults[page]; ok {
			allData = append(allData, data...)
		}
	}

	return &models.CloudResourceResponse{
		Data:       allData,
		TotalCount: totalCount,
	}, nil
}

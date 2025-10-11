package collector

import (
	"fmt"
	"time"

	"github.com/kaz-under-the-bridge/sysdig-cspm-utils/pkg/client"
	"github.com/kaz-under-the-bridge/sysdig-cspm-utils/pkg/database"
	"github.com/kaz-under-the-bridge/sysdig-cspm-utils/pkg/models"
)

// ComplianceCollector collects compliance requirements and associated resources
type ComplianceCollector struct {
	client *client.CSPMClient
	db     *database.Database
}

// NewComplianceCollector creates a new ComplianceCollector
func NewComplianceCollector(cspmClient *client.CSPMClient, db *database.Database) *ComplianceCollector {
	return &ComplianceCollector{
		client: cspmClient,
		db:     db,
	}
}

// CollectComplianceData collects compliance requirements and associated resources
func (cc *ComplianceCollector) CollectComplianceData(policyFilter string, pageSize, batchSize, apiDelay int) error {
	// Step 1: Get compliance requirements with controls
	fmt.Println("Step 1: Getting compliance requirements with controls...")
	complianceResp, err := cc.client.GetAllComplianceRequirementsWithControls(policyFilter, pageSize, batchSize, apiDelay)
	if err != nil {
		return fmt.Errorf("failed to get compliance requirements: %w", err)
	}

	fmt.Printf("  Retrieved %d requirements\n", len(complianceResp.Data))

	// Save requirements and controls to DB
	if err := cc.db.SaveComplianceRequirementsWithControls(complianceResp.Data); err != nil {
		return fmt.Errorf("failed to save compliance requirements: %w", err)
	}

	// Step 2: Get resources for each control
	fmt.Println("\nStep 2: Getting resources for each control...")
	totalControls := 0
	totalResources := 0
	failedControlsCount := 0

	for reqIdx, req := range complianceResp.Data {
		fmt.Printf("\n[%d/%d] Processing requirement: %s\n", reqIdx+1, len(complianceResp.Data), req.Name)

		if req.Pass {
			fmt.Println("  Skipping (passed requirement)")
			continue
		}

		for ctrlIdx, ctrl := range req.Controls {
			totalControls++
			fmt.Printf("  [%d/%d] Control %s: %s\n", ctrlIdx+1, len(req.Controls), ctrl.ID, ctrl.Name)

			// Skip controls with no resourceApiEndpoint
			if ctrl.ResourceAPIEndpoint == "" {
				fmt.Println("    No resourceApiEndpoint, skipping")
				continue
			}

			// Get resources for this control
			startTime := time.Now()
			resources, err := cc.client.GetAllCloudResources(ctrl.ResourceAPIEndpoint, pageSize, batchSize, apiDelay)
			if err != nil {
				fmt.Printf("    [WARN] Failed to get resources: %v\n", err)
				failedControlsCount++
				continue
			}

			duration := time.Since(startTime)

			// Classify resources by status
			var failed, passed, accepted int
			for _, res := range resources.Data {
				if res.Acceptance != nil {
					accepted++
				} else if !res.Passed {
					failed++
				} else {
					passed++
				}
			}

			fmt.Printf("    Retrieved %d resources in %s (Failed: %d, Passed: %d, Accepted: %d)\n",
				len(resources.Data), duration.Round(time.Millisecond), failed, passed, accepted)

			// Save resources to DB
			if len(resources.Data) > 0 {
				if err := cc.db.SaveCloudResources(resources.Data); err != nil {
					return fmt.Errorf("failed to save resources: %w", err)
				}

				// Save control-resource relations
				if err := cc.db.SaveControlResourceRelations(ctrl.ID, resources.Data); err != nil {
					return fmt.Errorf("failed to save control-resource relations: %w", err)
				}
			}

			totalResources += len(resources.Data)

			// Delay between API calls to avoid rate limiting
			if apiDelay > 0 {
				time.Sleep(time.Duration(apiDelay) * time.Second)
			}
		}
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Total requirements: %d\n", len(complianceResp.Data))
	fmt.Printf("Total controls processed: %d\n", totalControls)
	fmt.Printf("Total resources collected: %d\n", totalResources)
	if failedControlsCount > 0 {
		fmt.Printf("Failed controls (warnings): %d\n", failedControlsCount)
	}

	return nil
}

// CollectComplianceDataWithStats collects compliance data and returns statistics
func (cc *ComplianceCollector) CollectComplianceDataWithStats(policyFilter string, pageSize, batchSize, apiDelay int) (*database.ComplianceStats, error) {
	if err := cc.CollectComplianceData(policyFilter, pageSize, batchSize, apiDelay); err != nil {
		return nil, err
	}

	stats, err := cc.db.GetComplianceStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get compliance stats: %w", err)
	}

	return stats, nil
}

// GetComplianceStats returns current compliance statistics from the database
func (cc *ComplianceCollector) GetComplianceStats() (*database.ComplianceStats, error) {
	return cc.db.GetComplianceStats()
}

// CollectControlResources collects resources for a specific control
func (cc *ComplianceCollector) CollectControlResources(controlID string, endpoint string, pageSize, batchSize, apiDelay int) ([]models.CloudResource, error) {
	fmt.Printf("Collecting resources for control %s...\n", controlID)

	resources, err := cc.client.GetAllCloudResources(endpoint, pageSize, batchSize, apiDelay)
	if err != nil {
		return nil, fmt.Errorf("failed to get resources: %w", err)
	}

	fmt.Printf("Retrieved %d resources\n", len(resources.Data))

	// Save resources to DB
	if len(resources.Data) > 0 {
		if err := cc.db.SaveCloudResources(resources.Data); err != nil {
			return nil, fmt.Errorf("failed to save resources: %w", err)
		}

		// Save control-resource relations
		if err := cc.db.SaveControlResourceRelations(controlID, resources.Data); err != nil {
			return nil, fmt.Errorf("failed to save control-resource relations: %w", err)
		}
	}

	return resources.Data, nil
}

package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/kaz-under-the-bridge/sysdig-cspm-utils/pkg/client"
	"github.com/kaz-under-the-bridge/sysdig-cspm-utils/pkg/config"
	"github.com/kaz-under-the-bridge/sysdig-cspm-utils/pkg/database"
)

const version = "1.0.0"

func main() {
	var (
		configFile     = flag.String("config", "", "Path to configuration file")
		apiToken       = flag.String("token", "", "Sysdig API token")
		apiURL         = flag.String("url", "https://us2.app.sysdig.com", "Sysdig API base URL")
		command        = flag.String("command", "list", "Command to execute: list, collect")
		dbPath         = flag.String("db", "data/cspm.db", "SQLite database path")
		policyType     = flag.String("policy", "", "Filter by policy name (comma-separated for multiple, partial match)")
		platform       = flag.String("platform", "", "Filter by platform (AWS, GCP, Azure, Kubernetes)")
		zoneName       = flag.String("zone", "Entire Infrastructure", "Filter by zone name")
		includePass    = flag.Bool("include-pass", false, "Include passed compliance checks (default: failed only)")
		batchSize      = flag.Int("batch-size", 3, "Number of concurrent API requests for pagination (default 3)")
		apiDelay       = flag.Int("api-delay", 1, "Delay in seconds between API batches (default 1)")
		showHelp       = flag.Bool("help", false, "Show help")
		showVersion    = flag.Bool("version", false, "Show version")
	)
	flag.Parse()

	if *showVersion {
		fmt.Printf("sysdig-cspm-utils version %s\n", version)
		return
	}

	if *showHelp {
		printUsage()
		return
	}

	// Load configuration
	cfg, err := config.Load(*configFile, *apiToken, *apiURL)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if cfg.APIToken == "" {
		log.Fatal("API token is required. Set via -token flag or SYSDIG_API_TOKEN environment variable")
	}

	// Create CSPM client
	cspmClient := client.NewCSPMClient(cfg.APIURL, cfg.APIToken)

	// Execute command
	switch *command {
	case "list":
		err = listCompliance(cspmClient, *policyType, *platform, *zoneName, *includePass)
	case "collect":
		err = collectResources(cspmClient, *dbPath, *policyType, *platform, *zoneName, *includePass, *batchSize, *apiDelay)
	default:
		log.Fatalf("Unknown command: %s", *command)
	}

	if err != nil {
		log.Fatalf("Command failed: %v", err)
	}
}

// buildFilter constructs a Sysdig CSPM API filter string from CLI parameters
func buildFilter(policies, platform, zoneName string, includePass bool) string {
	conditions := []string{}

	// passフィルター（include-passフラグがfalseの場合のみ追加）
	if !includePass {
		conditions = append(conditions, `pass = "false"`)
	}

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

func printUsage() {
	fmt.Printf(`sysdig-cspm-utils version %s

A tool for collecting and analyzing Sysdig CSPM compliance violations
and associated resources across multi-cloud environments.

Usage:
  sysdig-cspm-utils [options]

Options:
  -config string
        Path to configuration file
  -token string
        Sysdig API token (or set SYSDIG_API_TOKEN environment variable)
  -url string
        Sysdig API base URL (default "https://us2.app.sysdig.com")
  -command string
        Command to execute: list, collect (default "list")
  -db string
        SQLite database path (default "data/cspm.db")
  -policy string
        Filter by policy name (comma-separated for multiple, partial match)
        Examples: "CIS AWS", "SOC 2", "CIS AWS,CIS GCP,SOC 2"
  -platform string
        Filter by platform: AWS, GCP, Azure, Kubernetes
  -zone string
        Filter by zone name (default "Entire Infrastructure")
  -help
        Show this help message
  -version
        Show version information

Commands:
  list    - List compliance requirements with violations
  collect - Collect compliance violations and associated resources to database

Examples:
  # List all compliance violations
  sysdig-cspm-utils -token YOUR_TOKEN -command list

  # Filter by specific policy (partial match)
  sysdig-cspm-utils -token YOUR_TOKEN -policy "CIS AWS"

  # Filter by multiple policies (comma-separated)
  sysdig-cspm-utils -token YOUR_TOKEN -policy "CIS AWS,SOC 2,CIS GCP"

  # Filter by platform
  sysdig-cspm-utils -token YOUR_TOKEN -platform "AWS"

  # Collect violations with full policy name
  sysdig-cspm-utils -token YOUR_TOKEN -command collect \
    -policy "CIS Amazon Web Services Foundations Benchmark v3.0.0" \
    -zone "Entire Infrastructure" \
    -db "data/cis_aws.db"

  # Collect multiple policies for specific platform
  sysdig-cspm-utils -token YOUR_TOKEN -command collect \
    -policy "CIS AWS,SOC 2" \
    -platform "AWS" \
    -db "data/aws_compliance.db"

Environment Variables:
  SYSDIG_API_TOKEN  - API token for authentication
  SYSDIG_API_URL    - Base URL for Sysdig API

`, version)
}

func listCompliance(cspmClient *client.CSPMClient, policyType, platform, zoneName string, includePass bool) error {
	passMode := "failed only"
	if includePass {
		passMode = "all (passed + failed)"
	}
	fmt.Printf("Getting compliance requirements (policy: %s, platform: %s, zone: %s, mode: %s)...\n", policyType, platform, zoneName, passMode)

	// Build filter based on parameters
	filter := buildFilter(policyType, platform, zoneName, includePass)
	fmt.Printf("Filter: %s\n\n", filter)

	// Get compliance violations
	response, err := cspmClient.GetComplianceRequirements(filter)
	if err != nil {
		return fmt.Errorf("failed to get compliance requirements: %w", err)
	}

	fmt.Printf("Found %d compliance violations:\n\n", response.TotalCount.Int())

	// Display results in table format
	fmt.Printf("%-40s %-25s %-10s %-10s %-15s\n", "REQUIREMENT", "POLICY", "PLATFORM", "SEVERITY", "FAILED CONTROLS")
	fmt.Println(strings.Repeat("-", 110))

	for _, req := range response.Data {
		name := req.Name
		if len(name) > 38 {
			name = name[:35] + "..."
		}

		policyName := req.PolicyName
		if len(policyName) > 23 {
			policyName = policyName[:20] + "..."
		}

		fmt.Printf("%-40s %-25s %-10s %-10s %-15d\n",
			name, policyName, req.Platform, req.Severity, req.FailedControls)
	}

	return nil
}

func collectResources(cspmClient *client.CSPMClient, dbPath, policyType, platform, zoneName string, includePass bool, batchSize, apiDelay int) error {
	passMode := "failed only"
	if includePass {
		passMode = "all (passed + failed)"
	}
	fmt.Printf("Collecting compliance violations and resources to %s...\n", dbPath)
	fmt.Printf("Parameters: policy=%s, platform=%s, zone=%s, mode=%s\n", policyType, platform, zoneName, passMode)

	// Initialize database
	db, err := database.NewDatabase(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Step 1: Get compliance violations
	fmt.Println("\nStep 1: Getting compliance violations...")

	filter := buildFilter(policyType, platform, zoneName, includePass)
	fmt.Printf("API Filter: %s\n", filter)

	// GetAllComplianceRequirementsを使用してページネーション対応（並列化）
	complianceResponse, err := cspmClient.GetAllComplianceRequirements(filter, 50, batchSize, apiDelay)
	if err != nil {
		return fmt.Errorf("failed to get compliance requirements: %w", err)
	}

	fmt.Printf("Found %d compliance violations\n", complianceResponse.TotalCount.Int())

	// Save compliance violations to database
	if err := db.SaveComplianceRequirements(complianceResponse.Data); err != nil {
		return fmt.Errorf("failed to save compliance requirements: %w", err)
	}

	// Step 2: Get resources
	if includePass {
		fmt.Println("Step 2: Getting all resources (passed + failed) for the policy...")
	} else {
		fmt.Println("Step 2: Getting all resources that failed the policy...")
	}

	// Build policy filter - use the first policy name from compliance violations
	var policyFilter string
	if len(complianceResponse.Data) > 0 {
		if includePass {
			// Include both passed and failed resources
			policyFilter = fmt.Sprintf(`policy in ("%s")`, complianceResponse.Data[0].PolicyName)
		} else {
			// Only failed resources
			policyFilter = fmt.Sprintf(`policy.failed in ("%s")`, complianceResponse.Data[0].PolicyName)
		}
	} else {
		fmt.Println("No compliance data found, skipping resource collection")
		return nil
	}

	fmt.Printf("Policy Filter: %s\n", policyFilter)

	// Get all resources（並列化）
	resourceResponse, err := cspmClient.GetAllInventoryResources(policyFilter, 50, batchSize, apiDelay)
	if err != nil {
		return fmt.Errorf("failed to get inventory resources: %w", err)
	}

	allResources := resourceResponse.Data

	if includePass {
		fmt.Printf("Collected %d resources for the policy (API reported totalCount: %d)\n", len(allResources), resourceResponse.TotalCount.Int())
	} else {
		fmt.Printf("Collected %d resources that failed the policy (API reported totalCount: %d)\n", len(allResources), resourceResponse.TotalCount.Int())
	}

	if len(allResources) < resourceResponse.TotalCount.Int() {
		fmt.Printf("Note: API's totalCount may include items not returned by pagination. Actual collected: %d\n", len(allResources))
	}
	fmt.Println("Note: Resources are collected at policy level. Individual control-to-resource mapping is not available from the API.")

	// Save resources to database
	if len(allResources) > 0 {
		if err := db.SaveInventoryResources(allResources); err != nil {
			return fmt.Errorf("failed to save inventory resources: %w", err)
		}
	}

	fmt.Printf("Successfully collected %d violations and %d resources to %s\n",
		complianceResponse.TotalCount.Int(), len(allResources), dbPath)

	return nil
}

package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/kaz-under-the-bridge/sysdig-cspm-utils/pkg/client"
	"github.com/kaz-under-the-bridge/sysdig-cspm-utils/pkg/collector"
	"github.com/kaz-under-the-bridge/sysdig-cspm-utils/pkg/config"
	"github.com/kaz-under-the-bridge/sysdig-cspm-utils/pkg/database"
)

const version = "1.0.0"

func main() {
	var (
		configFile   = flag.String("config", "", "Path to configuration file")
		apiToken     = flag.String("token", "", "Sysdig API token")
		apiURL       = flag.String("url", "https://us2.app.sysdig.com", "Sysdig API base URL")
		command      = flag.String("command", "list", "Command to execute: list, collect, risk-collect, risk-list, risk-delete")
		dbPath       = flag.String("db", "data/cspm.db", "SQLite database path")
		policyType   = flag.String("policy", "", "Filter by policy name (comma-separated for multiple, partial match)")
		platform     = flag.String("platform", "", "Filter by platform (AWS, GCP, Azure, Kubernetes)")
		zoneName     = flag.String("zone", "Entire Infrastructure", "Filter by zone name")
		batchSize    = flag.Int("batch-size", 3, "Number of concurrent API requests for pagination (default 3)")
		apiDelay     = flag.Int("api-delay", 1, "Delay in seconds between API batches (default 1)")
		controlID    = flag.String("control-id", "", "Filter by control ID (for risk-list)")
		acceptanceID = flag.String("acceptance-id", "", "Risk acceptance ID (for risk-delete)")
		showHelp     = flag.Bool("help", false, "Show help")
		showVersion  = flag.Bool("version", false, "Show version")
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

	var err error

	// risk-list command doesn't need API token
	if *command != "risk-list" {
		// Load configuration
		cfg, loadErr := config.Load(*configFile, *apiToken, *apiURL)
		if loadErr != nil {
			log.Fatalf("Failed to load configuration: %v", loadErr)
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
			err = listCompliance(cspmClient, *policyType, *platform, *zoneName)
		case "collect":
			err = collectResources(cspmClient, *dbPath, *policyType, *platform, *zoneName, *batchSize, *apiDelay)
		case "risk-collect":
			err = collectRiskAcceptances(cspmClient, *dbPath)
		case "risk-delete":
			err = deleteRiskAcceptance(cspmClient, *dbPath, *acceptanceID)
		default:
			log.Fatalf("Unknown command: %s", *command)
		}
	} else {
		// risk-list command only reads from database
		err = listRiskAcceptances(*dbPath, *controlID)
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
        Command to execute: list, collect, risk-collect, risk-list, risk-delete (default "list")
  -db string
        SQLite database path (default "data/cspm.db")
  -control-id string
        Filter by control ID (for risk-list)
  -acceptance-id string
        Risk acceptance ID (for risk-delete)
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
  list         - List compliance requirements with violations
  collect      - Collect compliance violations and associated resources to database
  risk-collect - Collect all risk acceptances from API to database
  risk-list    - List risk acceptances from database (optionally filtered by control ID)
  risk-delete  - Delete a risk acceptance by ID (from both API and database)

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

  # Collect all risk acceptances
  sysdig-cspm-utils -token YOUR_TOKEN -command risk-collect \
    -db "data/risk_acceptances.db"

  # List all risk acceptances
  sysdig-cspm-utils -command risk-list -db "data/risk_acceptances.db"

  # List risk acceptances for a specific control
  sysdig-cspm-utils -command risk-list -db "data/risk_acceptances.db" \
    -control-id "16022"

  # Delete a risk acceptance
  sysdig-cspm-utils -token YOUR_TOKEN -command risk-delete \
    -db "data/risk_acceptances.db" \
    -acceptance-id "6763aab48ebb8c821a3ddf89"

Environment Variables:
  SYSDIG_API_TOKEN  - API token for authentication
  SYSDIG_API_URL    - Base URL for Sysdig API

`, version)
}

func listCompliance(cspmClient *client.CSPMClient, policyType, platform, zoneName string) error {
	fmt.Printf("Getting compliance requirements (policy: %s, platform: %s, zone: %s)...\n", policyType, platform, zoneName)

	// Build filter based on parameters (failed only)
	filter := buildFilter(policyType, platform, zoneName, false)
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

func collectResources(cspmClient *client.CSPMClient, dbPath, policyType, platform, zoneName string, batchSize, apiDelay int) error {
	fmt.Printf("Collecting compliance violations and control resources to %s...\n", dbPath)
	fmt.Printf("Parameters: policy=%s, platform=%s, zone=%s\n", policyType, platform, zoneName)

	// Initialize database
	db, err := database.NewDatabase(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Build filter (failed only)
	filter := buildFilter(policyType, platform, zoneName, false)
	fmt.Printf("API Filter: %s\n\n", filter)

	// Create collector and run collection
	c := collector.NewComplianceCollector(cspmClient, db)

	if err := c.CollectComplianceData(filter, 50, batchSize, apiDelay); err != nil {
		return fmt.Errorf("failed to collect compliance data: %w", err)
	}

	fmt.Println("\n✓ Collection completed successfully")

	return nil
}

func collectRiskAcceptances(cspmClient *client.CSPMClient, dbPath string) error {
	fmt.Printf("Collecting risk acceptances to %s...\n\n", dbPath)

	// Initialize database
	db, err := database.NewDatabase(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Fetch all risk acceptances from API
	fmt.Println("Fetching risk acceptances from API...")
	acceptances, err := cspmClient.ListRiskAcceptances()
	if err != nil {
		return fmt.Errorf("failed to list risk acceptances: %w", err)
	}

	fmt.Printf("\nSaving %d risk acceptances to database...\n", len(acceptances))

	// Save to database
	if err := db.SaveRiskAcceptances(acceptances); err != nil {
		return fmt.Errorf("failed to save risk acceptances: %w", err)
	}

	fmt.Println("✓ Risk acceptance collection completed successfully")
	return nil
}

func listRiskAcceptances(dbPath, controlID string) error {
	// Initialize database
	db, err := database.NewDatabase(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Fetch risk acceptances from database
	acceptances, err := db.GetRiskAcceptances(controlID)
	if err != nil {
		return fmt.Errorf("failed to get risk acceptances: %w", err)
	}

	if controlID != "" {
		fmt.Printf("Risk acceptances for control %s:\n\n", controlID)
	} else {
		fmt.Printf("All risk acceptances:\n\n")
	}

	if len(acceptances) == 0 {
		fmt.Println("No risk acceptances found")
		return nil
	}

	// Display results in table format
	fmt.Printf("%-26s %-10s %-20s %-15s %-30s\n", "ID", "CONTROL", "REASON", "ACCEPT PERIOD", "USERNAME")
	fmt.Println(strings.Repeat("-", 105))

	for _, acc := range acceptances {
		id := acc.ID
		if len(id) > 24 {
			id = id[:21] + "..."
		}

		reason := acc.Reason
		if len(reason) > 18 {
			reason = reason[:15] + "..."
		}

		username := acc.Username
		if len(username) > 28 {
			username = username[:25] + "..."
		}

		fmt.Printf("%-26s %-10s %-20s %-15s %-30s\n",
			id, acc.ControlID, reason, acc.AcceptPeriod, username)
	}

	fmt.Printf("\nTotal: %d risk acceptances\n", len(acceptances))
	return nil
}

func deleteRiskAcceptance(cspmClient *client.CSPMClient, dbPath, acceptanceID string) error {
	if acceptanceID == "" {
		return fmt.Errorf("acceptance-id is required for risk-delete command")
	}

	fmt.Printf("Deleting risk acceptance %s...\n", acceptanceID)

	// Delete from API
	fmt.Println("  Deleting from Sysdig API...")
	if err := cspmClient.DeleteRiskAcceptance(acceptanceID); err != nil {
		return fmt.Errorf("failed to delete from API: %w", err)
	}

	// Delete from database
	fmt.Println("  Deleting from database...")
	db, err := database.NewDatabase(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	if err := db.DeleteRiskAcceptanceFromDB(acceptanceID); err != nil {
		return fmt.Errorf("failed to delete from database: %w", err)
	}

	fmt.Println("✓ Risk acceptance deleted successfully")
	return nil
}

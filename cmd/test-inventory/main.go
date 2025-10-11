package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kaz-under-the-bridge/sysdig-cspm-utils/pkg/client"
)

func main() {
	apiToken := os.Getenv("SYSDIG_API_TOKEN")
	apiURL := os.Getenv("SYSDIG_API_URL")
	if apiURL == "" {
		apiURL = "https://us2.app.sysdig.com"
	}

	if apiToken == "" {
		log.Fatal("SYSDIG_API_TOKEN environment variable is required")
	}

	cspmClient := client.NewCSPMClient(apiURL, apiToken)

	// テスト1: policy.failedでフィルタ
	fmt.Println("=== Test 1: policy.failed ===")
	filter1 := `policy.failed in ("CIS Amazon Web Services Foundations Benchmark v3.0.0")`
	fmt.Printf("Filter: %s\n", filter1)

	response1, err := cspmClient.GetInventoryResources(filter1, 10)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	} else {
		fmt.Printf("SUCCESS: Found %d resources (showing first 10)\n", response1.TotalCount.Int())
		for i, res := range response1.Data {
			fmt.Printf("  %d. %s (%s/%s)\n", i+1, res.Name, res.Platform, res.Type)
		}
	}

	fmt.Println()

	// テスト2: control.failed exists（違反があるリソース全て）
	fmt.Println("=== Test 2: control.failed exists ===")
	filter2 := `control.failed exists and policy.failed in ("CIS Amazon Web Services Foundations Benchmark v3.0.0")`
	fmt.Printf("Filter: %s\n", filter2)

	response2, err := cspmClient.GetInventoryResources(filter2, 10)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	} else {
		fmt.Printf("SUCCESS: Found %d resources (showing first 10)\n", response2.TotalCount.Int())
		for i, res := range response2.Data {
			fmt.Printf("  %d. %s (%s/%s)\n", i+1, res.Name, res.Platform, res.Type)
		}
	}
}

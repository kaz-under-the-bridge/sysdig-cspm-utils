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

	// 問題のある2つのコントロール名
	controlNames := []string{
		`1.16 Ensure IAM policies that allow full "*:*" administrative privileges are not attached`,
		`5.5 Ensure routing tables for VPC peering are "least access"`,
	}

	for _, controlName := range controlNames {
		fmt.Printf("\n=== Testing control: %s ===\n", controlName)

		// テスト1: 引用符をエスケープ
		escapedName := escapeQuotes(controlName)
		filter := fmt.Sprintf(`control.failed in ("%s")`, escapedName)
		fmt.Printf("Test 1 - Escaped quotes filter: %s\n", filter)

		response, err := cspmClient.GetInventoryResources(filter, 10)
		if err != nil {
			fmt.Printf("Test 1 ERROR: %v\n", err)
		} else {
			fmt.Printf("Test 1 SUCCESS: Found %d resources\n", response.TotalCount.Int())
			continue
		}

		// テスト2: 一重引用符を使用
		filter2 := fmt.Sprintf(`control.failed in ('%s')`, controlName)
		fmt.Printf("Test 2 - Single quotes filter: %s\n", filter2)

		response2, err2 := cspmClient.GetInventoryResources(filter2, 10)
		if err2 != nil {
			fmt.Printf("Test 2 ERROR: %v\n", err2)
		} else {
			fmt.Printf("Test 2 SUCCESS: Found %d resources\n", response2.TotalCount.Int())
		}
	}
}

func escapeQuotes(s string) string {
	// バックスラッシュでエスケープ
	result := ""
	for _, ch := range s {
		if ch == '"' {
			result += `\"`
		} else {
			result += string(ch)
		}
	}
	return result
}

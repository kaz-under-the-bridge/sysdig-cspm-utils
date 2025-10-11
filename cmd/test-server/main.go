package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kaz-under-the-bridge/sysdig-cspm-utils/internal/testutil"
)

func main() {
	// Create mock server with default config
	config := testutil.DefaultMockServerConfig()
	server := testutil.NewMockServer(config)
	defer server.Close()

	fmt.Printf("🚀 CSPM Mock Server started at: %s\n", server.URL)
	fmt.Println()
	fmt.Println("📋 Available Endpoints:")
	fmt.Println("  • Compliance Requirements: /api/cspm/v1/compliance/requirements")
	fmt.Println("  • Inventory Resources:     /api/cspm/v1/inventory/resources")
	fmt.Println()
	fmt.Println("🧪 Test Examples:")
	fmt.Printf("  curl -H 'Authorization: Bearer test-token' '%s/api/cspm/v1/compliance/requirements'\n", server.URL)
	fmt.Printf("  curl -H 'Authorization: Bearer test-token' '%s/api/cspm/v1/compliance/requirements?filter=pass%%3D\"false\"%%20and%%20policy.name%%20in%%20%%28\"CIS%%20Amazon%%20Web%%20Services%%20Foundations%%20Benchmark%%20v3.0.0\"%%29'\n", server.URL)
	fmt.Printf("  curl -H 'Authorization: Bearer test-token' '%s/api/cspm/v1/inventory/resources?filter=control.failed%%20in%%20%%28\"Networking%%20-%%20Disallowed%%20Public%%20Access\"%%29'\n", server.URL)
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop the server")

	// Wait for interrupt signal to gracefully shutdown the server
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("\n🛑 Server stopped")
}

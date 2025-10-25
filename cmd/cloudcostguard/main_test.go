package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"

	"cloudcostguard/internal/estimator"
	"cloudcostguard/internal/pricing"
	"cloudcostguard/internal/terraform"
)

const binaryName = "cloudcostguard"

func TestMain(m *testing.M) {
	// Build the binary for testing
	cmd := exec.Command("go", "build", "-o", binaryName)
	if err := cmd.Run(); err != nil {
		panic(err)
	}

	// Run the tests
	result := m.Run()

	// Clean up
	os.Remove(binaryName)

	os.Exit(result)
}

func TestCostEstimationWithRealPricingData(t *testing.T) {
	priceList := pricing.NewPriceList()
	err := priceList.LoadFromFile("../../testdata/aws_pricing.json")
	if err != nil {
		t.Fatalf("Failed to load pricing data: %v", err)
	}

	tests := []struct {
		name         string
		planPath     string
		expectedCost float64
	}{
		{"Create t2.micro EC2", "../../testdata/create_ec2.json", 0.0116 * 730},
		{"Create S3 and db.t2.micro RDS", "../../testdata/create_s3_rds.json", (0.017 * 730) + 1.0}, // RDS is hourly, S3 is monthly
		{"Update t2.micro to t2.small", "../../testdata/update_ec2.json", (0.023 - 0.0116) * 730},
		{"Delete t2.micro EC2", "../../testdata/delete_ec2.json", -0.0116 * 730},
		{"Create 100GB gp2 EBS Volume", "../../testdata/create_ebs.json", 100 * 0.10}, // EBS is per GB/month
		{"Create Application Load Balancer", "../../testdata/create_lb.json", 0.0225 * 730},
		{"Create two t2.micro EC2 instances with for_each", "../../testdata/create_ec2_foreach.json", 2 * 0.0116 * 730},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			planFile, err := os.Open(tt.planPath)
			if err != nil {
				t.Fatalf("Failed to open plan file: %v", err)
			}
			defer planFile.Close()

			plan, err := terraform.ParsePlan(planFile)
			if err != nil {
				t.Fatalf("Failed to parse plan file: %v", err)
			}

			cost, err := estimator.Estimate(plan, priceList)
			if err != nil {
				t.Fatalf("Estimator failed unexpectedly: %v", err)
			}

			if fmt.Sprintf("%.2f", cost) != fmt.Sprintf("%.2f", tt.expectedCost) {
				t.Errorf("Expected cost %.2f, but got %.2f", tt.expectedCost, cost)
			}
		})
	}
}


func TestCLIArgumentPassing(t *testing.T) {
	// This test ensures that the CLI arguments are correctly parsed and used.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify that the request is for the correct repo and PR
		if !strings.Contains(r.URL.Path, "test-org/test-repo/issues/99") {
			t.Errorf("Expected request to be for test-org/test-repo/issues/99, but got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	os.Setenv("GITHUB_TOKEN", "test-token")
	os.Setenv("GITHUB_API_URL", server.URL)
	os.Setenv("CCG_TEST_MODE", "true")

	cmd := exec.Command("./"+binaryName, "analyze", "../../testdata/create_ec2.json", "test-org/test-repo", "99")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Command failed unexpectedly: %v. Output: %s", err, string(output))
	}

	if !strings.Contains(string(output), "Successfully posted cost analysis") {
		t.Errorf("Expected success message, but got: %s", string(output))
	}
}

# CloudCostGuard 🤖

**Stop guessing, start saving. Real-time, pre-commit cloud cost estimation for engineering teams.**

## The Problem: Flying Blind in the Cloud

Cloud waste is a multi-billion dollar problem. Most of it isn't malicious—it's accidental. It happens when well-meaning developers make infrastructure changes without understanding the financial impact. Existing FinOps tools are built for finance teams and only show you the bill *after* you've overspent.

## The Solution: Proactive, Developer-First Cost Feedback

CloudCostGuard is a SaaS product that bridges this gap by providing **pre-commit cloud cost estimation directly within the developer workflow**.

By integrating with your CI/CD pipeline, CloudCostGuard analyzes Terraform changes in a pull request and automatically posts a comment detailing the estimated monthly cost impact. This provides immediate, actionable feedback, empowering your engineers to make cost-conscious decisions *before* their code is merged.

## MVP Features

- **Terraform Plan Analysis:** Ingests `terraform show -json` output.
- **Real-Time AWS Pricing:** Fetches the latest pricing data from the AWS Price List API.
- **Expanded Resource Coverage:** Provides monthly cost estimates for `aws_instance`, `aws_s3_bucket`, `aws_db_instance`, `aws_ebs_volume`, and `aws_lb`.
- **Real-World Terraform Support:** Correctly handles resources created with `count` and `for_each`.
- **Flexible Configuration:** Can be configured via a `.cloudcostguard.yml` file for cleaner CI/CD integration.
- **GitHub Integration:** Posts a clear, concise cost summary directly to your pull requests.

## Getting Started

### Prerequisites

- Go 1.22 or later
- Terraform
- A GitHub Token with `repo` scope (`GITHUB_TOKEN`)

### How to Use

1. **Generate a Terraform Plan:**
   ```bash
   terraform plan -out=plan.out
   terraform show -json plan.out > plan.json
   ```

2. **Configure (Optional):**
   Create a `.cloudcostguard.yml` file in your repository root:
   ```yaml
   github:
     repo: your-org/your-repo
     pr_number: 123 # Or use CI environment variables
   ```

3. **Run the `analyze` command:**
   The tool will automatically pick up the plan file and configuration.
   ```bash
   ./cloudcostguard analyze
   ```

This will post a comment similar to this in your PR:

> ## CloudCostGuard Analysis 🤖
>
> Estimated Monthly Cost Impact: **$22.47**

## CI/CD Integration (Example with GitHub Actions)

```yaml
name: CloudCostGuard

on:
  pull_request:

jobs:
  cost-analysis:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.22

      - name: Build CloudCostGuard
        run: go build -o cloudcostguard ./cmd/cloudcostguard

      - name: Run Analysis
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          terraform plan -out=plan.out
          terraform show -json plan.out > plan.json
          ./cloudcostguard analyze plan.json ${{ github.repository }} ${{ github.event.pull_request.number }}
```

## The Future: A Full-Fledged FinOps Platform

This MVP is just the beginning. Our roadmap includes:
- **Predictive Cost Modeling:** Using ML to provide even more accurate, context-aware estimates.
- **Broader Iac/Cloud Support:** Support for all major cloud providers and IaC tools.
- **Policy Enforcement:** Set budgets and guardrails to automatically flag or block costly changes.
- **A Rich Web Dashboard:** For historical analysis and trend monitoring.

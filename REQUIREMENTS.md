# Requirements: CloudCostGuard MVP

## 1. Functional Requirements

- **FR1: Terraform Plan Analysis:**
  - The CLI must be able to ingest a Terraform plan file in its JSON representation (`terraform show -json plan.out`).
  - It must be able to parse this JSON to identify created, updated, and deleted resources.

- **FR2: Cost Estimation:**
  - The system must have a built-in price list for a core set of AWS resources:
    - `aws_instance` (EC2)
    - `aws_s3_bucket` (S3)
    - `aws_db_instance` (RDS)
  - It must calculate the estimated monthly cost of the resources identified in the Terraform plan.

- **FR3: GitHub Integration:**
  - The CLI must be able to take a GitHub repository, pull request number, and authentication token as input.
  - It must be able to post a formatted comment to the specified pull request containing a summary of the estimated cost changes.

- **FR4: End-to-End `analyze` Command:**
  - The `analyze` command must orchestrate the full workflow, from parsing the plan file to posting the GitHub comment.

## 2. Non-Functional Requirements

- **NFR1: CI/CD Integration:** The tool must be designed to be easily integrated into common CI/CD systems (e.g., GitHub Actions).
- **NFR2: Accuracy:** The cost estimates for the supported resources must be accurate to within 5% of the actual monthly cost, assuming constant usage.
- **NFR3: Performance:** The `analyze` command should complete its execution in under 30 seconds for a typical Terraform plan.

## 3. Out of Scope for MVP

- Support for cloud providers other than AWS.
- Support for IaC tools other than Terraform.
- Predictive cost modeling (this is a post-MVP feature).
- A web-based UI or dashboard.
- Custom policy enforcement.

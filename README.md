# CloudCostGuard 🤖

**Stop guessing, start saving. A production-ready FinOps platform for engineering teams.**

## Overview

CloudCostGuard is a Go-based CLI tool designed to provide pre-commit cloud cost estimation for engineering teams. It integrates with your CI/CD pipeline to analyze Terraform plans, calculate the estimated monthly cost impact of infrastructure changes, and post the results directly to your GitHub pull requests. This allows developers and teams to understand the financial implications of their changes *before* they are merged, promoting cost-conscious engineering and preventing budget overruns.

## How it Works

The workflow is simple and designed to be non-intrusive to the development process:
1. A developer creates a Terraform plan and saves it as a JSON file.
2. The `cloudcostguard` CLI is run in a CI/CD job, pointing to the plan file.
3. The CLI sends the plan to a backend service for analysis.
4. The backend service uses up-to-date AWS pricing data to calculate the monthly cost delta.
5. The CLI receives the cost estimate and posts a comment to the relevant GitHub pull request.

## Features

- **Client-Server Architecture:** A lightweight Go CLI communicates with a powerful backend service for all heavy lifting, ensuring a small footprint in your CI environment.
- **Production-Ready Pricing Engine:** The backend features a pricing service that periodically fetches AWS pricing data and stores it in a PostgreSQL database for fast, reliable queries.
- **Comprehensive Resource Coverage:** Provides monthly cost estimates for a wide range of AWS services.
- **Flexible Configuration:** The CLI can be configured via command-line arguments, environment variables, or a `.cloudcostguard.yml` file.

For a more detailed explanation of the system design, see `ARCHITECTURE.md`.

## Supported Resources

CloudCostGuard currently supports the following AWS resources:
- `aws_instance`
- `aws_db_instance`
- `aws_ebs_volume`
- `aws_lb`
- `aws_s3_bucket`
- `aws_nat_gateway`
- `aws_lambda_function`
- `aws_ecs_service` (Fargate launch type)
- `aws_eks_cluster`
- `aws_eks_node_group`

## Getting Started (Local Development)

This project is orchestrated with Docker Compose for a simple, one-command setup.

### Prerequisites

- Docker and Docker Compose
- Go (for building the CLI)
- A GitHub Personal Access Token with `repo` scope.

### 1. Initial Setup

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/your-org/cloudcostguard.git
    cd cloudcostguard
    ```

2.  **Set up environment variables:**
    Create a `.env` file in the root of the project:
    ```env
    GITHUB_TOKEN=your_github_token
    # The DATABASE_URL is pre-configured for docker-compose
    DATABASE_URL=postgres://postgres:password@db:5432/pricing?sslmode=disable
    ```
    The `GITHUB_TOKEN` is required for posting comments to pull requests.

### 2. Start the Backend Services

```bash
docker-compose up --build
```
This command will start the PostgreSQL database and the backend service. The backend will begin fetching the AWS pricing data on its first run, which may take several minutes.

### 3. Run an Analysis

In a separate terminal:

1.  **Generate a Terraform Plan:**
    Create a Terraform plan and export it to JSON. For example:
    ```bash
    terraform plan -out=plan.out
    terraform show -json plan.out > plan.json
    ```

2.  **Build the CLI:**
    ```bash
    go build -o cloudcostguard ./cmd/cloudcostguard
    ```

3.  **Run the CLI Client:**
    You can provide the repository and pull request number as arguments:
    ```bash
    ./cloudcostguard analyze plan.json your-org/your-repo 123
    ```
    Alternatively, you can configure this in `.cloudcostguard.yml`.

## Usage

The primary command for the CLI is `analyze`. This command reads a Terraform plan, sends it to the backend for cost estimation, and outputs the results.

### `analyze` Command

```bash
./cloudcostguard analyze [PLAN_JSON_PATH] [REPO] [PR_NUMBER] [flags]
```

**Arguments:**

- `PLAN_JSON_PATH` (optional): The path to the Terraform plan JSON file. Defaults to `plan.json`.
- `REPO` (optional): The GitHub repository in the format `owner/repo`.
- `PR_NUMBER` (optional): The pull request number.

**Flags:**

- `--region`: The AWS region to use for pricing (e.g., `us-west-2`). Overrides the region in the config file.
- `--format`: The output format. Can be `table` (default) or `json`.
  - `table`: Outputs a Markdown table and posts it as a comment to the specified GitHub pull request.
  - `json`: Outputs a JSON object to standard output. This is useful for programmatic analysis in CI/CD pipelines.

## Configuration

CloudCostGuard can be configured in three ways, in order of precedence:

1.  **Command-line Arguments and Flags:**
    The most direct way to configure the CLI is through command-line arguments and flags.

2.  **`.cloudcostguard.yml` file:**
    Create a `.cloudcostguard.yml` file in the directory where you run the CLI:
    ```yaml
    github:
      repo: your-org/your-repo
      pr_number: 123
    region: us-east-1
    usage_estimates:
      nat_gateway_gb_processed: 100
      lambda_monthly_requests: 1000000
      lambda_avg_duration_ms: 500
      s3_storage_gb: 100
      s3_monthly_put_requests: 10000
    ```

3.  **Environment Variables:**
    - `GITHUB_TOKEN`: (Required) Your GitHub API token.
    - `CCG_BACKEND_URL`: The URL of the CloudCostGuard backend service. Defaults to `http://localhost:8080`.
    - `DATABASE_URL`: The connection string for the PostgreSQL database.

## Testing

To run the full suite of unit and integration tests:
```bash
go test ./...
```

To run the end-to-end tests, which spin up the entire environment:
```bash
docker-compose -f docker-compose.e2e.yml up --build --abort-on-container-exit
```

## CI/CD Integration

The `cloudcostguard` CLI is a lightweight, self-contained binary. You can build this client and run it in your CI/CD pipeline. Configure it to point to your hosted CloudCostGuard backend via the `CCG_BACKEND_URL` environment variable.

## Deployment (Kubernetes)

This project includes Kubernetes manifests to deploy the backend service.

### 1. Database Migrations

Before deploying the application, you must run the database migrations. The backend Docker image includes a `migrate` command for this purpose.

```bash
docker build -t cloudcostguard/backend:latest -f backend/Dockerfile .

docker run -it --rm \
  -e DATABASE_URL="your_database_url" \
  cloudcostguard/backend:latest \
  migrate
```

### 2. Deploy to Kubernetes

The Kubernetes manifests are located in the `k8s` directory.

```bash
kubectl apply -f k8s/
```

This will create a `Deployment` and a `Service` for the backend. You will need to create a `Secret` named `db-credentials` with the database host and password.

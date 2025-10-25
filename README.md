# CloudCostGuard 🤖 (V1)

**Stop guessing, start saving. A production-ready FinOps platform for engineering teams.**

## Architecture: From CLI to SaaS

CloudCostGuard has evolved from a simple CLI tool into a client-server application to support its growth as a SaaS product. This architecture is designed for scalability, reliability, and maintainability. For more details, see `ARCHITECTURE.md`.

## Features

- **Client-Server Model:** A lightweight CLI client communicates with a powerful backend service for all heavy lifting.
- **Production-Ready Pricing Engine:** The backend features a pricing service that periodically fetches AWS pricing data and stores it in a PostgreSQL database for fast, reliable queries.
- **Expanded Resource Coverage:** Provides monthly cost estimates for `aws_instance`, `aws_s3_bucket`, `aws_db_instance`, `aws_ebs_volume`, and `aws_lb`.
- **Flexible Configuration:** The CLI can be configured via a `.cloudcostguard.yml` file.

## Getting Started (Local Development)

This project is now orchestrated with Docker Compose for a simple, one-command setup.

### Prerequisites

- Docker and Docker Compose
- A GitHub Token with `repo` scope (`GITHUB_TOKEN`)

### 1. Start the Services

```bash
docker-compose up --build
```
This will start the PostgreSQL database and the backend service. The backend will begin fetching the AWS pricing data, which may take several minutes on the first run.

### 2. Run an Analysis

In a separate terminal:

1. **Generate a Terraform Plan:**
   ```bash
   terraform plan -out=plan.out
   terraform show -json plan.out > plan.json
   ```

2. **Run the CLI Client:**
   Build and run the CLI. It is pre-configured to talk to the backend service running via Docker Compose.
   ```bash
   go build -o cloudcostguard ./cmd/cloudcostguard
   GITHUB_TOKEN=your_token ./cloudcostguard analyze plan.json your-org/your-repo 123
   ```

## CI/CD Integration

The `cloudcostguard` CLI is now a lightweight client. You can build this client and run it in your CI/CD pipeline, configuring it to point to your hosted CloudCostGuard backend via the `CCG_BACKEND_URL` environment variable.

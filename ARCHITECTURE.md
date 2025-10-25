# Architecture

CloudCostGuard has evolved from a simple CLI tool into a client-server application to support its growth as a SaaS product. This architecture is designed for scalability, reliability, and maintainability.

## System Diagram

```
[Developer's CI/CD Environment]       [CloudCostGuard SaaS Platform]
+--------------------------------+     +--------------------------------+
|                                |     |                                |
|   +------------------------+   |     |   +------------------------+   |
|   | cloudcostguard CLI     | --HTTP POST--> | Backend Service (Go)   |   |
|   | (Client)               |   |     |   | (API)                  |   |
|   +------------------------+   |     |   +-----------+------------+   |
|             |                |     |               |                |
|             |                |     |               | SQL            |
|             v                |     |               v                |
|   +------------------------+   |     |   +------------------------+   |
|   | terraform plan.json    |   |     |   | PostgreSQL Database    |   |
|   +------------------------+   |     |   | (Pricing Data)         |   |
|                                |     |   +------------------------+   |
+--------------------------------+     +--------------------------------+
```

## ADR-002: Client-Server Architecture

### Context

The original MVP was a monolithic CLI tool that contained all logic for pricing, estimation, and GitHub communication. While simple, this had several major drawbacks:
- **Scalability:** Every CLI run had to download large pricing files, which is slow and inefficient.
- **Maintainability:** The pricing and estimation logic was tightly coupled to the CLI, making it hard to update.
- **Business Model:** It did not provide a clear path to a multi-tenant SaaS offering.

### Decision

We will refactor CloudCostGuard into a client-server model.
- The **CLI** will be a lightweight client responsible only for parsing the local Terraform plan and communicating with the backend.
- The **Backend Service** will be a central Go application that contains all the complex business logic:
  - A `pricing-service` to periodically fetch and store AWS pricing data.
  - A persistent PostgreSQL database to store this data.
  - An `/estimate` API endpoint to run the cost estimation logic.

### Rationale

- **Performance & Efficiency:** The CLI is now extremely fast, as it no longer downloads multi-megabyte pricing files. Pricing data is fetched once by the backend and shared by all clients.
- **Scalability:** The backend can be scaled independently of the clients, and the database provides a robust and queryable data store.
- **Centralized Logic:** The core estimation logic is now centralized in the backend, allowing us to update and improve it without requiring users to update their CLI version. This is critical for a SaaS product.

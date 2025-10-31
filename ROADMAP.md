# CloudCostGuard Project Roadmap

This document outlines the high-level roadmap for the CloudCostGuard project, following the successful completion of the initial production hardening.

## Phase 6: Advanced Hardening

This phase focuses on improving the resilience and operational maturity of the backend service.

-   **Distributed Tracing:** Integrate OpenTelemetry to provide end-to-end tracing, offering deeper insights into request lifecycles and performance bottlenecks.
-   **CI/CD Hardening:** Enhance the GitHub Actions workflow to automatically run the full suite of tests (unit, integration, and E2E) on every pull request to catch issues earlier.
-   **Runbook Documentation:** Create a comprehensive `RUNBOOK.md` with standard operating procedures (SOPs) for common alerts, deployment procedures, and troubleshooting steps.

## Phase 7: Core Product Enhancement

This phase will focus on expanding the core cost estimation and recommendation capabilities of the platform.

-   **Expanded Resource Coverage:** Add support for additional high-value AWS services (e.g., Lambda, ECS, EKS).
-   **Improved Cost Accuracy:** Refine the pricing models to account for more complex scenarios, such as data transfer costs, provisioned IOPS, and enterprise discount programs.
-   **Advanced Recommendations:** Enhance the recommendation engine to provide more sophisticated suggestions, such as identifying idle resources or recommending architectural changes.

## Phase 8: CLI Experience Improvement

This phase will enhance the usability and feature set of the `cloudcostguard` command-line interface.

-   **Interactive Mode:** Add an interactive mode to guide users through the analysis process.
-   **Enhanced Output Formats:** Provide additional output formats, such as JSON or HTML reports, in addition to the GitHub comment.
-   **Plan Comparison:** Implement a feature to compare the cost impact of two different Terraform plans.

# Known Limitations & Future Work

Transparency is critical for building trust. This document outlines the known limitations of the CloudCostGuard MVP's cost estimation engine.

## Current Blind Spots

The MVP estimator is designed for common use cases but does not currently account for several complex pricing factors. Users should be aware that the provided estimate is a baseline and may not reflect the full, final cost.

- **Region Specificity:** All pricing is currently hardcoded to the `us-east-1` (N. Virginia) AWS region. Costs for resources in other regions will be inaccurate.
- **Reserved Instances & Savings Plans:** The estimator does not know if a company has pre-purchased reserved instances or savings plans, and therefore calculates all costs using on-demand pricing.
- **Spot Pricing:** The cost of `aws_instance` assumes on-demand pricing and does not account for the variable nature of spot instances.
- **Data Egress:** The estimator does not account for network data transfer costs, which can be a significant factor in some workloads.
- **Usage-Based Resources:** For resources like S3, the estimate only includes the baseline storage cost and does not project costs based on the number of requests (e.g., GET, PUT).
- **Tiered Pricing:** The estimator does not account for tiered pricing (e.g., the cost per GB of S3 storage decreases after the first 50TB).
- **Marketplace & Third-Party Costs:** Any software costs from the AWS Marketplace are not included.
- **Complex Terraform Modules:** The estimator does not yet fully support Terraform modules that abstract away resource definitions.

## Mitigation & Roadmap

- **Clarity in Comments:** The PR comment will eventually include a disclaimer noting that the estimate is for on-demand pricing and excludes data transfer.
- **Future Work:** Our technical roadmap is prioritized to address these limitations:
  1. **Integrate with AWS Cost & Usage Report (CUR):** To factor in a customer's actual negotiated rates, savings plans, and historical usage.
  2. **Expanded Resource Support:** Continue to add support for more AWS services.
  3. **Terraform Module Support:** Improve the parser to introspect and correctly estimate resources within modules.

# Key Performance Indicators (KPIs) & Targets

This document outlines the measurable targets for the `CloudCostGuard` product, designed to track progress towards product-market fit and business viability.

| KPI | Target | Measurement Method |
|---|---|---|
| **Time to First Value (TTV)** | **≤ 30 minutes** | Measured from the time a user signs up to the time the first PR comment is successfully posted in their repository. Tracked via application analytics. |
| **Adoption Rate** | **≥ 60%** | The percentage of pull requests containing Terraform changes that are successfully analyzed by CloudCostGuard in a pilot repository, measured over a 1-month period. |
| **Detection Value** | **≥ 30%** | The percentage of PRs where a cost delta >$100/month is flagged, and the developer subsequently makes a cost-reducing change to the PR. Measured via manual analysis during the pilot phase. |
| **Estimate Accuracy** | **≤ 20%** | The median absolute percentage error when comparing CloudCostGuard's estimates for core resources against the actual AWS bill. Validated in the pilot report. |
| **Free → Paid Conversion Rate** | **2–5%** | (Future Goal) The percentage of users on the free/trial tier who convert to a paid plan within the first 6 months of product launch. |

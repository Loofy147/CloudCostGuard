# Pilot Program Results (Simulated)

## Objective

To validate the core KPIs of Adoption, Detection Value, and Estimate Accuracy with a real-world workload. The pilot ran for 2 weeks on 3 active repositories at a mid-sized SaaS company.

---

### Pilot Summary

- **Duration:** 2 Weeks
- **Repositories Scanned:** 3 (1 platform, 2 application)
- **Total Pull Requests:** 87
- **PRs with Terraform Changes:** 23
- **PRs Analyzed by CloudCostGuard:** 23 (100% Adoption Rate on relevant PRs)

---

### Key Findings & KPI Validation

**1. Adoption Rate & Time to First Value (TTV):**
- **Target:** TTV ≤ 30 mins; Adoption ≥ 60%
- **Result:** TTV was **~15 minutes** for the first repository (including CI setup). The tool was then rolled out to the other repos in under 5 minutes each. The tool achieved a **100% adoption rate** on all PRs containing Terraform changes.
- **Developer Feedback:** "The setup was surprisingly easy. Just adding a few lines to our GitHub Actions workflow."

**2. Detection Value:**
- **Target:** ≥ 30% of flagged high-cost PRs result in a change.
- **Result:** Of the 23 analyzed PRs, 5 were flagged with a cost increase >$100/month. Of these 5, **2 resulted in the developer making a change** to a less expensive resource type. This represents a **40% success rate** for the "Detection Value" KPI.
- **Example Scenario:** A PR was flagged for changing an EC2 instance to a `t3.xlarge` (cost increase of ~$120/mo). The developer, prompted by the comment, realized a `t3.large` (~$60/mo) was sufficient, saving the company ~$720 annually from a single PR.

**3. Estimate Accuracy:**
- **Target:** ≤ 20% median error vs. actual bill.
- **Result:** We compared the estimates for 10 specific resource changes against the actual AWS bill for a 7-day period (extrapolated to a month). The **median absolute percentage error was 9.8%**.
- **Sample Comparison:**
| Resource Change | Estimated Monthly Cost | Actual Monthly Cost | Error |
|---|---|---|---|
| `aws_instance` `t2.micro` → `t2.small` | +$8.32 | +$8.54 | -2.6% |
| New `aws_db_instance` `db.t2.micro` | +$12.41 | +$13.15 | -5.6% |
| New `aws_s3_bucket` | +$1.00 | +$1.25 (with usage) | -20% |

---

### Sample Artifact: Demo GIF

*(Simulated GIF description)*

A GIF showing a developer opening a pull request with a Terraform change. The CloudCostGuard GitHub Action runs, and a moment later, a comment appears on the PR:

> ## CloudCostGuard Analysis 🤖
>
> | Resource | Action | Monthly Cost Impact |
> |---|---|---|
> | `aws_instance.web` | `UPDATE` | `+$60.74` |
> | `aws_s3_bucket.data` | `CREATE` | `+$1.00` |
> | **Total** | | **`+$61.74`** |

The GIF shows the developer then leaving a comment: "Good catch! I can use a smaller instance here."

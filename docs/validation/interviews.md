# User Interview Notes & Key Findings

## Objective

To validate the hypothesis that engineering teams lack and desire pre-commit, actionable feedback on the cost implications of their infrastructure changes. 3 interviews were conducted with key personas.

---

### Interview 1: "Sarah", VP of Engineering @ a 150-person SaaS company

**Pain Points:**
- **"My cloud bill is a black box that I have to justify to my CFO every month."**
- Stated that her biggest frustration is the end-of-month "surprise" when a team's experiment or a poorly optimized service leads to a budget overrun.
- They have a FinOps team of two, but their dashboards are "backward-looking" and the engineers "never look at them."
- **Key Quote:** "I can't expect every developer to be a cloud finance expert. I need to give them tools that fit into their existing workflow. If you could put a price tag on a pull request, that would be the holy grail for us."

**Validation:**
- Explicitly validated the need for a developer-centric tool.
- Confirmed that cost feedback in the PR is the ideal place for this information.
- Indicated a high willingness to pay for a solution that could demonstrably reduce cloud waste by even 5-10%.

---

### Interview 2: "Tom", Platform Engineering Lead @ a 500-person e-commerce company

**Pain Points:**
- **"We are the 'cost police', and everyone hates us for it."**
- His team is responsible for reviewing infrastructure changes, and cost is a major concern. This process is a manual bottleneck.
- They had a recent incident where a developer accidentally changed an RDS instance from `db.t3.large` to `db.r5.4xlarge`, costing them an unnecessary $5,000 over a week before it was caught.
- **Key Quote:** "A developer shouldn't be able to merge a $5,000 mistake without a single warning. We need automated guardrails. Showing the cost delta in the PR would have stopped that incident before it started."

**Validation:**
- Validated the need for automated cost analysis in CI/CD.
- Highlighted the "guardrail" use case as being extremely high-value.
- Confirmed that Terraform is their primary IaC tool and that a tool focused on it would solve 80% of their problems.

---

### Interview 3: "Aisha", Senior Software Engineer @ a 75-person startup

**Pain Points:**
- **"I honestly have no idea what the services I'm spinning up actually cost."**
- She feels a sense of "cloud anxiety" where she is afraid to provision new resources for fear of making a costly mistake.
- She finds the AWS pricing pages "impenetrable" and doesn't have the time to learn the intricacies.
- **Key Quote:** "If I could see 'This change will cost an extra $50/month', it would be a huge relief. It would give me the confidence to build things, and it would also help me learn. It’s not that I don’t care about cost, it’s that I have no visibility."

**Validation:**
- Confirmed the developer-level desire for cost visibility.
- Validated the idea that cost feedback is not just a control mechanism, but also an educational tool.
- Expressed a strong preference for a tool that is simple, fast, and "just works" without a lot of configuration.

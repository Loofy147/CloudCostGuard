# Security Policy

This document outlines the security posture and recommendations for using CloudCostGuard.

## 1. Data Handling

- **Plan Data:** Terraform plan files are processed within the CI/CD environment and are not stored by CloudCostGuard's services.
- **GitHub Token:** The `GITHUB_TOKEN` is used directly by the CLI in the customer's CI/CD environment to post comments. It is never transmitted to or stored by CloudCostGuard.

## 2. CI/CD Security Best Practices

We strongly recommend the following best practices when integrating CloudCostGuard into your CI/CD pipeline:

- **Use Ephemeral, Short-Lived Tokens:** When possible, use your CI provider's built-in, short-lived authentication tokens (e.g., the default `GITHUB_TOKEN` in GitHub Actions) that are automatically scoped to the repository and expire after the job is complete.
- **Principle of Least Privilege:** If you must use a classic Personal Access Token, ensure it is created with the minimum required scope (`repo` for public repositories, `public_repo` for public). Do not grant it unnecessary permissions.
- **Store Tokens as Secrets:** Always store the `GITHUB_TOKEN` as an encrypted secret in your CI/CD provider's secrets management system. Never hardcode it in your workflow files.

## 3. Multi-Account AWS Architecture

For organizations with multiple AWS accounts, CloudCostGuard's cost estimation relies on the Terraform plan being generated with the correct AWS credentials and role assumptions. The tool itself does not assume any AWS roles; it simply analyzes the resulting plan. It is the responsibility of the user to ensure that their CI/CD environment has the correct permissions to run `terraform plan` in a multi-account context.

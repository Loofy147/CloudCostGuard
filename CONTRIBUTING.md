# Contributing to CloudCostGuard

First off, thank you for considering contributing! We welcome any help, from bug reports to new features.

## Getting Started

To get started with local development, please follow the instructions in the `README.md` file. The project is fully containerized using Docker Compose, so you'll need Docker and Docker Compose installed.

### Development Workflow

1.  **Fork the repository** on GitHub.
2.  **Clone your fork** to your local machine.
3.  **Create a new branch** for your changes: `git checkout -b your-feature-name`.
4.  **Make your changes**, and be sure to add or update tests as appropriate.
5.  **Run the tests** to ensure everything is still working:
    - Unit and integration tests: `go test ./...`
    - End-to-end tests: `docker-compose -f docker-compose.e2e.yml up --build --abort-on-container-exit`
6.  **Commit your changes** with a clear and descriptive commit message.
7.  **Push your branch** to your fork on GitHub.
8.  **Open a pull request** to the main CloudCostGuard repository.

## Reporting Bugs

If you find a bug, please open an issue on our GitHub repository. Be sure to include:
- A clear and descriptive title.
- A detailed description of the problem, including steps to reproduce it.
- Any relevant logs or error messages.

## Suggesting Enhancements

If you have an idea for a new feature or an improvement to an existing one, please open an issue to discuss it. This allows us to coordinate our efforts and make sure your contribution is in line with the project's goals.

Thank you for your contributions!

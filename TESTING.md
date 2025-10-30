# Testing Strategy

This document outlines the testing strategy for CloudCostGuard.

## Unit Tests

Unit tests are written in Go and are located in the same package as the code they test. They are named `_test.go`.

To run all unit tests, use the following command:

```
go test ./...
```

## Integration Tests

Integration tests are written in Go and are located in the `backend/internal/api` package. They are named `integration_test.go`.

These tests use `dockertest` to spin up a real PostgreSQL container to test the full API lifecycle.

To run the integration tests, you will need to have Docker installed and running. Then, use the following command:

```
go test -tags=integration ./...
```

## End-to-End (E2E) Tests

E2E tests use `docker-compose` to spin up the entire application stack, including the backend, database, a mock GitHub server, and a CLI tester.

The `cli-tester` service runs a script that executes the CloudCostGuard CLI against a series of test cases and compares the output to expected results.

To run the E2E tests, you will need to have Docker and Docker Compose installed. Then, use the following command:

```
docker-compose -f docker-compose.e2e.yml up --build --abort-on-container-exit
```

## Load Tests

Load tests use `k6` to simulate concurrent user traffic and measure the API's performance and error rates under stress.

The load test script is located at `test/load/estimate_load.js`.

To run the load tests, you will need to have `k6` installed. Then, use the following command:

```
k6 run test/load/estimate_load.js
```

#!/bin/sh

set -e

echo "Running E2E tests..."

/app/cloudcostguard analyze /testdata/create_ec2.json test-org/test-repo 1 > /tmp/create_ec2.out
diff /test/e2e/expected/create_ec2.out /tmp/create_ec2.out

/app/cloudcostguard analyze /testdata/create_rds.json test-org/test-repo 2 > /tmp/create_rds.out
diff /test/e2e/expected/create_rds.out /tmp/create_rds.out

echo "E2E tests passed!"

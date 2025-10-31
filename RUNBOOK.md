# CloudCostGuard Runbook

This document provides standard operating procedures (SOPs) for common operational tasks related to the CloudCostGuard backend service.

## Deployment

The backend service is deployed to Kubernetes. The deployment process is as follows:

1.  **Build and Push the Docker Image:**
    ```bash
    docker build -t cloudcostguard/backend:latest -f backend/Dockerfile .
    docker push cloudcostguard/backend:latest
    ```

2.  **Run Database Migrations:**
    ```bash
    docker run -it --rm \
      -e DATABASE_URL="your_database_url" \
      cloudcostguard/backend:latest \
      migrate
    ```

3.  **Deploy to Kubernetes:**
    ```bash
    kubectl apply -f k8s/
    ```

## Troubleshooting

### High Error Rate

- **Check the logs:** Look for a high number of 5xx errors in the logs. The logs should include a `request_id` that can be used to trace the request through the system.
- **Check the database:** Ensure that the database is running and that the application can connect to it.
- **Check the pricing cache:** Ensure that the pricing cache is populated with data.

### High Latency

- **Check the metrics:** Look at the `http_request_duration_seconds` and `estimation_duration_seconds` metrics to identify which part of the system is slow.
- **Check the database:** Look for slow queries in the database logs.
- **Check the pricing cache:** Ensure that the pricing cache is populated with data.

## Responding to Alerts

### Database Connection Error

- **Check the database:** Ensure that the database is running and that the application can connect to it.
- **Check the credentials:** Ensure that the database credentials are correct.

### High CPU Usage

- **Check the metrics:** Look at the `http_requests_total` metric to see if there is a spike in traffic.
- **Check the logs:** Look for any errors that could be causing a high CPU usage.
- **Scale up the deployment:** If the high CPU usage is due to a spike in traffic, you may need to scale up the deployment.

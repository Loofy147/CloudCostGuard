import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
    stages: [
        { duration: '2m', target: 100 },  // Ramp to 100 users
        { duration: '5m', target: 100 },  // Stay at 100 users
        { duration: '2m', target: 0 },    // Ramp down
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'], // 95% of requests under 500ms
        http_req_failed: ['rate<0.01'],   // Less than 1% error rate
    },
};

export default function() {
    const payload = JSON.stringify({
        plan: {
            resource_changes: [
                {
                    address: "aws_instance.test",
                    type: "aws_instance",
                    change: { actions: ["create"] },
                    after: { instance_type: "t2.micro" }
                }
            ]
        }
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer test-key',
        },
    };

    const res = http.post('http://localhost:8080/estimate', payload, params);

    check(res, {
        'status is 200': (r) => r.status === 200,
        'has total cost': (r) => JSON.parse(r.body).total_monthly_cost !== undefined,
    });

    sleep(1);
}

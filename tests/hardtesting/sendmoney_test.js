import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
    stages: [
        { duration: '30s', target: 1000 },
        { duration: '1m', target: 1000 },
        { duration: '30s', target: 0 },
    ],
    thresholds: {
        'http_req_duration': ['p(95)<50'],
        'http_req_failed': ['rate<0.01'],
    },
};

export default function () {
    const payload = JSON.stringify({
        employeeId: '12345',
        amount: 10,
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer YOUR_TOKEN',
        },
    };

    const res = http.post('http://localhost:8080/api/sendCoin', payload, params);

    check(res, {
        'status is 200': (r) => r.status === 200,
    });

    sleep(1);
}
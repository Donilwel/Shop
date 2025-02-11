import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
    stages: [
        { duration: '30s', target: 1000 }, // Разгон до 1000 RPS
        { duration: '2m', target: 1000 },  // Держим нагрузку
        { duration: '30s', target: 0 },    // Завершаем
    ],
    thresholds: {
        'http_req_duration': ['p(99)<50'],  // 99% запросов < 50 мс
        'http_req_failed': ['rate<0.0001'], // 99.99% успешных запросов
    },
};

let BASE_URL = 'http://localhost:8080/api';

function authenticate() {
    let payload = JSON.stringify({ email: `user@test.com`, password: `password123` });

    let params = { headers: { 'Content-Type': 'application/json' } };
    let res = http.post(`${BASE_URL}/auth`, payload, params);

    check(res, {
        'Auth status 200': (r) => r.status === 200,
    });

    let body = JSON.parse(res.body);
    return body.token;
}

function sendCoins(token) {
    let senderID = Math.floor(Math.random() * 100000) + 1;
    let receiverID = Math.floor(Math.random() * 100000) + 1;
    while (receiverID === senderID) receiverID = Math.floor(Math.random() * 100000) + 1;

    let payload = JSON.stringify({ toUser: `user_${receiverID}`, coin: Math.floor(Math.random() * 10) + 1 });

    let params = {
        headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
        },
    };

    let res = http.post(`${BASE_URL}/sendCoin`, payload, params);

    check(res, {
        'SendCoin status 200': (r) => r.status === 200,
        'SendCoin response < 50ms': (r) => r.timings.duration < 50,
    });
}

export default function () {
    let token = authenticate();
    sendCoins(token);
    sleep(0.1); // Балансируем нагрузку
}

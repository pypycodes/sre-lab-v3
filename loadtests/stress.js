import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  stages: [
    { duration: '2m', target: 50 },
    { duration: '3m', target: 100 },
    { duration: '2m', target: 0 },
  ],
};

export default function () {
  http.get('http://localhost:8080/orders');
  http.get('http://localhost:8080/orders/1001');
  http.get('http://localhost:8080/inventory');
  sleep(0.2);
}

import http from 'k6/http';
import { check, sleep } from 'k6';
import { NEXT_BASE_URL, BLASTRA_NODE_BASE_URL, BLASTRA_GO_BASE_URL, SUMMARY_FILE } from './config.js';

export const options = {
  scenarios: {
    nextjs: {
      executor: 'constant-vus',
      vus: 10,
      duration: '30s',
      env: { TARGET: 'next' },
      tags: { app: 'nextjs' },
    },
    'blastra-node': {
      executor: 'constant-vus',
      vus: 10,
      duration: '30s',
      env: { TARGET: 'blastra-node' },
      tags: { app: 'blastra-node' },
    },
    'blastra-go': {
      executor: 'constant-vus',
      vus: 10,
      duration: '30s',
      env: { TARGET: 'blastra-go' },
      tags: { app: 'blastra-go' },
    },
  },
  thresholds: {
    'http_req_duration{app:nextjs}': ['p(95)<500'],
    'http_req_duration{app:blastra-node}': ['p(95)<500'],
    'http_req_duration{app:blastra-go}': ['p(95)<500'],
  },
  summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(95)'],
  summaryTimeUnit: 'ms',
};

export default function () {
  const target = __ENV.TARGET;
  let baseUrl, app;

  switch(target) {
    case 'next':
      baseUrl = NEXT_BASE_URL;
      app = 'nextjs';
      break;
    case 'blastra-node':
      baseUrl = BLASTRA_NODE_BASE_URL;
      app = 'blastra-node';
      break;
    case 'blastra-go':
      baseUrl = BLASTRA_GO_BASE_URL;
      app = 'blastra-go';
      break;
  }

  // Test homepage (SSR)
  const homeRes = http.get(`${baseUrl}/`, {
    tags: { app: app }
  });
  check(homeRes, {
    'homepage status is 200': (r) => r.status === 200,
    'homepage time < 200ms': (r) => r.timings.duration < 200,
    'homepage has content': (r) => r.body.includes('Project 1'),
  });
  sleep(1);

  // Test navigation links
  const aboutRes = http.get(`${baseUrl}/about`, {
    tags: { app: app }
  });
  check(aboutRes, {
    'about page status is 200': (r) => r.status === 200,
    'about page time < 200ms': (r) => r.timings.duration < 200,
  });
  sleep(1);

  // Test project link
  const projectRes = http.get(`${baseUrl}/project/1`, {
    tags: { app: app }
  });
  check(projectRes, {
    'project page status is 200': (r) => r.status === 200,
    'project page time < 200ms': (r) => r.timings.duration < 200,
  });
  sleep(1);
}

export function handleSummary(data) {
  // Create a simplified summary structure
  const summary = {
    'nextjs': {
      avg: data.metrics['http_req_duration{app:nextjs}']?.values.avg || 0,
      med: data.metrics['http_req_duration{app:nextjs}']?.values.med || 0,
      p95: data.metrics['http_req_duration{app:nextjs}']?.values['p(95)'] || 0,
      max: data.metrics['http_req_duration{app:nextjs}']?.values.max || 0
    },
    'blastra-node': {
      avg: data.metrics['http_req_duration{app:blastra-node}']?.values.avg || 0,
      med: data.metrics['http_req_duration{app:blastra-node}']?.values.med || 0,
      p95: data.metrics['http_req_duration{app:blastra-node}']?.values['p(95)'] || 0,
      max: data.metrics['http_req_duration{app:blastra-node}']?.values.max || 0
    },
    'blastra-go': {
      avg: data.metrics['http_req_duration{app:blastra-go}']?.values.avg || 0,
      med: data.metrics['http_req_duration{app:blastra-go}']?.values.med || 0,
      p95: data.metrics['http_req_duration{app:blastra-go}']?.values['p(95)'] || 0,
      max: data.metrics['http_req_duration{app:blastra-go}']?.values.max || 0
    }
  };

  return {
    [SUMMARY_FILE]: JSON.stringify(summary, null, 2)
  };
}

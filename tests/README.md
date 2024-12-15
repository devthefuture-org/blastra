# Framework Performance Benchmarks

This directory contains k6 performance tests comparing Next.js and two Blastra implementations (Node.js and Go) with identical routing structures and functionality.

## Directory Structure

```
tests/
├── config.js                # Centralized configuration
├── homepage.js             # Homepage benchmarks
├── dynamic-routes.js       # Dynamic routing benchmarks
├── analyze-results.js      # Results analysis script
├── benchmark-summary.json  # Generated test results
├── README.md              # Documentation
├── .gitignore             # Excludes build artifacts
├── next-app/              # Next.js comparison app
└── blastra-app/           # Symlink to template
```

## Test Setup

1. Install k6:
```bash
# MacOS
brew install k6

# Windows
choco install k6

# Linux
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6
```

2. Setup Next.js app:
```bash
cd tests/next-app
yarn install
yarn run build
```

3. Setup Blastra apps:
```bash
# Node.js version
cd tests/blastra-app
yarn install
yarn run build

# Go version
# Follow Go setup instructions
```

## Running Tests

1. Start Next.js app (port 3000):
```bash
cd tests/next-app
yarn start
```

2. Start Blastra Node.js app (port 5173):
```bash
cd tests/blastra-app
yarn start
```

3. Start Blastra Go app (port 8080):
```bash
# Start the Go server
```

4. Run benchmarks:
```bash
# From project root, run each test
# Each test will generate benchmark-summary.json
k6 run --out json=tests/benchmark-summary-homepage.json tests/homepage.js
k6 run --out json=tests/benchmark-summary-dynamic-routes.json tests/dynamic-routes.js
```

5. Analyze results:
```bash
node tests/analyze-results.js tests/benchmark-summary-homepage.json
node tests/analyze-results.js tests/benchmark-summary-dynamic-routes.json
```

## Configuration

The `tests/config.js` file centralizes common settings:
```javascript
// Base URLs for the frameworks
export const NEXT_BASE_URL = 'http://localhost:3000';
export const BLASTRA_NODE_BASE_URL = 'http://localhost:5173';
export const BLASTRA_GO_BASE_URL = 'http://localhost:8080';

// Path for benchmark results
export const SUMMARY_FILE = 'benchmark-summary.json';
```

## Test Scenarios

### Homepage Test (homepage.js)
Tests the root route ('/') performance:
- Initial page load
- Static content rendering
- Navigation links
- Configuration:
  * 10 concurrent users
  * 30-second duration
  * 500ms threshold for 95th percentile

### Dynamic Routes Test (dynamic-routes.js)
Tests the project routes ('/project/[projectId]'):
- Dynamic parameter handling
- Data fetching
- Configuration:
  * Ramping from 0 to 20 users
  * 2-minute duration
  * 1000ms threshold for 95th percentile

## Understanding Results

The analyzer processes the benchmark summary and provides a detailed comparison:

```
=== Performance Comparison Report ===

Response Time Comparison:
------------------------------------------------------------------------------------------------------------------
Metric          | Next.js      | Blastra Node | Blastra Go   | Blastra Node vs Next     | Blastra Go vs Next
--------------- | ------------ | ------------ | ------------ | ------------------------ | ---------------------
Average         | 168.83ms     | 164.90ms     | 2.11ms       | 2.3% faster (1.0x)       | 98.7% faster (80.0x)
Median          | 139.78ms     | 143.72ms     | 686.25µs     | 2.8% slower (1.0x)       | 99.5% faster (203.7x)
95th Percentile | 296.45ms     | 294.69ms     | 1.12ms       | 0.6% faster (1.0x)       | 99.6% faster (264.2x)
Maximum         | 846.94ms     | 814.96ms     | 814.64ms     | 3.8% faster (1.0x)       | 3.8% faster (1.0x)
```

### Metric Explanations

1. Average Response Time
   - Mean time for request completion
   - Most useful for general performance comparison
   - Considers all requests during the test

2. Median (50th Percentile)
   - Middle value of all response times
   - More stable than average for skewed distributions
   - Better represents typical user experience

3. 95th Percentile
   - 95% of requests are faster than this
   - Key indicator of consistent performance
   - Important for understanding worst-case scenarios

4. Maximum Response Time
   - Slowest recorded response
   - Useful for identifying outliers
   - Helps spot potential performance issues

### Summary Format

The benchmark summary file (`benchmark-summary.json`) contains:
- Detailed metrics for all three frameworks
- Tagged data for easy comparison
- Complete statistical information
- Test configuration details

## Notes

- All frameworks implement identical routing patterns
- Tests run against production builds
- Each framework runs on a different port to prevent interference
- Run tests multiple times for statistical significance
- All metrics are reported in milliseconds for consistency
- Summary file is generated in the tests directory
- The analyzer provides detailed error reporting if metrics are missing

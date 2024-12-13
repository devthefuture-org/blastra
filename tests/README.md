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
Average         | 6.22ms       | 2.81ms       | 2.83ms       | 54.8% faster (2.2x)      | 54.5% faster (2.2x)
Median          | 4.33ms       | 2.08ms       | 1.23ms       | 52.1% faster (2.1x)      | 71.6% faster (3.5x)
95th Percentile | 14.22ms      | 5.38ms       | 6.12ms       | 62.1% faster (2.6x)      | 56.9% faster (2.3x)
Maximum         | 136.93ms     | 18.33ms      | 96.09ms      | 86.6% faster (7.5x)      | 29.8% faster (1.4x)
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

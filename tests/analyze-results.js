#!/usr/bin/env node

import { readFileSync } from 'fs';
import { fileURLToPath } from 'url';
import { dirname } from 'path';
import { SUMMARY_FILE } from './config.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

function calculateImprovement(nextValue, value) {
    // Check for falsy values (e.g., null, undefined, 0)
    if (!value || !nextValue) {
        return 'no comparison available';
    }
    
    // Calculate percentage difference relative to Next.js
    const percentDiff = ((value - nextValue) / nextValue) * 100;
    
    // Calculate multiplier (how many times faster/slower)
    const multiplier = nextValue / value;
    
    if (value < nextValue) {
        // Service is faster than Next.js
        return `${Math.abs(percentDiff).toFixed(1)}% faster (${multiplier.toFixed(1)}x)`;
    } else {
        // Service is slower than Next.js
        return `${percentDiff.toFixed(1)}% slower (${multiplier.toFixed(1)}x)`;
    }
}


function formatTime(ms) {
    if (!ms || ms === 0) {
        return 'N/A';
    }
    if (ms < 1) {
        return `${(ms * 1000).toFixed(2)}µs`;
    }
    return `${ms.toFixed(2)}ms`;
}

function processK6Output(content) {
    const lines = content.split('\n').filter(line => line.trim());
    const metrics = {
        'nextjs': { values: [] },
        'blastra-node': { values: [] },
        'blastra-go': { values: [] }
    };

    // Process each line
    for (const line of lines) {
        try {
            const data = JSON.parse(line);
            
            // Look for http_req_duration metrics
            if (data.type === 'Point' && data.metric.startsWith('http_req_duration')) {
                const tags = data.data.tags;
                if (tags && tags.app) {
                    metrics[tags.app].values.push(data.data.value);
                }
            }
        } catch (e) {
            // Skip invalid JSON lines
            continue;
        }
    }

    // Calculate statistics for each framework
    const result = {};
    for (const [framework, data] of Object.entries(metrics)) {
        if (data.values.length === 0) continue;

        const sorted = data.values.sort((a, b) => a - b);
        const sum = sorted.reduce((a, b) => a + b, 0);
        const p95Index = Math.floor(sorted.length * 0.95);

        result[framework] = {
            avg: sum / sorted.length,
            med: sorted[Math.floor(sorted.length / 2)],
            p95: sorted[p95Index],
            max: sorted[sorted.length - 1]
        };
    }

    return result;
}

function analyzeResults(resultsPath) {
    try {
        console.log(`Reading results from: ${resultsPath}`);
        const content = readFileSync(resultsPath, 'utf8');
        const data = processK6Output(content);

        if (!data.nextjs || !data['blastra-node'] || !data['blastra-go']) {
            console.error('Missing framework data in metrics. Available data:', JSON.stringify(data, null, 2));
            process.exit(1);
        }

        console.log('\n=== Performance Comparison Report ===\n');
        
        console.log('Response Time Comparison:');
        console.log('------------------------------------------------------------------------------------------------------------------');
        console.log('Metric          | Next.js      | Blastra Node | Blastra Go   | Blastra Node vs Next     | Blastra Go vs Next');
        console.log('--------------- | ------------ | ------------ | ------------ | ------------------------ | ---------------------');

        const metrics = ['avg', 'med', 'p95', 'max'];
        const metricNames = {
            'avg': 'Average',
            'med': 'Median',
            'p95': '95th Percentile',
            'max': 'Maximum'
        };

        for (const metric of metrics) {
            const nextValue = data.nextjs[metric];
            const blastraNodeValue = data['blastra-node'][metric];
            const blastraGoValue = data['blastra-go'][metric];

            const blastraNodeComparison = calculateImprovement(nextValue, blastraNodeValue);
            const blastraGoComparison = calculateImprovement(nextValue, blastraGoValue);

            console.log(
                `${metricNames[metric].padEnd(15)} | ` +
                `${formatTime(nextValue).padEnd(12)} | ` +
                `${formatTime(blastraNodeValue).padEnd(12)} | ` +
                `${formatTime(blastraGoValue).padEnd(12)} | ` +
                `${blastraNodeComparison.padEnd(23)}  | ` +
                blastraGoComparison
            );
        }

    } catch (error) {
        console.error('Error analyzing results:', error.message);
        console.error('Make sure you are using the k6 JSON output file');
        console.error(`Default path: ${SUMMARY_FILE}`);
        process.exit(1);
    }
}

// Check if results file is provided
const resultsPath = process.argv[2] || SUMMARY_FILE;
if (!resultsPath) {
    console.error('Please provide the path to k6 JSON output file');
    console.error('Usage: node analyze-results.js [path-to-output.json]');
    console.error(`Default: ${SUMMARY_FILE}`);
    process.exit(1);
}

analyzeResults(resultsPath);

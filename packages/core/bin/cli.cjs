#!/usr/bin/env node

const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs');

// Dynamic import for ESM logger
let logger;
async function initLogger() {
  const { logger: loggerInstance } = await import('../src/utils/logger.js');
  logger = loggerInstance;
}

function spawnAsync(...args) {
  return new Promise((resolve, reject) => {
    const child = spawn(...args);
    
    child.on('error', (err) => {
      reject(err);
    });

    child.on('exit', (code) => {
      if (code === 0) {
        resolve();
      } else {
        reject(new Error(`Process exited with code ${code}`));
      }
    });
  });
}

async function handleBuild() {
  try {
    logger.info('📦 Building client bundle...');
    await spawnAsync('vite', ['build'], { stdio: 'inherit' });
    
    logger.info('📦 Building server bundle...');
    await spawnAsync('vite', [
      'build',
      '--ssr',
      'virtual:blastra/entry-server.jsx',
      '--outDir',
      'dist/server'
    ], { stdio: 'inherit' });
    
    logger.success('Build complete!');
  } catch (error) {
    logger.error('Build failed:', error.message);
    throw error;
  }
}

async function handleStart() {
  try {
    logger.info('✴️ Starting production server...');
    const coreRoot = path.resolve(__dirname, '..');
    return spawnAsync('node', [path.join(coreRoot, 'server.js')], {
      stdio: 'inherit',
      env: { ...process.env, NODE_ENV: 'production' }
    });
  } catch (error) {
    logger.error('Server start failed:', error.message);
    throw error;
  }
}

function showHelp(version) {
  logger.info(`✴️ Blastra CLI v${version} - Static Site Generator

Usage: blastra <command> [options]

Commands:
  🔥 dev         Start development server
  📦 build       Build for production (client + server)
  🚀 start       Start production server
  👀 preview     Build and preview production build
  🎨 render      Server-side render a specific URL

Examples:
  blastra dev
  blastra build
  blastra start
  blastra preview
  blastra render /about`);
}

async function main() {
  await initLogger();

  // Get command from either direct node execution or through yarn
  const args = process.argv.slice(2);
  const command = args[0];

  // Read version from package.json
  const packageJson = JSON.parse(fs.readFileSync(path.join(__dirname, '../package.json'), 'utf8'));
  const version = packageJson.version;

  // Handle help flags
  if (!command || command === '--help' || command === '-h') {
    showHelp(version);
    process.exit(command ? 0 : 1);
  }

  // Handle Ctrl+C gracefully
  process.on('SIGINT', () => {
    logger.info('\nGracefully shutting down...');
    process.exit(0);
  });

  const coreRoot = path.resolve(__dirname, '..');

  try {
    switch (command) {
      case 'dev': {
        logger.info('🔥 Starting development server...');
        await spawnAsync('node', [path.join(coreRoot, 'server.js')], {
          stdio: 'inherit',
          env: { ...process.env, NODE_ENV: 'development' }
        });
        break;
      }

      case 'build:client': {
        logger.info('📦 Building client bundle...');
        await spawnAsync('vite', ['build'], { stdio: 'inherit' });
        break;
      }

      case 'build:server': {
        logger.info('📦 Building server bundle...');
        await spawnAsync('vite', [
          'build',
          '--ssr',
          'virtual:blastra/entry-server.jsx',
          '--outDir',
          'dist/server'
        ], { stdio: 'inherit' });
        break;
      }

      case 'build': {
        await handleBuild();
        break;
      }

      case 'start': {
        await handleStart();
        break;
      }
      
      case 'preview': {
        await handleBuild();
        await handleStart();
        break;
      }

      case 'render': {
        const url = args[1];
        if (!url) {
          logger.error('Please specify a URL to render');
          logger.info('Usage: blastra render <url>');
          logger.info('Example: blastra render /about');
          process.exit(1);
        }
        logger.info('🎨 Rendering URL:', url);
        await spawnAsync('node', [path.join(coreRoot, 'output.js'), url], {
          stdio: 'inherit'
        });
        break;
      }

      default: {
        logger.error(`Unknown command: ${command}`);
        logger.info('💡 Run "blastra --help" for usage information');
        process.exit(1);
      }
    }
  } catch (error) {
    logger.error('Command execution failed:', error.message);
    process.exit(1);
  }
}

main().catch(error => {
  console.error('Failed to initialize CLI:', error);
  process.exit(1);
});

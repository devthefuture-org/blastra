#!/usr/bin/env node

import { Command } from 'commander';
import { execSync } from 'child_process';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import chalk from 'chalk';
import prompts from 'prompts';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Get package version
const packageJson = JSON.parse(fs.readFileSync(path.join(__dirname, 'package.json'), 'utf8'));
const version = packageJson.version;

// ASCII art and welcome message
console.log(chalk.cyan(`
⚡️ BLASTRA v${version} ✴️
${chalk.bold('Fast. Reliable. Stellar.')}
${chalk.dim('Prepare for liftoff... 🚀')}
`));

const program = new Command()
  .name('create-blastra')
  .description('Launch your next stellar project with Blastra')
  .version(version)
  .argument('[project-name]', 'name of your cosmic project', 'blastra-app')
  .option('-y, --yes', 'Hyperspeed mode: skip prompts, use defaults')
  .parse(process.argv);

async function enableCorepack() {
  try {
    execSync('corepack enable', { stdio: 'ignore' });
    return true;
  } catch (error) {
    return false;
  }
}

async function initializeGit(targetDir) {
  try {
    // Initialize git repository
    execSync('git init -b main', { cwd: targetDir, stdio: 'ignore' });
    console.log(chalk.green('✨ Git repository initialized on main branch'));

    // Create initial commit
    execSync('git add .', { cwd: targetDir, stdio: 'ignore' });
    execSync('git commit -m "🚀 Initial commit from Blastra"', { cwd: targetDir, stdio: 'ignore' });
    console.log(chalk.green('✨ Created initial commit'));

    return true;
  } catch (error) {
    console.warn(chalk.yellow('⚠️ Git initialization failed. You can manually initialize your space-time coordinates later.'));
    return false;
  }
}

async function initializeHusky(targetDir) {
  try {
    // Add husky as a dev dependency
    execSync('yarn add -D husky', { cwd: targetDir, stdio: 'inherit' });
    
    // Add prepare script to package.json
    const pkgJsonPath = path.join(targetDir, 'package.json');
    const pkgJson = JSON.parse(fs.readFileSync(pkgJsonPath, 'utf8'));
    pkgJson.scripts = pkgJson.scripts || {};
    pkgJson.scripts.prepare = 'husky';
    fs.writeFileSync(pkgJsonPath, JSON.stringify(pkgJson, null, 2) + '\n');
    
    // Create .husky directory and pre-commit hook
    execSync('yarn prepare', { cwd: targetDir, stdio: 'inherit' });
    fs.writeFileSync(
      path.join(targetDir, '.husky', 'pre-commit'),
      '#!/usr/bin/env sh\n. "$(dirname -- "$0")/_/husky.sh"\n\nyarn lint-staged\n',
      { mode: 0o755 }
    );
    
    console.log(chalk.green('✨ Husky hooks initialized'));
    return true;
  } catch (error) {
    console.warn(chalk.yellow('⚠️ Husky initialization failed. You can set up git hooks manually later.'));
    return false;
  }
}

async function run() {
  let projectName = program.args[0] || 'blastra-app';
  
  if (!program.opts().yes) {
    const response = await prompts({
      type: 'text',
      name: 'projectName',
      message: '🌟 Name your stellar project',
      initial: projectName
    });
    
    if (!response.projectName) {
      console.log(chalk.red('\n💥 Houston, we need a project name!'));
      process.exit(1);
    }
    projectName = response.projectName;
  }

  const targetDir = path.join(process.cwd(), projectName);

  // Check if directory exists
  if (fs.existsSync(targetDir)) {
    console.error(chalk.red(`\n💥 Cosmic collision detected! Directory ${projectName} already exists. Choose a different name or clear the space debris.`));
    process.exit(1);
  }

  console.log(chalk.cyan('\n🛸 Initiating launch sequence...'));

  // Copy template files
  const templateDir = path.join(__dirname, 'template');
  fs.cpSync(templateDir, targetDir, { recursive: true });

  // Rename gitignore file
  const gitignorePath = path.join(targetDir, 'gitignore');
  const dotGitignorePath = path.join(targetDir, '.gitignore');
  if (fs.existsSync(gitignorePath)) {
    fs.renameSync(gitignorePath, dotGitignorePath);
  }

  // Copy .env.example to .env
  const envExamplePath = path.join(targetDir, '.env.example');
  const envPath = path.join(targetDir, '.env');
  if (fs.existsSync(envExamplePath)) {
    fs.copyFileSync(envExamplePath, envPath);
    console.log(chalk.green('✨ Environment configuration initialized'));
  }

  // Update package.json with project name
  const projectPackageJsonPath = path.join(targetDir, 'package.json');
  const projectPackageJson = JSON.parse(fs.readFileSync(projectPackageJsonPath, 'utf8'));
  projectPackageJson.name = projectName;
  fs.writeFileSync(projectPackageJsonPath, JSON.stringify(projectPackageJson, null, 2) + '\n');

  // Configure yarn to use node-modules linker
  const yarnConfig = 'nodeLinker: node-modules\n';
  fs.writeFileSync(path.join(targetDir, '.yarnrc.yml'), yarnConfig);

  // Install dependencies
  console.log(chalk.cyan('\n🌌 Gathering cosmic dependencies...\n'));
  try {
    // Try to enable corepack first
    await enableCorepack();
    
    try {
      execSync('yarn install', { cwd: targetDir, stdio: 'inherit' });
    } catch (error) {
      console.log(chalk.yellow(`
🛠️  Looks like we need to calibrate your engines!

1. First, ensure you have Corepack enabled:
   ${chalk.cyan('corepack enable')}

2. Then, try the installation again:
   ${chalk.cyan('cd')} ${projectName}
   ${chalk.cyan('yarn install')}

For more information about Corepack, visit:
${chalk.blue('https://yarnpkg.com/corepack')}
`));
      process.exit(1);
    }

    // Initialize git and husky after dependencies are installed
    await initializeGit(targetDir);
    await initializeHusky(targetDir);

  } catch (error) {
    console.error(chalk.red('\n💥 Dependency gathering failed! Try manual recalibration with yarn install.'));
    process.exit(1);
  }

  console.log(chalk.green(`
🎯 Mission Success! ${projectName} has achieved stable orbit at ${targetDir}

${chalk.bold('🚀 Launch Commands:')}

  ${chalk.cyan('yarn dev')}
    Ignite the development thrusters

  ${chalk.cyan('yarn preview')}
    Test flight in production mode

  ${chalk.cyan('yarn build')}
    Prepare for deployment to production space

  ${chalk.cyan('yarn start')}
    Launch in production mode

${chalk.bold('🐳 Docker Deployment:')}
  Build your stellar container:
    ${chalk.cyan('docker build -t my-blastra-app .')}
  
  Launch into orbit:
    ${chalk.cyan('docker run -p 3000:3000 my-blastra-app')}
  
  Your app will be optimized with our Go-powered stellar engine! 🚀

${chalk.bold('🌟 Begin Your Journey:')}

  ${chalk.cyan('cd')} ${projectName}
  ${chalk.cyan('yarn dev')}

${chalk.bold('May the code be with you! 🌌')}
`));
}

run().catch((error) => {
  console.error(chalk.red('\n💥 Mission Critical Error:'), error);
  process.exit(1);
});

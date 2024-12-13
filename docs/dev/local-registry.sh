#!/bin/bash

# Exit on errors
set -e

# Install Verdaccio globally if not installed
if ! command -v verdaccio &> /dev/null; then
  echo "Installing Verdaccio..."
  npm install -g verdaccio
fi

# Start Verdaccio in the background
echo "Starting Verdaccio..."
verdaccio &

# Wait a few seconds for Verdaccio to start
sleep 5

# Set Verdaccio registry in Yarn's global configuration
echo "Configuring Yarn to use Verdaccio registry globally..."
yarn config set npmRegistryServer http://localhost:4873 --home
yarn config set unsafeHttpWhitelist --json '["localhost"]' --home

# Create a test user in Verdaccio
echo "Creating a test user in Verdaccio..."
npm adduser --registry http://localhost:4873

# Set the npmAuthToken in Yarn's global configuration
echo "Configuring Yarn authentication globally..."
TOKEN=$(grep '//localhost:4873/:_authToken=' ~/.npmrc | cut -d '=' -f 2)
yarn config set npmAuthToken $TOKEN --home

echo "Local npm registry is ready to use globally!"

# Instructions
echo -e "\nTo publish a package:"
echo "  yarn npm publish"

echo -e "\nTo reset global configuration:"
echo "  yarn config unset npmRegistryServer --home"
echo "  yarn config unset npmAuthToken --home"
echo "  yarn config unset unsafeHttpWhitelist --home"

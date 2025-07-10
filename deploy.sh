#!/bin/bash

# DQ Vault Simple Deployment Script
set -e

echo "🚀 Starting DQ Vault deployment..."

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "📄 Creating .env file..."
    touch .env
fi

# Build Docker image
echo "🔨 Building Docker image..."
docker build -t dq .

# Start containers
echo "🐳 Starting containers..."
docker-compose up -d

# Wait for containers to start
echo "⏳ Waiting for containers to start..."
sleep 10

# Get container ID
CONTAINER_ID=$(docker ps -q -f ancestor=dq)

if [ -z "$CONTAINER_ID" ]; then
    echo "❌ Error: Container not found"
    exit 1
fi

echo "✅ Container started successfully"
echo "📊 Container ID: $CONTAINER_ID"
echo ""
echo "🔧 Next steps:"
echo "1. Enter the container: docker exec -it $CONTAINER_ID sh"
echo "2. Initialize vault: vault operator init"
echo "3. Unseal vault with 3 keys: vault operator unseal (run 3 times)"
echo "4. Follow the deployment guide in DEPLOYMENT.md"
echo ""
echo "📋 Useful commands:"
echo "  View logs: docker-compose logs -f"
echo "  Check status: docker exec -it $CONTAINER_ID vault status"
echo "  Stop: docker-compose down"
echo ""
echo "🎉 Deployment script completed!" 
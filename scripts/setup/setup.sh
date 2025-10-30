#!/bin/bash

# Setup script voor Nieuws Scraper backend

set -e

echo "🚀 Setting up Nieuws Scraper Backend..."

# Check prerequisites
echo "Checking prerequisites..."

if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.22 or higher."
    exit 1
fi

if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker."
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose."
    exit 1
fi

echo "✅ Prerequisites met"

# Create .env if it doesn't exist
if [ ! -f .env ]; then
    echo "Creating .env file..."
    cp .env.example .env
    echo "✅ Created .env file (please edit with your settings)"
else
    echo "ℹ️  .env file already exists"
fi

# Download Go dependencies
echo "Downloading Go dependencies..."
go mod download
go mod tidy
echo "✅ Dependencies downloaded"

# Create necessary directories
echo "Creating directories..."
mkdir -p bin logs output data
echo "✅ Directories created"

# Start Docker services
echo "Starting Docker services..."
docker-compose up -d postgres redis nats
echo "⏳ Waiting for services to be ready..."
sleep 10

# Check if services are healthy
if docker-compose ps | grep -q "postgres.*healthy"; then
    echo "✅ PostgreSQL is ready"
else
    echo "⚠️  PostgreSQL may not be ready yet"
fi

if docker-compose ps | grep -q "redis.*healthy"; then
    echo "✅ Redis is ready"
else
    echo "⚠️  Redis may not be ready yet"
fi

echo ""
echo "✨ Setup complete!"
echo ""
echo "Next steps:"
echo "1. Edit .env file with your configuration"
echo "2. Run 'make run' to start the API locally"
echo "   OR"
echo "   Run 'make docker-up' to start all services with Docker"
echo ""
echo "API will be available at: http://localhost:8080"
echo "Health check: curl http://localhost:8080/health"
echo ""
echo "For more commands, run 'make help'"
# Docker run script for Windows PowerShell
# This script helps run the application with Docker Compose

Write-Host "Nieuws Scraper - Docker Setup" -ForegroundColor Green
Write-Host "================================" -ForegroundColor Green

# Check if Docker is installed
try {
    $dockerVersion = docker --version 2>$null
    Write-Host "✓ Docker found: $dockerVersion" -ForegroundColor Green
} catch {
    Write-Host "✗ Docker not found. Please install Docker Desktop first." -ForegroundColor Red
    Write-Host "Download from: https://www.docker.com/products/docker-desktop" -ForegroundColor Yellow
    exit 1
}

# Check if Docker Compose is available
try {
    $composeVersion = docker-compose --version 2>$null
    Write-Host "✓ Docker Compose found: $composeVersion" -ForegroundColor Green
} catch {
    try {
        $composeVersion = docker compose version 2>$null
        Write-Host "✓ Docker Compose (v2) found: $composeVersion" -ForegroundColor Green
        $useV2 = $true
    } catch {
        Write-Host "✗ Docker Compose not found" -ForegroundColor Red
        exit 1
    }
}

# Create .env file if it doesn't exist
if (!(Test-Path ".env")) {
    Write-Host "Creating .env file from .env.example..." -ForegroundColor Yellow
    Copy-Item ".env.example" ".env"
    Write-Host "✓ .env file created" -ForegroundColor Green
}

# Function to run docker-compose commands
function Run-DockerCompose {
    param([string]$command)

    if ($useV2) {
        $cmd = "docker compose $command"
    } else {
        $cmd = "docker-compose $command"
    }

    Write-Host "Running: $cmd" -ForegroundColor Cyan
    Invoke-Expression $cmd
}

# Main menu
function Show-Menu {
    Write-Host ""
    Write-Host "Choose an action:" -ForegroundColor Yellow
    Write-Host "1. Start all services (PostgreSQL + Redis + App)"
    Write-Host "2. Stop all services"
    Write-Host "3. View logs"
    Write-Host "4. Rebuild and restart"
    Write-Host "5. Clean up (remove containers and volumes)"
    Write-Host "6. Check service status"
    Write-Host "0. Exit"
    Write-Host ""
}

do {
    Show-Menu
    $choice = Read-Host "Enter your choice (0-6)"

    switch ($choice) {
        "1" {
            Write-Host "Starting all services..." -ForegroundColor Green
            Run-DockerCompose "up -d"
            Write-Host ""
            Write-Host "Services started!" -ForegroundColor Green
            Write-Host "  - API: http://localhost:8080" -ForegroundColor Cyan
            Write-Host "  - PostgreSQL: localhost:5432" -ForegroundColor Cyan
            Write-Host "  - Redis: localhost:6379" -ForegroundColor Cyan
            Write-Host ""
            Write-Host "To view logs: Run this script again and choose option 3" -ForegroundColor Yellow
        }
        "2" {
            Write-Host "Stopping all services..." -ForegroundColor Yellow
            Run-DockerCompose "down"
            Write-Host "✓ Services stopped" -ForegroundColor Green
        }
        "3" {
            Write-Host "Showing logs (press Ctrl+C to exit)..." -ForegroundColor Yellow
            Run-DockerCompose "logs -f"
        }
        "4" {
            Write-Host "Rebuilding and restarting services..." -ForegroundColor Yellow
            Run-DockerCompose "down"
            Run-DockerCompose "build --no-cache"
            Run-DockerCompose "up -d"
            Write-Host "✓ Services rebuilt and restarted" -ForegroundColor Green
        }
        "5" {
            Write-Host "Cleaning up containers and volumes..." -ForegroundColor Red
            Run-DockerCompose "down -v --remove-orphans"
            Write-Host "✓ Cleanup complete" -ForegroundColor Green
        }
        "6" {
            Write-Host "Checking service status..." -ForegroundColor Yellow
            Run-DockerCompose "ps"
        }
        "0" {
            Write-Host "Goodbye!" -ForegroundColor Green
            break
        }
        default {
            Write-Host "Invalid choice. Please try again." -ForegroundColor Red
        }
    }

    if ($choice -ne "0" -and $choice -ne "3") {
        Read-Host "Press Enter to continue"
    }

} while ($choice -ne "0")
#!/usr/bin/env pwsh
# Script to list all tables in the PostgreSQL database

Write-Host "🔍 Listing PostgreSQL Tables..." -ForegroundColor Cyan
Write-Host ""

# Load environment variables
if (Test-Path ".env") {
    Get-Content ".env" | ForEach-Object {
        if ($_ -match '^\s*([^#][^=]+?)\s*=\s*(.+?)\s*$') {
            $name = $matches[1]
            $value = $matches[2]
            [Environment]::SetEnvironmentVariable($name, $value, "Process")
        }
    }
}

# Database connection details
$POSTGRES_HOST = $env:POSTGRES_HOST
$POSTGRES_PORT = $env:POSTGRES_PORT
$POSTGRES_USER = $env:POSTGRES_USER
$POSTGRES_PASSWORD = $env:POSTGRES_PASSWORD
$POSTGRES_DB = $env:POSTGRES_DB

Write-Host "📦 Database: $POSTGRES_DB" -ForegroundColor Green
Write-Host "🖥️  Host: $POSTGRES_HOST:$POSTGRES_PORT" -ForegroundColor Green
Write-Host ""

# Set password for psql
$env:PGPASSWORD = $POSTGRES_PASSWORD

# Check if psql is available
$psqlAvailable = Get-Command psql -ErrorAction SilentlyContinue

if ($psqlAvailable) {
    Write-Host "📋 Tables in database:" -ForegroundColor Yellow
    Write-Host ""
    
    # List all tables with their row counts
    $query = @"
SELECT 
    schemaname as schema,
    tablename as table_name,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables
WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
ORDER BY schemaname, tablename;
"@
    
    psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c $query
    
    Write-Host ""
    Write-Host "📊 Table details with row counts:" -ForegroundColor Yellow
    Write-Host ""
    
    # Get row counts for each table
    $rowCountQuery = @"
SELECT 
    schemaname as schema,
    tablename as table_name,
    n_live_tup as row_count
FROM pg_stat_user_tables
ORDER BY schemaname, tablename;
"@
    
    psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c $rowCountQuery
    
    Write-Host ""
    Write-Host "🔍 Materialized views:" -ForegroundColor Yellow
    Write-Host ""
    
    # List materialized views
    $mvQuery = @"
SELECT 
    schemaname as schema,
    matviewname as view_name,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||matviewname)) as size
FROM pg_matviews
ORDER BY schemaname, matviewname;
"@
    
    psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c $mvQuery
    
} else {
    Write-Host "❌ psql command not found. Installing via Go script..." -ForegroundColor Red
    Write-Host ""
    
    # Create a temporary Go script to list tables
    $goScript = @"
package main

import (
    "database/sql"
    "fmt"
    "os"
    _ "github.com/lib/pq"
)

func main() {
    connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        os.Getenv("POSTGRES_HOST"),
        os.Getenv("POSTGRES_PORT"),
        os.Getenv("POSTGRES_USER"),
        os.Getenv("POSTGRES_PASSWORD"),
        os.Getenv("POSTGRES_DB"),
    )

    db, err := sql.Open("postgres", connStr)
    if err != nil {
        fmt.Printf("❌ Error connecting to database: %v\n", err)
        os.Exit(1)
    }
    defer db.Close()

    fmt.Println("📋 Tables in database:")
    fmt.Println()

    // List tables
    rows, err := db.Query(`
        SELECT 
            schemaname,
            tablename,
            pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
        FROM pg_tables
        WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
        ORDER BY schemaname, tablename
    `)
    if err != nil {
        fmt.Printf("❌ Error querying tables: %v\n", err)
        os.Exit(1)
    }
    defer rows.Close()

    fmt.Printf("%-15s %-30s %-10s\n", "SCHEMA", "TABLE", "SIZE")
    fmt.Println("----------------------------------------------------------------")

    for rows.Next() {
        var schema, table, size string
        if err := rows.Scan(&schema, &table, &size); err != nil {
            fmt.Printf("Error scanning row: %v\n", err)
            continue
        }
        fmt.Printf("%-15s %-30s %-10s\n", schema, table, size)
    }

    fmt.Println()
    fmt.Println("📊 Table row counts:")
    fmt.Println()

    // Get row counts
    rows2, err := db.Query(`
        SELECT 
            schemaname,
            tablename,
            n_live_tup
        FROM pg_stat_user_tables
        ORDER BY schemaname, tablename
    `)
    if err != nil {
        fmt.Printf("❌ Error querying row counts: %v\n", err)
        os.Exit(1)
    }
    defer rows2.Close()

    fmt.Printf("%-15s %-30s %-15s\n", "SCHEMA", "TABLE", "ROW COUNT")
    fmt.Println("----------------------------------------------------------------")

    for rows2.Next() {
        var schema, table string
        var count int64
        if err := rows2.Scan(&schema, &table, &count); err != nil {
            fmt.Printf("Error scanning row: %v\n", err)
            continue
        }
        fmt.Printf("%-15s %-30s %-15d\n", schema, table, count)
    }

    fmt.Println()
    fmt.Println("🔍 Materialized views:")
    fmt.Println()

    // List materialized views
    rows3, err := db.Query(`
        SELECT 
            schemaname,
            matviewname,
            pg_size_pretty(pg_total_relation_size(schemaname||'.'||matviewname)) as size
        FROM pg_matviews
        ORDER BY schemaname, matviewname
    `)
    if err != nil {
        fmt.Printf("❌ Error querying materialized views: %v\n", err)
        os.Exit(1)
    }
    defer rows3.Close()

    fmt.Printf("%-15s %-30s %-10s\n", "SCHEMA", "VIEW", "SIZE")
    fmt.Println("----------------------------------------------------------------")

    hasViews := false
    for rows3.Next() {
        hasViews = true
        var schema, view, size string
        if err := rows3.Scan(&schema, &view, &size); err != nil {
            fmt.Printf("Error scanning row: %v\n", err)
            continue
        }
        fmt.Printf("%-15s %-30s %-10s\n", schema, view, size)
    }

    if !hasViews {
        fmt.Println("No materialized views found")
    }
}
"@

    # Save and run the Go script
    $goScript | Out-File -FilePath "scripts/list-tables-temp.go" -Encoding UTF8
    
    Write-Host "Running Go script..." -ForegroundColor Cyan
    go run scripts/list-tables-temp.go
    
    # Clean up
    Remove-Item "scripts/list-tables-temp.go" -ErrorAction SilentlyContinue
}

Write-Host ""
Write-Host "✅ Done!" -ForegroundColor Green
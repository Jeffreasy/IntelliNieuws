package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	// Get database credentials from environment
	host := getEnv("POSTGRES_HOST", "localhost")
	port := getEnv("POSTGRES_PORT", "5432")
	user := getEnv("POSTGRES_USER", "postgres")
	password := getEnv("POSTGRES_PASSWORD", "postgres")
	dbname := getEnv("POSTGRES_DB", "nieuws_scraper")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbname)

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Fatalf("Error connecting to database: %v\n", err)
	}
	defer conn.Close(ctx)

	fmt.Println("=================================================================")
	fmt.Printf("Database: %s @ %s:%s\n", dbname, host, port)
	fmt.Println("=================================================================")
	fmt.Println()

	// List all tables
	fmt.Println("TABLES:")
	fmt.Println("-----------------------------------------------------------------")
	fmt.Printf("%-15s %-35s %-12s\n", "SCHEMA", "TABLE NAME", "SIZE")
	fmt.Println("-----------------------------------------------------------------")

	tablesQuery := `
		SELECT 
			schemaname,
			tablename,
			pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
		FROM pg_tables
		WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
		ORDER BY schemaname, tablename
	`

	rows, err := conn.Query(ctx, tablesQuery)
	if err != nil {
		log.Printf("Error querying tables: %v\n", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var schema, table, size string
			if err := rows.Scan(&schema, &table, &size); err != nil {
				log.Printf("Error scanning row: %v\n", err)
				continue
			}
			fmt.Printf("%-15s %-35s %-12s\n", schema, table, size)
		}
	}

	fmt.Println()
	fmt.Println()

	// Get row counts
	fmt.Println("TABLE ROW COUNTS:")
	fmt.Println("-----------------------------------------------------------------")
	fmt.Printf("%-15s %-35s %-15s\n", "SCHEMA", "TABLE NAME", "ROWS")
	fmt.Println("-----------------------------------------------------------------")

	rowCountQuery := `
		SELECT
			schemaname,
			relname,
			n_live_tup
		FROM pg_stat_user_tables
		ORDER BY n_live_tup DESC, schemaname, relname
	`

	rows2, err := conn.Query(ctx, rowCountQuery)
	if err != nil {
		log.Printf("Error querying row counts: %v\n", err)
	} else {
		defer rows2.Close()
		for rows2.Next() {
			var schema, table string
			var count int64
			if err := rows2.Scan(&schema, &table, &count); err != nil {
				log.Printf("Error scanning row: %v\n", err)
				continue
			}
			fmt.Printf("%-15s %-35s %-15d\n", schema, table, count)
		}
	}

	fmt.Println()
	fmt.Println()

	// List materialized views
	fmt.Println("MATERIALIZED VIEWS:")
	fmt.Println("-----------------------------------------------------------------")
	fmt.Printf("%-15s %-35s %-12s\n", "SCHEMA", "VIEW NAME", "SIZE")
	fmt.Println("-----------------------------------------------------------------")

	mvQuery := `
		SELECT 
			schemaname,
			matviewname,
			pg_size_pretty(pg_total_relation_size(schemaname||'.'||matviewname)) as size
		FROM pg_matviews
		ORDER BY schemaname, matviewname
	`

	rows3, err := conn.Query(ctx, mvQuery)
	if err != nil {
		log.Printf("Error querying materialized views: %v\n", err)
	} else {
		defer rows3.Close()
		hasViews := false
		for rows3.Next() {
			hasViews = true
			var schema, view, size string
			if err := rows3.Scan(&schema, &view, &size); err != nil {
				log.Printf("Error scanning row: %v\n", err)
				continue
			}
			fmt.Printf("%-15s %-35s %-12s\n", schema, view, size)
		}
		if !hasViews {
			fmt.Println("No materialized views found")
		}
	}

	fmt.Println()
	fmt.Println()

	// Show column details for main tables
	fmt.Println("COLUMN DETAILS (articles table):")
	fmt.Println("-----------------------------------------------------------------")
	fmt.Printf("%-30s %-25s %-10s\n", "COLUMN NAME", "DATA TYPE", "NULLABLE")
	fmt.Println("-----------------------------------------------------------------")

	columnsQuery := `
		SELECT 
			column_name,
			data_type,
			is_nullable
		FROM information_schema.columns
		WHERE table_schema = 'public' 
		AND table_name = 'articles'
		ORDER BY ordinal_position
	`

	rows4, err := conn.Query(ctx, columnsQuery)
	if err != nil {
		log.Printf("Error querying columns: %v\n", err)
	} else {
		defer rows4.Close()
		for rows4.Next() {
			var colName, dataType, nullable string
			if err := rows4.Scan(&colName, &dataType, &nullable); err != nil {
				log.Printf("Error scanning row: %v\n", err)
				continue
			}
			fmt.Printf("%-30s %-25s %-10s\n", colName, dataType, nullable)
		}
	}

	fmt.Println()
	fmt.Println("=================================================================")
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

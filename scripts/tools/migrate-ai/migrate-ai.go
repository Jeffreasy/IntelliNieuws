package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	fmt.Println("=== AI Migration & Processing Tool ===")
	fmt.Println()

	// Database connection string
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=nieuws_scraper sslmode=disable"

	fmt.Println("Connecting to database...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Printf("ERROR: Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		fmt.Printf("ERROR: Failed to ping database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Connected to database")
	fmt.Println()

	// Check if migration already applied
	var columnExists bool
	err = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'articles' AND column_name = 'ai_processed'
		)
	`).Scan(&columnExists)

	if err != nil {
		fmt.Printf("ERROR: Failed to check migration status: %v\n", err)
		os.Exit(1)
	}

	if columnExists {
		fmt.Println("✓ AI columns already exist in database")
		fmt.Println()

		// Check how many articles exist and how many are already processed
		var totalArticles, processedArticles int
		err = pool.QueryRow(ctx, `
			SELECT 
				COUNT(*) as total,
				COUNT(*) FILTER (WHERE ai_processed = TRUE) as processed
			FROM articles
		`).Scan(&totalArticles, &processedArticles)

		if err != nil {
			fmt.Printf("ERROR: Failed to get article counts: %v\n", err)
			os.Exit(1)
		}

		unprocessed := totalArticles - processedArticles

		fmt.Printf("Database Status:\n")
		fmt.Printf("  Total articles: %d\n", totalArticles)
		fmt.Printf("  Already processed with AI: %d\n", processedArticles)
		fmt.Printf("  Pending AI processing: %d\n", unprocessed)
		fmt.Println()

		if unprocessed > 0 {
			fmt.Println("NOTE: You have unprocessed articles!")
			fmt.Println()
			fmt.Println("Options:")
			fmt.Println("  1. Start the API server - it will automatically process articles in background")
			fmt.Println("     Command: go run cmd/api/main.go")
			fmt.Println()
			fmt.Println("  2. Or trigger processing manually via API:")
			fmt.Println("     curl -X POST http://localhost:8080/api/v1/ai/process/trigger \\")
			fmt.Println("       -H \"X-API-Key: test123geheim\"")
			fmt.Println()
			fmt.Printf("  The AI processor will handle %d articles in batches of 10\n", unprocessed)
			fmt.Printf("  Estimated time: ~%d minutes (at 5 min intervals)\n", (unprocessed/10)*5)
			fmt.Printf("  Estimated cost: ~$%.2f (at $0.002 per article)\n", float64(unprocessed)*0.002)
		} else if totalArticles > 0 {
			fmt.Println("✓ All articles are already processed with AI!")
			fmt.Println()
			fmt.Println("You can start using the AI features:")
			fmt.Println("  - Sentiment analysis")
			fmt.Println("  - Entity extraction")
			fmt.Println("  - Category classification")
			fmt.Println("  - Keyword extraction")
			fmt.Println("  - Trending topics")
		} else {
			fmt.Println("No articles in database yet.")
			fmt.Println()
			fmt.Println("Next steps:")
			fmt.Println("  1. Start the API server: go run cmd/api/main.go")
			fmt.Println("  2. Trigger a scrape: curl -X POST http://localhost:8080/api/v1/scrape \\")
			fmt.Println("       -H \"X-API-Key: test123geheim\"")
			fmt.Println("  3. Articles will be automatically processed with AI")
		}

		fmt.Println()
		fmt.Println("=== Migration Check Complete ===")
		return
	}

	// Migration not applied yet, apply it now
	fmt.Println("AI columns not found. Applying migration...")
	fmt.Println()

	// Read simple migration file (without stored procedures)
	sqlContent, err := os.ReadFile("migrations/003_add_ai_columns_simple.sql")
	if err != nil {
		fmt.Printf("ERROR: Failed to read migration file: %v\n", err)
		fmt.Println("Make sure migrations/003_add_ai_columns_simple.sql exists")
		os.Exit(1)
	}

	// Execute migration
	_, err = pool.Exec(ctx, string(sqlContent))
	if err != nil {
		fmt.Printf("ERROR: Failed to execute migration: %v\n", err)
		fmt.Println()
		fmt.Println("Error details:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("✓ SUCCESS: AI migration applied!")
	fmt.Println()
	fmt.Println("Added to database:")
	fmt.Println("  - AI processing columns (ai_processed, ai_sentiment, etc.)")
	fmt.Println("  - Indexes for efficient queries")
	fmt.Println("  - Functions for sentiment stats and trending topics")
	fmt.Println("  - Views for AI-enriched articles")
	fmt.Println()

	// Check if there are existing articles to process
	var existingArticles int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM articles").Scan(&existingArticles)
	if err != nil {
		fmt.Printf("WARNING: Could not count existing articles: %v\n", err)
	} else if existingArticles > 0 {
		fmt.Printf("Found %d existing articles that need AI processing.\n", existingArticles)
		fmt.Println()
		fmt.Println("When you start the API server, the AI processor will:")
		fmt.Printf("  - Process these %d articles in batches of 10\n", existingArticles)
		fmt.Printf("  - Run every 5 minutes automatically\n")
		fmt.Printf("  - Estimated time: ~%d minutes\n", (existingArticles/10)*5)
		fmt.Printf("  - Estimated cost: ~$%.2f (at $0.002 per article)\n", float64(existingArticles)*0.002)
		fmt.Println()
	}

	fmt.Println("Start the API server to enable AI processing:")
	fmt.Println("  go run cmd/api/main.go")
	fmt.Println()
	fmt.Println("=== Migration Complete ===")
}

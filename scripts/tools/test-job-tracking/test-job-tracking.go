package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jeffrey/intellinieuws/internal/models"
	"github.com/jeffrey/intellinieuws/internal/repository"
	"github.com/jeffrey/intellinieuws/pkg/logger"
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
	dbPool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("Error connecting to database: %v\n", err)
	}
	defer dbPool.Close()

	fmt.Println("=================================================================")
	fmt.Println("Testing Scraping Job Tracking")
	fmt.Println("=================================================================")
	fmt.Println()

	// Initialize logger and repository
	log := logger.New(logger.Config{
		Level:  "info",
		Format: "json",
	})

	jobRepo := repository.NewScrapingJobRepository(dbPool, log)

	// Test 1: Create a job
	fmt.Println("TEST 1: Creating a test job...")
	jobID, err := jobRepo.CreateJob(ctx, "test-source")
	if err != nil {
		fmt.Printf("❌ Failed to create job: %v\n", err)
		return
	}
	fmt.Printf("✅ Job created with ID: %d\n", jobID)
	fmt.Println()

	// Test 2: Start the job
	fmt.Println("TEST 2: Starting the job...")
	err = jobRepo.StartJob(ctx, jobID)
	if err != nil {
		fmt.Printf("❌ Failed to start job: %v\n", err)
		return
	}
	fmt.Printf("✅ Job %d started\n", jobID)
	fmt.Println()

	// Simulate some work
	time.Sleep(1 * time.Second)

	// Test 3: Complete the job
	fmt.Println("TEST 3: Completing the job...")
	err = jobRepo.CompleteJob(ctx, jobID, 42)
	if err != nil {
		fmt.Printf("❌ Failed to complete job: %v\n", err)
		return
	}
	fmt.Printf("✅ Job %d completed with 42 articles\n", jobID)
	fmt.Println()

	// Test 4: Create and fail a job
	fmt.Println("TEST 4: Creating and failing a job...")
	failJobID, err := jobRepo.CreateJob(ctx, "test-failed-source")
	if err != nil {
		fmt.Printf("❌ Failed to create job: %v\n", err)
		return
	}
	err = jobRepo.StartJob(ctx, failJobID)
	if err != nil {
		fmt.Printf("❌ Failed to start job: %v\n", err)
		return
	}
	err = jobRepo.FailJob(ctx, failJobID, "Test error message")
	if err != nil {
		fmt.Printf("❌ Failed to fail job: %v\n", err)
		return
	}
	fmt.Printf("✅ Job %d failed with error\n", failJobID)
	fmt.Println()

	// Test 5: Get recent jobs
	fmt.Println("TEST 5: Retrieving recent jobs...")
	jobs, err := jobRepo.GetRecentJobs(ctx, 10)
	if err != nil {
		fmt.Printf("❌ Failed to get recent jobs: %v\n", err)
		return
	}
	fmt.Printf("✅ Retrieved %d recent jobs:\n", len(jobs))
	for _, job := range jobs {
		fmt.Printf("   ID: %d, Source: %s, Status: %s, Articles: %d\n",
			job.ID, job.Source, job.Status, job.ArticleCount)
		if job.Status == models.JobStatusFailed {
			fmt.Printf("      Error: %s\n", job.Error)
		}
	}
	fmt.Println()

	// Test 6: Get job stats
	fmt.Println("TEST 6: Getting job statistics...")
	stats, err := jobRepo.GetJobStats(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to get job stats: %v\n", err)
		return
	}
	fmt.Printf("✅ Job Statistics:\n")
	if last24h, ok := stats["last_24h"].(map[string]interface{}); ok {
		fmt.Printf("   Total: %v\n", last24h["total"])
		fmt.Printf("   Completed: %v\n", last24h["completed"])
		fmt.Printf("   Failed: %v\n", last24h["failed"])
		fmt.Printf("   Total Articles: %v\n", last24h["total_articles"])
		fmt.Printf("   Avg Duration: %.2f seconds\n", last24h["avg_duration_seconds"])
	}
	fmt.Println()

	fmt.Println("=================================================================")
	fmt.Println("✅ All tests passed!")
	fmt.Println("=================================================================")
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

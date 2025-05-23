package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/halftoothed/urlite/internal/handlers"
	"github.com/halftoothed/urlite/internal/middleware"
	"github.com/halftoothed/urlite/internal/models"
	"github.com/halftoothed/urlite/internal/redis"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func init() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {

	var err error

	redis.InitRedis()

	// Databse Initialization
	dbHost := os.Getenv("DB_HOST")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbUser := os.Getenv("DB_USERNAME")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	if dbHost == "" || dbPassword == "" || dbUser == "" || dbName == "" || dbPort == "" {
		log.Fatal("Database environment variables are not properly set")
	}

	// Build DSN string dynamically
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbHost, dbUser, dbPassword, dbName, dbPort,
	)

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to the database: %v", err)
	}

	if err := db.AutoMigrate(&models.Url{}); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Database migrated successfully.")

	// GIN Setup
	router := gin.Default()

	router.Use(middleware.RateLimiter())
	router.POST("/shorten", handlers.ShortenURL(db))
	router.GET("/:code", handlers.ResolveURL(db))

	port := os.Getenv("PORT")
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Server failed to start:", err)
	}

}

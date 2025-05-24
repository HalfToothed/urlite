package handlers

import (
	"errors"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/halftoothed/urlite/internal/models"
	"github.com/halftoothed/urlite/internal/redis"
	"gorm.io/gorm"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateShortCode(n int) string {
	rand.Seed(time.Now().UnixNano())
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteByte(charset[rand.Intn(len(charset))])
	}
	return b.String()
}

func ShortenURL(db *gorm.DB) gin.HandlerFunc {

	return func(c *gin.Context) {

		var req models.ShortenRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		var shortCode string
		for {
			shortCode = generateShortCode(6)
			var existing models.Url
			if err := db.Where("short_code = ?", shortCode).First(&existing).Error; errors.Is(err, gorm.ErrRecordNotFound) {
				break // unique
			}
		}

		var expires_at *time.Time

		if req.TTL != "" {
			duration, err := time.ParseDuration(req.TTL)

			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid TTL"})
			}

			t := time.Now().Add(duration)

			expires_at = &t

		}

		newUrl := models.Url{
			ShortCode: shortCode,
			LongURL:   req.URL,
			ExpiresAt: expires_at,
		}

		if err := db.Create(&newUrl).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save URL"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"short_url": c.Request.Host + "/" + shortCode,
		})
	}
}

func ResolveURL(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		code := c.Param("code")
		key := "shorturl:" + code

		val, err := redis.Rdb.Get(redis.Ctx, key).Result()

		if err == nil {
			// Cache hit — redirect
			c.Redirect(http.StatusFound, val)
			return
		}

		var url models.Url
		if err := db.Where("short_code = ?", code).First(&url).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Short URL is not found"})
			return
		}

		log.Println(url.ExpiresAt)

		if url.ExpiresAt != nil && url.ExpiresAt.Before(time.Now()) {
			// Optional: remove from Redis
			redis.Rdb.Del(redis.Ctx, key)
			c.JSON(http.StatusGone, gin.H{"error": "URL has expired"})
			return
		}

		err = redis.Rdb.Set(redis.Ctx, key, url.LongURL, time.Hour).Err()
		if err != nil {
			log.Println("Failed to cache in Redis:", err)
		}

		c.Redirect(http.StatusFound, url.LongURL)
	}
}

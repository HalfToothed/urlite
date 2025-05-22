package handlers

import (
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/halftoothed/urlite/internal/models"
	"gorm.io/gorm"
)

type shortenRequest struct {
	URL string `json:"url" binding:"required"`
}

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

		var req shortenRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		var shortCode string
		for {
			shortCode = generateShortCode(6)
			var existing models.Url
			if err := db.Where("short_code = ?", shortCode).First(&existing).Error; err == gorm.ErrRecordNotFound {
				break // unique
			}
		}

		newUrl := models.Url{
			ShortCode: shortCode,
			LongURL:   req.URL,
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

		var url models.Url
		if err := db.Where("short_code = ?", code).First(&url).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Short URL is not found"})
			return
		}

		c.Redirect(http.StatusFound, url.LongURL)
	}
}

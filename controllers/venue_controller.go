package controllers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/umesh/ginapi/config"
	"github.com/umesh/ginapi/models"
)

// Helper function to get full image URL
func getFullImageURL(filename string) string {
	if filename == "" {
		return ""
	}
	return fmt.Sprintf("http://localhost:8080/uploads/%s", filename)
}

func CreateVenue(c *gin.Context) {
	name := c.PostForm("name")
	location := c.PostForm("location")
	size := c.PostForm("size")

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image file is required"})
		return
	}

	// Save image with unique name
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(file.Filename))
	savePath := filepath.Join("uploads", filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Insert into database
	result, err := config.DB.Exec(`
		INSERT INTO venues (name, location, size, image)
		VALUES (?, ?, ?, ?)`,
		name, location, size, filename,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	venue := models.Venue{
		ID:       uint(id),
		Name:     name,
		Location: location,
		Size:     size,
		Image:    getFullImageURL(filename), // Store full URL in response
	}

	c.JSON(http.StatusCreated, venue)
}

func GetVenues(c *gin.Context) {
	rows, err := config.DB.Query("SELECT id, name, location, size, image FROM venues")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var venues []models.Venue
	for rows.Next() {
		var venue models.Venue
		var imageFilename string
		err := rows.Scan(&venue.ID, &venue.Name, &venue.Location, &venue.Size, &imageFilename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		venue.Image = getFullImageURL(imageFilename)
		venues = append(venues, venue)
	}

	c.JSON(http.StatusOK, venues)
}

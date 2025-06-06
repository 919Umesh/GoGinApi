package utils

import "github.com/gin-gonic/gin"

func RespondWithError(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{"error": message})
	c.Abort()
}

func RespondWithJSON(c *gin.Context, code int, payload interface{}) {
	c.JSON(code, payload)
}

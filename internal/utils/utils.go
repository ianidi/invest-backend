package utils

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

//Error responds with a error
func Error(c *gin.Context, err error) bool {
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return true
	}
	return false
}

func ShouldBindJSON(c *gin.Context, query interface{}) error {
	if err := c.ShouldBindJSON(query); err != nil {
		return errors.New("VALIDATION")
	}

	return nil
}

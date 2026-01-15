package middleware

import (
	"log"
	"net/http"

	"github.com/douglasl/tokyo-commute-optimizer/internal/models"
	"github.com/gin-gonic/gin"
)

// ErrorHandler is a middleware that handles panics and errors
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)

				c.JSON(http.StatusInternalServerError, models.ErrorResponse{
					Error:   "Internal Server Error",
					Message: "An unexpected error occurred",
					Code:    http.StatusInternalServerError,
				})

				c.Abort()
			}
		}()

		c.Next()

		// Check if there are any errors in the context
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Request Error",
				Message: err.Error(),
				Code:    http.StatusInternalServerError,
			})
		}
	}
}

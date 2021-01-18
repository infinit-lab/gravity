package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func SessionMiddle() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if len(token) == 0 {
			token = c.Query("token")
		}

		if len(token) != 0 {
			_, err := GetSession(token)
			if err == nil {
				UpdateSession(token)
				c.Set("Token", token)
				c.Next()
				return
			}
		}
		response := Response{
			Result: false,
			Error:  "Unauthorized",
		}
		c.JSON(http.StatusUnauthorized, response)
		c.Abort()
		return
	}
}

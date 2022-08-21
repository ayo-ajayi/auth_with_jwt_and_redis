package ctrl

import (
	//"log"
	"net/http"

	"github.com/ayo-ajayi/proj/token"
	"github.com/gin-gonic/gin"
)

func TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err:= token.TokenValid(c.Request);err != nil {
			c.JSON(http.StatusUnauthorized, err.Error())
			c.Abort()
			return
		}
		c.Next()
	}
}
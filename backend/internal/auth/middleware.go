package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const CtxUserID = "userID"

func BearerMiddleware(s Signer) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		parts := strings.SplitN(h, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "missing token",
			})
			return
		}
		claims, err := s.Parse(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "invalid token",
			})
			return
		}
		c.Set(CtxUserID, claims.Subject)
		c.Next()
	}
}

func UserID(c *gin.Context) (string, bool) {
	v, ok := c.Get(CtxUserID)
	if !ok {
		return "", false
	}
	id, ok := v.(string)
	return id, ok && id != ""
}

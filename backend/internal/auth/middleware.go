package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const CtxUserID = "userID"

func BearerMiddleware(s Signer, cookieName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerFromHeader(c.GetHeader("Authorization"))
		if token == "" && cookieName != "" {
			if cookie, err := c.Request.Cookie(cookieName); err == nil && cookie.Value != "" {
				token = cookie.Value
			}
		}
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "missing token",
			})
			return
		}
		claims, err := s.Parse(token)
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

func bearerFromHeader(header string) string {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 {
		return ""
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}

func UserID(c *gin.Context) (string, bool) {
	v, ok := c.Get(CtxUserID)
	if !ok {
		return "", false
	}
	id, ok := v.(string)
	return id, ok && id != ""
}

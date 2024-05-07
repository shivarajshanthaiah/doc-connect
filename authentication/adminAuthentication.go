package authentication

import (
	"doc-connect/models"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var jwtkey = []byte("adminkey")

func GenerateAdminToken(username string) (string, error) {
	//setting token expiration time

	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &models.AdminClaims{
		Username:       username,
		StandardClaims: jwt.StandardClaims{ExpiresAt: expirationTime.Unix()},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtkey)
}

// verify Admin Token
func AdminAuthentication(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.AdminClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtkey, nil
	})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*models.AdminClaims); ok && token.Valid {
		return claims.Username, nil
	}
	return "", errors.New("invalid token")
}

//Admin Auth middleware

func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing the authorization header"})
			return
		}

		authHeader := strings.TrimSpace(strings.TrimPrefix(tokenString, "Bearer"))

		username, err := AdminAuthentication(authHeader)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("username", username)
		c.Next()
	}
}

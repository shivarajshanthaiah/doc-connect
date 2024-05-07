package authentication

import (
	"doc-connect/configuration"
	"doc-connect/models"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtkeyy = []byte("doctorkey")

// Generating token
func GenerateDoctorToken(doctorEmail string, doctorId uint) (string, error) {
	//setting token expiration time
	claims := &models.DoctorClaims{
		Id:          doctorId,
		DoctorEmail: doctorEmail,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtkeyy)
}

// verify Doctor Token
func DoctorAuthentication(tokenString string) (string, uint, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.DoctorClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtkeyy, nil
	})
	if err != nil {
		return "", 0, err
	}

	if claims, ok := token.Claims.(*models.DoctorClaims); ok && token.Valid {
		return claims.DoctorEmail, claims.Id, nil
	}
	return "", 0, errors.New("invalid token")
}

// Doctor Auth middleware
func DoctorAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing the authorization header"})
			return
		}

		authHeader := strings.TrimSpace(strings.TrimPrefix(tokenString, "Bearer"))

		email, id, err := DoctorAuthentication(authHeader)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		fmt.Println(id, email)
		c.Set("email", email)
		c.Set("doctor_id", id)
		c.Next()
	}
}

// retrieves Doctor information from the database
func GetDoctorByEmail(email string) (*models.Doctor, error) {
	var doctor models.Doctor
	if err := configuration.DB.Where("email = ?", email).First(&doctor).Error; err != nil {
		return nil, err
	}
	return &doctor, nil
}

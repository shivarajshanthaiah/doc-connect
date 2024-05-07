package authentication

import (
	"doc-connect/models"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var jwtKey = []byte("secretKey")

// Generating jwt token for patient
func GeneratePatientToken(patientID int, phone string) (string, error) {

	claims := &models.PatientClaims{
		PatientID: patientID,
		Phone:     phone,
		StandardClaims: jwt.StandardClaims{
			// ExpiresAt: expirationTime.Unix()},
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			IssuedAt:  time.Now().Unix(),
		}}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil

}

func AuthenticatePatient(signedStringToken string) (string, int, error) {
	// Parse the token
	var patientClaims models.PatientClaims
	token, err := jwt.ParseWithClaims(signedStringToken, &patientClaims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtKey), nil // Replace with your secret key
	})

	if err != nil {
		return "", 0, err
	}
	//check the token is valid
	if !token.Valid {
		return "", 0, errors.New("token is not valid")
	}
	//type assert the claims from the token object
	claims, ok := token.Claims.(*models.PatientClaims)

	if !ok {
		err = errors.New("couldn't parse claims")
		return "", 0, err
	}
	phone := claims.Phone
	patientIDS := float64(claims.PatientID)
	if claims.ExpiresAt < time.Now().Unix() {
		err = errors.New("token expired")
		return "", 0, err
	}
	patientID := int(patientIDS)

	return phone, patientID, nil
}

func PatientAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from the request header or other sources
		tokenString := c.GetHeader("Authorization")

		// Check if token exists
		if tokenString == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "User Authorization is missing"})
			return
		}

		// Trim the token to get the actual token string
		authHeader := strings.Replace(tokenString, "Bearer ", "", 1)
		phone, patientID, err := AuthenticatePatient(authHeader)
		if err != nil {
			//fmt.Println("Error authenticating user:", err)
			c.AbortWithStatusJSON(401, gin.H{"error": err.Error()})
			return
		}
		c.Set("patientID", patientID)
		fmt.Println("Authenticated user:", phone)

	}
}

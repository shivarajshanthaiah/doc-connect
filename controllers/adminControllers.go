package controllers

import (
	"doc-connect/authentication"
	"doc-connect/configuration"
	"doc-connect/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Amin login
func AdminLogin(c *gin.Context) {
	var admin models.Admin
	if err := c.BindJSON(&admin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var dbAdmin models.Admin
	if err := configuration.DB.Where("username = ?", admin.Username).First(&dbAdmin).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	if len(dbAdmin.Password) > 0 && dbAdmin.Password[0] == '$' {
		if err := bcrypt.CompareHashAndPassword([]byte(dbAdmin.Password), []byte(admin.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"erroe": "Invalid username or password"})
			return
		}
	} else {
		if dbAdmin.Password != admin.Password {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(admin.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		dbAdmin.Password = string(hashedPassword)
		if err := configuration.DB.Save(&dbAdmin).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
			return
		}
	}

	token, err := authentication.GenerateAdminToken(admin.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "token": token})
}

func AdminLogout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

package controllers

import (
	"context"
	"doc-connect/authentication"
	"doc-connect/configuration"
	"doc-connect/models"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var validate = validator.New()

// Signup handles the registration of a new doctor.
func Signup(c *gin.Context) {
	var doctor models.Doctor

	if err := c.ShouldBindJSON(&doctor); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"Status":  "Failed",
			"message": "Binding error",
			"data":    err.Error(),
		})
		return
	}

	// Validate doctor struct fields
	if err := validate.Struct(doctor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Failed",
			"message": "Please fill all the mandatory fields",
			"data":    err.Error(),
		})
		return
	}

	// Check if email is already in use
	var existingDoctor models.Doctor
	if err := configuration.DB.Where("email = ?", doctor.Email).First(&existingDoctor).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"status":  "Failed",
			"message": "Email already in use",
			"data":    "Choose another email",
		})
		return
	} else if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Staus":   "Failed",
			"message": "Database error",
			"data":    err.Error(),
		})
		return
	}

	// Check if phone number is already in use
	if err := configuration.DB.Where("phone = ?", doctor.Phone).First(&existingDoctor).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"status":  "Failed",
			"message": "Phone number already in use",
			"data":    "Choose another phone number",
		})
		return

	} else if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Staus":   "Failed",
			"message": "Database error",
			"data":    err.Error(),
		})
		return
	}

	// Check if licence is already in use
	if err := configuration.DB.Where("license_number = ?", doctor.LicenseNumber).First(&existingDoctor).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"status":  "Failed",
			"message": "Licence number already in use",
			"data":    "Choose another Licence number",
		})
		return

	} else if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Staus":   "Failed",
			"message": "Database error",
			"data":    err.Error(),
		})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(doctor.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Failed",
			"message": "Failed to hash password",
			"data":    err.Error(),
		})
		return
	}

	doctor.Password = string(hashedPassword)

	//Check if hospital ID exists and is active
	var hospital models.Hospital
	if err := configuration.DB.First(&hospital, doctor.HospitalID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"Error": "Hospital doesn't exists"})
		return
	}

	if hospital.Status == "Deactive" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Hospital doesn't exists"})
		return
	}

	// Generate OTP and send it via email
	otp := authentication.GenerateOTP(6)
	authentication.SendOTPByEmail(otp, doctor.Email)

	// Marshal doctor data to JSON
	jsonData, err := json.Marshal(doctor)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "Failed",
			"message": "Failed to marshal json data",
			"data":    err.Error(),
		})
		return
	}

	// Store OTP in Redis with a key based on the doctor's email
	if err := configuration.Client.Set(context.Background(), "otp"+doctor.Email, otp, 300*time.Second).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Failed",
			"message": "Redis error",
			"data":    err.Error(),
		})
		return
	}

	// Store doctor data in Redis with a key based on the doctor's email
	if err := configuration.Client.Set(context.Background(), "user"+doctor.Email, jsonData, 1200*time.Second).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Failed",
			"message": "Redis error",
			"data":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Success",
		"message": "Go to verfication page",
		"data":    nil,
	})
}

// VerifyOTP for doctor signup.
func VerifyOTP(c *gin.Context) {
	var doctorData models.Doctor
	type OTPString struct {
		Email string `json:"email"`
		Otp   string `json:"otp"`
	}

	var doctor OTPString
	if err := c.ShouldBindJSON(&doctor); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"status":  "Failed",
			"message": "Binding error",
			"data":    err.Error(),
		})
		return
	}
	if doctor.Otp == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "Failed",
			"message": "OTP not entered",
			"data":    nil,
		})
		return
	}

	// Retrieve OTP from Redis
	otp, err := configuration.Client.Get(context.Background(), "otp"+doctor.Email).Result()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "Failed",
			"message": "otp not found",
			"data":    err.Error(),
		})
		return
	}

	// Validate OTP
	if authentication.ValidateOTP(otp, doctor.Otp) {
		// If OTP is valid, retrieve doctor data from Redis
		user, err := configuration.Client.Get(context.Background(), "user"+doctor.Email).Result()
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"status": "Failed",
				"message": "User details missing",
				"data":    err.Error(),
			})
			return
		}

		// Unmarshal doctor data
		err = json.Unmarshal([]byte(user), &doctorData)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "Failed",
				"message": "Error in unmarshaling json data",
				"data":    err.Error(),
			})
			return
		}

		// Create doctor record in the database
		doctorData.Verified = "false"
		doctorData.Approved = "false"
		configuration.DB.Create(&doctorData)
		c.JSON(http.StatusOK, gin.H{
			"status":  "Success",
			"message": "Signup successful",
			"data":    doctorData,
		})
	}

}

// DoctorLogin
func DoctorLogin(c *gin.Context) {
	var doctors models.Doctor
	if err := c.BindJSON(&doctors); err != nil {
		c.JSON(400, gin.H{"Error": err.Error()})
		return
	}

	// Finding doctor by email
	var existingDoctor models.Doctor
	if err := configuration.DB.Where("email = ?", doctors.Email).First(&existingDoctor).Error; err != nil {
		c.JSON(401, gin.H{"error": "invalid is email"})
		return
	}

	// Comparing password hashes
	if err := bcrypt.CompareHashAndPassword([]byte(existingDoctor.Password), []byte(doctors.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}

	// Checking if the doctor is approved
	if existingDoctor.Approved != "true" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Doctor not approved yet"})
		return
	}

	// Generating JWT token for authenticated doctor
	token, err := authentication.GenerateDoctorToken(doctors.Email, existingDoctor.DoctorID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "token": token})

}

package controllers

import (
	"doc-connect/authentication"
	"doc-connect/configuration"
	"doc-connect/models"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/twilio/twilio-go"
	verify "github.com/twilio/twilio-go/rest/verify/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

// PatientLogin handles the patient login process
func PatientLogin(c *gin.Context) {
	var loginReq struct {
		Phone    string `json:"phone" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.BindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if the provided phone number exists in the database
	var existingPatient models.Patient
	if err := configuration.DB.Where("phone = ?", loginReq.Phone).First(&existingPatient).Error; err != nil {
		// Phone number not found in the database
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid phone number or phone number is not present"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(existingPatient.Password), []byte(loginReq.Password)); err != nil {
		// Incorrect password
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid phone number or password"})
		return
	}

	// Generate JWT token for the patient
	token, err := authentication.GeneratePatientToken(existingPatient.PatientID, loginReq.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Return the token
	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"message": "Login sucessful",
		"token":   token,
	})
}

// Function to handle patient signup
func PatientSignup(c *gin.Context) {
	var patient models.Patient
	// Binding JSON data to patient struct
	if err := c.BindJSON(&patient); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println(patient)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(patient.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	patient.Password = string(hashedPassword)

	var existingPatient models.Patient
	if err := configuration.DB.Where("phone = ?", patient.Phone).First(&existingPatient).Error; err == nil {
		// Patient already exists, return error
		c.JSON(http.StatusConflict, gin.H{"message": "Patient already exists"})
		return
	} else if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "database error"})
		return
	}

	// Send OTP to the patient's phone number
	err1 := SendOTP(patient.Phone)
	if err1 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send OTP", "data": err1.Error()})
		return
	}

	patientData, err := json.Marshal(&patient)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal patient", "data": err.Error()})
		return
	}
	// Store phone number in Redis for OTP verification
	key := fmt.Sprintf("user:%s", patient.Phone)
	err = configuration.SetRedis(key, patientData, time.Minute*5)
	if err != nil {
		fmt.Println("Error setting user in Redis:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"Status": false, "Data": nil, "Message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Message": "Otp generated successfully. Proceed to verification page>>>"})

}

// Function to send OTP to the patient's phone number
func SendOTP(phoneNumber string) error {
	accountSID := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTHTOKEN")

	// Initialize Twilio client
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSID,
		Password: authToken,
	})

	//create SMS message for OTP verification
	from := os.Getenv("TWILIO_PHONENUMBER")
	params := verify.CreateVerificationParams{}
	params.SetTo("+918762334325")
	params.SetChannel("sms")
	println(from)
	response, err := client.VerifyV2.CreateVerification(os.Getenv("TWILIO_SERVIES_ID"), &params)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	fmt.Println(response)
	return nil
}

// Function to verify OTP and create patient record
func UserOtpVerify(c *gin.Context) {
	accountSID := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTHTOKEN")

	// Bind OTP verification request data
	var OTPverify models.VerifyOTP
	if err := c.BindJSON(&OTPverify); err != nil {
		// fmt.Println("i'm here")
		fmt.Println("Error parsing JSON:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"Status": false, "Data": nil, "Message": "Failed to parse JSON data"})
		return
	}

	// Check if OTP is empty
	if OTPverify.Otp == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Status": false, "Message": "OTP is required"})
	}

	// Initialize Twilio client for OTP verification
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSID,
		Password: authToken,
	})

	// Create parameters for OTP verification check
	params := verify.CreateVerificationCheckParams{}
	params.SetTo("+918762334325")
	params.SetCode(OTPverify.Otp)

	// Verify OTP with Twilio
	response, err := client.VerifyV2.CreateVerificationCheck(os.Getenv("TWILIO_SERVIES_ID"), &params)
	if err != nil {
		fmt.Println("err", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"Status": false, "Data": nil, "Message": "error in veifying provided OTP"})
		return
	} else if *response.Status != "approved" {
		c.JSON(http.StatusInternalServerError, gin.H{"Status": false, "Data": nil, "Message": "Wrong OTP provided"})
		return
	}

	// Retrieve patient data from Redis
	key := fmt.Sprintf("user:%s", OTPverify.Phone)
	value, err := configuration.GetRedis(key)
	if err != nil {
		fmt.Println("Error retrieving OTP from Redis:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"Data":    nil,
			"Message": "Internal server error",
		})
		return
	}

	// Bind user data from request body
	var userData models.Patient

	err = json.Unmarshal([]byte(value), &userData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unmarshal patient", "data": err.Error()})
		return
	}

	// Create user record
	if err := configuration.DB.Create(&userData).Error; err != nil {
		fmt.Println("Error creating Patient:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"Status": false, "Data": nil, "Message": "Failed to create user"})
		return
	}

	// Create wallet for the user with balance 0
	wallet := models.Wallet{
		UserID: userData.PatientID,
		Amount: 0,
	}

	// Create wallet record
	if err := configuration.DB.Create(&wallet).Error; err != nil {
		fmt.Println("Error creating Wallet:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"Status": false, "Data": nil, "Message": "Failed to create wallet"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  true,
		"Message": "OTP verified successfully and user has been created. Login to continue...",
	})
}

// User logout
func PatientLogout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "You are successfully logged out"})
}

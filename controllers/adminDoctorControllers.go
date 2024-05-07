package controllers

import (
	"doc-connect/configuration"
	"doc-connect/models"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// View verified doctors
func ViewVerifiedDoctors(c *gin.Context) {
	var doctors []models.Doctor

	if err := configuration.DB.Where("verified = ?", "true").Find(&doctors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching verified Doctors"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Verified doctors list fetched sucessfully",
		"data":    doctors,
	})
}

// View Not verified doctors
func ViewNotVerifiedDoctors(c *gin.Context) {
	var doctors []models.Doctor

	if err := configuration.DB.Where("verified = ?", "false").Find(&doctors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching Doctors list"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Doctors list fetched sucessfully",
		"data":    doctors,
	})
}

// View Verified and approved dospitals
func ViewVerifiedApprovedDoctors(c *gin.Context) {
	var doctors []models.Doctor

	if err := configuration.DB.Where("verified = ? AND approved = ?", "true", "true").Find(&doctors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching verified and approved Doctors"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Verified and approved doctors list fetched successfully",
		"data":    doctors,
	})
}

// View verified but not approved doctors
func ViewVerifiedNotApprovedDoctors(c *gin.Context) {
	var doctors []models.Doctor

	if err := configuration.DB.Where("verified = ? AND approved = ?", "true", "false").Find(&doctors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching verified and approved Doctors"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Verified and approved doctors list fetched successfully",
		"data":    doctors,
	})
}

// Update doctor credentials
func UpdateDoctor(c *gin.Context) {
	var doctor models.Doctor
	doctorID := c.Param("id")

	if err := configuration.DB.First(&doctor, doctorID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"Error": "No doctor with this ID"})
		return
	}

	if err := c.BindJSON(&doctor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if err := configuration.DB.Save(&doctor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	// Send email to the doctor with the updated details
	if err := SendEmailToDoctor(doctor); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to send email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"message": "Doctor detailes have been updated sucessfully sucessfully",
		"data":    doctor,
	})
}

func SendEmailToDoctor(doctor models.Doctor) error {
	// Construct email message with updated details
	subject := "Your details have been updated"
	body := fmt.Sprintf("Hello %s,\n\nYour details have been successfully updated.\n\nUpdated details:\nName: %s\nSpecialization: %s\nEmail: %s\nPhone: %s\nLicense Number: %s\nVerified: %s\nApproved: %s\n",
		doctor.Name, doctor.Name, doctor.Specialization, doctor.Email, doctor.Phone, doctor.LicenseNumber, doctor.Verified, doctor.Approved)

	// Set up SMTP authentication information
	SMTPemail := os.Getenv("Email")
	SMTPpass := os.Getenv("Password")
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Authenticate with SMTP server
	auth := smtp.PlainAuth("", SMTPemail, SMTPpass, smtpHost)

	// Compose email headers
	headers := make(map[string]string)
	headers["From"] = "Your Name <" + SMTPemail + ">"
	headers["To"] = doctor.Email
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/plain; charset=\"utf-8\""
	headers["Content-Transfer-Encoding"] = "base64"

	// Construct email message
	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)

	// Send email using SMTP server
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, SMTPemail, []string{doctor.Email}, []byte(msg.String()))
	if err != nil {
		log.Println("Error sending email:", err)
		return err
	}

	return nil
}

// View all Doctors list
func ViewDoctors(c *gin.Context) {
	var doctors []models.Doctor

	if err := configuration.DB.Find(&doctors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Doctors not found"})
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Doctors list fetehed successfully",
		"data":    doctors,
	})
}

// Get doctors details by id
func GetDoctorByID(c *gin.Context) {
	var doctor models.Doctor
	doctorID := c.Param("id")

	if err := configuration.DB.First(&doctor, doctorID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"Error": "Couldn't Get doctor details"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Doctor details fetched successfully",
		"data":    doctor,
	})

}

// Get doctros details by speciality
func GetDoctorBySpeciality(c *gin.Context) {
	var doctors []models.Doctor
	doctorSpeciality := c.Param("specialization")
	if err := configuration.DB.Where("specialization = ?", doctorSpeciality).Find(&doctors).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "No doctors found with the specified speciality"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error":   "Couldn't Get doctors details",
			"details": err.Error()})
		return
	}
	if len(doctors) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No doctors found with the specified speciality"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Doctors details list fetched successfully",
		"data":    doctors,
	})
}

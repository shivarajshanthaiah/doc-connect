package controllers

import (
	"doc-connect/configuration"
	"doc-connect/models"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetAppointmenentHistory(c *gin.Context) {
	var appointmentHistory []models.Appointment
	patientID := c.Param("id")

	if err := configuration.DB.Where("patient_id = ?", patientID).Find(&appointmentHistory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid patient id"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error":   "Couldn't Get doctors details",
			"details": err.Error()})
		return
	}
	if len(appointmentHistory) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No history found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Doctors details list fetched successfully",
		"data":    appointmentHistory,
	})
}

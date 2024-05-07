package controllers

import (
	"doc-connect/configuration"
	"doc-connect/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Viewing hospitals
func ViewHospitals(c *gin.Context) {
	var hospital []models.Hospital

	if err := configuration.DB.Find(&hospital).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Hospital not found"})
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Hospitals list fetehed successfully",
		"data":    hospital,
	})
}

// Adding hospitals
func AddHospital(c *gin.Context) {
	var hospital models.Hospital
	if err := c.BindJSON(&hospital); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if a hospital with the same name and location already exists
	var existingHospital models.Hospital
	if err := configuration.DB.Where("name = ? AND location = ?", hospital.Name, hospital.Location).First(&existingHospital).Error; err == nil {
		// Hospital with the same name and location already exists
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Hospital already exists",
			"message": "A hospital with the same name and location already exists",
		})
		return
	} else if err != gorm.ErrRecordNotFound {
		// Database error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	hospital.Status = "Active"
	if err := configuration.DB.Create(&hospital).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Hospital details are added successfully",
		"data":    hospital})

}

//Search hospitals

func SearchHospital(c *gin.Context) {
	var hospital models.Hospital

	hospitalID := c.Param("id")
	if err := configuration.DB.First(&hospital, hospitalID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"Error": "Staff not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Hospital details fetched sucessfully",
		"data":    hospital,
	})
}

// Update hospital
func UpdateHospital(c *gin.Context) {
	var hospital models.Hospital
	hospitalID := c.Param("id")

	if err := configuration.DB.First(&hospital, hospitalID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"Error": "Hospital not found"})
		return
	}
	if err := c.BindJSON(&hospital); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if err := configuration.DB.Save(&hospital).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"message": "Hospital details updated sucessfully",
		"data":    hospital,
	})
}

// Remove/delete hospital
func RemoveHospital(c *gin.Context) {
	var hospital models.Hospital

	hospitalID := c.Param("id")
	if err := configuration.DB.First(&hospital, hospitalID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"Status":  "Failed",
			"message": "Hospital id not found",
			"data":    err.Error(),
		})
		return
	}

	if err := c.BindJSON(&hospital); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hospital.Status = "Deactive"
	configuration.DB.Save(&hospital)
	c.JSON(http.StatusOK, gin.H{
		"Status":  "success",
		"message": "Hospital details removed successfully",
	})
}

// View deleted hospital
func ViewDeletedHospitals(c *gin.Context) {
	var hospitals []models.Hospital

	if err := configuration.DB.Where("status = ?", "Deactive").Find(&hospitals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching deleted hospitals"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Deleted Hospitals list fetched successfully",
		"Data":    hospitals,
	})
}

// View Active hospitals
func ViewActiveHospitals(c *gin.Context) {
	var hospitals []models.Hospital

	if err := configuration.DB.Where("status = ?", "Active").Find(&hospitals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching hospitals"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Active Hospitals": hospitals})
}

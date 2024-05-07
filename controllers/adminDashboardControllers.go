package controllers

import (
	"doc-connect/configuration"
	"doc-connect/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Func to get booking status
func GetBookingStatusCounts(c *gin.Context) {
	var totalBookings int64
	// Query the database to count the total number of bookings
	result := configuration.DB.Model(&models.Appointment{}).Count(&totalBookings)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to fetch total bookings"})
		return
	}

	// Query the database to count the number of confirmed bookings
	var confirmedBookings int64
	confirmedResults := configuration.DB.Model(&models.Appointment{}).Where("booking_status = ?", "confirmed").Count(&confirmedBookings)
	if confirmedResults.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to fetch confirmed bookings"})
		return
	}

	// Query the database to count the number of completed bookings
	var completedBookings int64
	completedResult := configuration.DB.Model(&models.Appointment{}).Where("booking_status = ?", "completed").Count(&completedBookings)
	if completedResult.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to fetch completed bookings"})
		return
	}

	// Query the database to count the number of cancelled bookings
	var cancelledBookings int64
	cancelledResult := configuration.DB.Model(&models.Appointment{}).Where("booking_status = ?", "cancelled").Count(&cancelledBookings)
	if cancelledResult.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to fetch cancelled bookings"})
		return
	}

	// Construct and send the JSON response with booking counts
	c.JSON(http.StatusOK, gin.H{
		"Status":            "Sucess",
		"Message":           "Booking details fetched sucessfully",
		"TotalBookings":     totalBookings,
		"ConfirmedBookings": confirmedBookings,
		"CompletedBookings": completedBookings,
		"CancelledBookings": cancelledBookings,
	})
}

// Func to get doctor-wise bookings
func GetDoctorWiseBookings(c *gin.Context) {
	// Defined a struct to store doctor-wise data
	var doctorData []struct {
		DoctorID     int     `json:"doctor_id"`
		BookingCount int     `json:"booking_count"`
		TotalRevenue float64 `json:"total_revenue"`
	}

	// Query the database to get doctor-wise data
	result := configuration.DB.Table("appointments").
		Select("appointments.doctor_id, COUNT(*) as booking_count, SUM(invoices.total_amount) as total_revenue").
		Joins("INNER JOIN invoices ON appointments.appointment_id = invoices.appointment_id").
		Group("appointments.doctor_id").
		Scan(&doctorData)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch doctor-wise data"})
		return
	}

	//Constructing and sending JSON response
	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"message":    "Doctor-wise data fetched successfully",
		"doctorData": doctorData,
	})
}

// Func to get department-wise bookings
func GetDepartmentWiseBookings(c *gin.Context) {
	// Defined a struct to store department-wise data
	var departmentData []struct {
		Specialization string  `json:"specialization"`
		BookingCount   int     `json:"booking_count"`
		TotalRevenue   float64 `json:"total_revenue"`
	}

	// Query the database to get doctor-wise data
	result := configuration.DB.Table("appointments").
		Select("doctors.specialization as specialization, COUNT(*) as booking_count, SUM(invoices.total_amount) as total_revenue").
		Joins("JOIN doctors ON appointments.doctor_id = doctors.doctor_id").
		Joins("JOIN invoices ON appointments.appointment_id = invoices.appointment_id").
		Group("doctors.specialization").
		Scan(&departmentData)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch department-wise data"})
		return
	}

	//Constructing and sending JSON response
	c.JSON(http.StatusOK, gin.H{
		"status":         "success",
		"message":        "Department-wise data fetched successfully",
		"departmentData": departmentData,
	})
}

// Defined a struct to store Revenue
type Revenue struct {
	Day   *float64 `json:"day"`
	Week  *float64 `json:"week"`
	Month *float64 `json:"month"`
	Year  *float64 `json:"year"`
}

// Func to get total revenue
func GetTotalRevenue(c *gin.Context) {
	now := time.Now()

	// Get the start and end time for the day
	startofDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endofDay := startofDay.AddDate(0, 0, 1).Add(-time.Second)

	// Get the start and end time for the week
	startofWeek := startofDay.AddDate(0, 0, -int(now.Weekday()))
	endofWeek := startofWeek.AddDate(0, 0, 7).Add(-time.Second)

	// Get the start and end time for the month
	startofMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endofMonth := startofMonth.AddDate(0, 1, 0).Add(-time.Second)

	// Get the start and end time for the year
	startofYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	endofYear := startofYear.AddDate(1, 0, 0).Add(-time.Second)

	// Query the database to get the total revenue for different timeframes
	var revenue Revenue
	result := configuration.DB.Model(&models.Invoice{}).
		Select("SUM(total_amount) as total_revenue").
		Where("payment_status = ?", "Paid").
		Where("updated_at BETWEEN ? AND ?", startofDay, endofDay).
		Scan(&revenue.Day)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to fetch revenue for the day"})
		return
	}

	// Fetching revenue for the week
	result = configuration.DB.Model(&models.Invoice{}).
		Select("SUM(total_amount) as total_revenue").
		Where("payment_status = ?", "Paid").
		Where("updated_at BETWEEN ? AND ?", startofWeek, endofWeek).
		Scan(&revenue.Week)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to fetch revenue for the week"})
		return
	}

	// Fetching revenue for the month
	result = configuration.DB.Model(&models.Invoice{}).
		Select("SUM(total_amount) as total_revenue").
		Where("payment_status = ?", "Paid").
		Where("updated_at BETWEEN ? AND ?", startofMonth, endofMonth).
		Scan(&revenue.Month)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to fetch revenue for the month"})
		return
	}

	// Fetching revenue for the year
	result = configuration.DB.Model(&models.Invoice{}).
		Select("SUM(total_amount) as total_revenue").
		Where("payment_status = ?", "Paid").
		Where("updated_at BETWEEN ? AND ?", startofYear, endofYear).
		Scan(&revenue.Year)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to fetch revenue for the year"})
		return
	}

	// Respond with the total revenue for different timeframes
	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"message": "Revenure details fetched sucessfully",
		"Revenue": revenue,
	})
}

// Defined a struct to store Revenue for the specified date range
type SpecificRevenue struct {
	Revenue *float64 `json:"revenue"`
}

func GetSpecificRevenue(c *gin.Context) {
	// Retrieving start and end date from query parameters
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error
	// Parse start date
	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid start date format. Use YYYY-MM-DD"})
			return
		}
	} else {
		startDate = time.Now() // If start date not provided, default to current date
	}

	// Parse end date
	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid end date format. Use YYYY-MM-DD"})
			return
		}
	} else {
		endDate = time.Now()
	}

	// Query the database to get the total revenue for the specified date range
	var specificRevenue SpecificRevenue
	result := configuration.DB.Model(&models.Invoice{}).
		Select("SUM(total_amount) as total_revenue").
		Where("payment_status = ?", "Paid").
		Where("updated_at BETWEEN ? AND ?", startDate, endDate).
		Scan(&specificRevenue.Revenue)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to fetch revenue for specific date range"})
		return
	}

	// Respond with the total revenue for the specified date range
	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Revenue details fetched successfully",
		"Revenue": specificRevenue,
	})
}

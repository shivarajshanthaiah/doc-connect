package controllers

import (
	"bytes"
	"doc-connect/configuration"
	"doc-connect/models"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"gorm.io/gorm"
)

// ViewHospital retrieves a list of active hospitals
func ViewHospital(c *gin.Context) {
	var hospitals []models.Hospital

	if err := configuration.DB.Where("status = ?", "Active").Find(&hospitals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Hospital not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Hospitals list fetehed successfully",
		"data":    hospitals,
	})
}

// DoctorLogout
func DoctorLogout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "You are successfully logged out"})
}

// SaveAvailability saves the availability of a doctor
func SaveAvailability(c *gin.Context) {
	var availability models.DoctorAvailability

	if err := c.BindJSON(&availability); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	doctorID, ok := c.Get("doctor_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Doctor not authenticated"})
		return
	}
	fmt.Println("herr", doctorID)
	availability.DoctorID = doctorID.(uint)

	// Check if doctor exists
	var doctor models.Doctor
	if err := configuration.DB.Where("doctor_id = ?", availability.DoctorID).First(&doctor).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Doctor id not found"})
		return
	}

	// Check if doctor is approved
	if doctor.Approved != "true" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Doctor not found"})
		return
	}

	// Check if availability for the given date already exists
	var existingAvailability models.DoctorAvailability
	if err := configuration.DB.Where("doctor_id = ? AND date = ?", availability.DoctorID, availability.Date).First(&existingAvailability).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Availability already exists for this date"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check availability"})
		return
	}

	// Create new availability record in the database
	if err := configuration.DB.Create(&availability).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create availability"})
		return
	}

	c.JSON(http.StatusOK, availability)
}

// AddPrescription
func AddPrescription(c *gin.Context) {
	var prescription models.Prescription
	if err := c.BindJSON(&prescription); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	doctorID, ok := c.Get("doctor_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Doctor not authenticated"})
		return
	}
	fmt.Println("herr", doctorID)
	prescription.DoctorID = doctorID.(uint)

	// Check if doctor exists
	var doctor models.Doctor
	if err := configuration.DB.Where("doctor_id = ?", doctorID).First(&doctor).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid doctor ID"})
		return
	}

	// Check if patient exists
	var patient models.Patient
	if err := configuration.DB.Where("patient_id = ?", prescription.PatientID).First(&patient).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid patient ID"})
		return
	}

	// Check if appointment exists for the doctor and patient
	var appointment models.Appointment
	if err := configuration.DB.Where("doctor_id = ? AND patient_id = ? AND appointment_id = ?", doctorID, prescription.PatientID, prescription.AppointmentID).First(&appointment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No confirmed appointment found for the doctor and patient"})
		return
	}

	switch appointment.BookingStatus {
	case "pending":
		c.JSON(http.StatusBadRequest, gin.H{"error": "Appointment is not confirmed"})
		return
	case "completed":
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prescription already added for this appointment"})
		return
	case "cancelled":
		c.JSON(http.StatusBadRequest, gin.H{"error": "Appointment has been cancelled"})
		return
	}

	// Create new prescription record in the database
	if err := configuration.DB.Create(&prescription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add prescription"})
		return
	}

	// Update appointment status to completed
	if err := configuration.DB.Model(&appointment).Update("booking_status", "completed").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update appointment status"})
		return
	}

	// Generate PDF invoice
	pdfInvoice, err := GeneratePrescriptionPDF(appointment, doctor, patient, prescription)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF invoice"})
		return
	}

	// // Send payment confirmation email with PDF invoice attached
	err = SendPrescriptionEmail("Prescription attachment", patient.Email, "invoice.pdf", pdfInvoice)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to send email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":      "Success",
		"Message":     "Prescription added sucessfully",
		"pescription": prescription,
	})
}

func GetAppHistory(c *gin.Context) {
	var appointment []models.Appointment
	doctorID := c.Param("id")

	if err := configuration.DB.Where("doctor_id = ?", doctorID).Find(&appointment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid doctor id"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error":   "Couldn't Get doctors details",
			"details": err.Error()})
		return
	}
	if len(appointment) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No history found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Status":  "Success",
		"Message": "Doctors details list fetched successfully",
		"data":    appointment,
	})
}

func GetDoctorAppointmentsByDate(c *gin.Context) {
	doctorID := c.Param("doctor_id")
	dateStr := c.Query("date")

	// Parse the date string into time.Time format
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	// Query appointments
	var appointments []models.Appointment
	if err := configuration.DB.Where("doctor_id = ? AND appointment_date = ? AND booking_status = ?", doctorID, date, "confirmed").Find(&appointments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch appointments"})
		return
	}

	// Extract booked time slots
	bookedTimeSlots := make(map[string]bool)
	for _, appointment := range appointments {
		bookedTimeSlots[appointment.AppointmentTimeSlot] = true
	}

	// Respond with booked time slots
	c.JSON(http.StatusOK, bookedTimeSlots)
}

// Generates a professional PDF prescription
func GeneratePrescriptionPDF(appointment models.Appointment, doctor models.Doctor, patient models.Patient, prescription models.Prescription) ([]byte, error) {
	// Initialize PDF document
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.AddPage()

	// Set font and font size
	pdf.SetFont("Arial", "B", 14)

	// Title (Doctor Prescription)
	pdf.SetTextColor(0, 0, 0) // Black color
	pdf.CellFormat(0, 10, "Doctor Prescription", "", 1, "C", false, 0, "")

	// Doctor details section
	pdf.SetFont("Arial", "B", 12)
	add1Detail(pdf, "Doctor Name:", doctor.Name, true)
	add1Detail(pdf, "Specialization:", doctor.Specialization, false)

	// Patient details section
	pdf.SetFont("Arial", "B", 12)
	pdf.SetY(pdf.GetY() + 10) // Move down
	add1Detail(pdf, "Patient Name:", patient.Name, true)
	add1Detail(pdf, "Age:", patient.Age, false)
	add1Detail(pdf, "Gender:", patient.Gender, false)

	// Appointment details section
	pdf.SetFont("Arial", "B", 12)
	pdf.SetY(pdf.GetY() + 10) // Move down
	add1Detail(pdf, "Appointment Date:", appointment.AppointmentDate.Format("2006-01-02"), true)
	add1Detail(pdf, "Time Slot:", appointment.AppointmentTimeSlot, false)

	// Prescription details section
	pdf.SetFont("Arial", "B", 12)
	pdf.SetY(pdf.GetY() + 10) // Move down
	add1Detail(pdf, "Prescription ID:", fmt.Sprintf("%d", prescription.ID), true)
	add1Detail(pdf, "Instructions:", prescription.PrescriptionText, false)

	// Prescription note
	pdf.SetFont("Arial", "", 10)
	pdf.SetY(pdf.GetY() + 10) // Move down
	pdf.MultiCell(0, 5, "Follow the instructions given by the doctor properly. Your health is all that matters!", "", "C", false)

	// Output PDF to buffer
	var pdfBuffer bytes.Buffer
	err := pdf.Output(&pdfBuffer)
	if err != nil {
		return nil, err
	}

	return pdfBuffer.Bytes(), nil
}

// addDetail adds a detail line to the PDF
func add1Detail(pdf *gofpdf.Fpdf, label, value string, isHeader bool) {
	if isHeader {
		pdf.SetFont("Arial", "B", 12)
	} else {
		pdf.SetFont("Arial", "", 12)
	}
	pdf.CellFormat(0, 10, label, "", 1, "", false, 0, "")
	pdf.CellFormat(0, 10, value, "", 1, "", false, 0, "")
}

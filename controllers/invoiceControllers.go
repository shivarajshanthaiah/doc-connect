package controllers

import (
	"bytes"
	"doc-connect/configuration"
	"doc-connect/models"
	"errors"
	"fmt"

	// "fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
	"github.com/razorpay/razorpay-go"
	"gorm.io/gorm"
)

func GetInvoice(c *gin.Context) {
	var invoice []models.Invoice
	if err := configuration.DB.Find(&invoice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error occured while receiving the invoice",
		})
		return
	}
	c.JSON(http.StatusOK, invoice)
}

// To make payment offline
func PayInvoiceOffline(c *gin.Context) {
	// Struct to hold the payment request parameters
	var paymentRequest struct {
		InvoiceID uint `json:"invoice_id"`
	}

	if err := c.BindJSON(&paymentRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	var invoice models.Invoice
	if err := configuration.DB.Where("invoice_id = ?", paymentRequest.InvoiceID).First(&invoice).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
		return
	}

	if invoice.PaymentStatus == "Paid" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invoice already paid"})
		return
	}

	// Update payment status to "Paid" for offline payment
	invoice.PaymentStatus = "Paid"
	invoice.PaymentMethod = "Offline"
	if err := configuration.DB.Save(&invoice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment status"})
		return
	}

	// Update corresponding appointment status
	var appointment models.Appointment
	if err := configuration.DB.Where("appointment_id = ?", invoice.AppointmentID).First(&appointment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch appointment"})
		return
	}

	// Update appointment status to "confirmed" and payment status to "paid"
	appointment.BookingStatus = "confirmed"
	appointment.PaymentStatus = "paid"
	if err := configuration.DB.Save(&appointment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update appointment status"})
		return
	}

	// Fetch doctor and patient details based on the booking
	var doctor models.Doctor
	if err := configuration.DB.First(&doctor, appointment.DoctorID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch doctor details"})
		return
	}

	var patient models.Patient
	if err := configuration.DB.First(&patient, appointment.PatientID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch patient details"})
		return
	}

	// Generate PDF invoice
	pdfInvoice, err := GeneratePaidPDFInvoice(appointment, invoice, doctor, patient)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF invoice"})
		return
	}

	// Send payment confirmation email with PDF invoice attached
	err = SendInvoiceEmail("Payment successful for invoice", appointment.PatientEmail, "invoice.pdf", pdfInvoice)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Invoice payment successful",
		"invoice": invoice,
	})
}

// PageVariable struct holds data to be passed to the HTML template.
type PageVariable struct {
	AppointmentID string
}

// Function for processing online payments
func MakePaymentOnline(c *gin.Context) {

	invoiceID := c.Query("id")
	// Convert the invoice ID from string to integer
	id, err := strconv.Atoi(invoiceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
	}

	// Retrieve the invoice corresponding to the provided ID from the database
	var invoice models.Invoice
	if err := configuration.DB.First(&invoice, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "Failed",
				"message": "Invoice Not Found",
				"data":    err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch the invoice",
		})
		return
	}

	// Check if the invoice is already paid
	if invoice.PaymentStatus == "Paid" {
		c.JSON(400, gin.H{"error": "Invoice is already paid"})
		return
	}

	// Create a RazorPay payment record in the database with the invoice ID and total amount.
	razorpayment := &models.RazorPay{
		InvoiceID:  uint(invoice.InvoiceID),
		AmountPaid: invoice.TotalAmount,
	}

	razorpayment.RazorPaymentID = generateUniqueID()
	if err := configuration.DB.Create(&razorpayment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create razor payment"})
		return
	}

	// Convert total amount to paisa (multiply by 100) for RazorPay API
	amountInPaisa := invoice.TotalAmount * 100
	razorpayClient := razorpay.NewClient(os.Getenv("RazorPay_key_id"), os.Getenv("RazorPay_key_secret"))

	// Prepare data for creating a RazorPay order.
	data := map[string]interface{}{
		"amount":   amountInPaisa,
		"currency": "INR",
		"receipt":  "some_receipt_id",
	}

	// Create a RazorPay order using the RazorPay API
	body, err := razorpayClient.Order.Create(data, nil)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Faied to create razorpay orer"})
	}

	// Extract the order ID from the response body returned by the RazorPay API
	value := body["id"]
	str := value.(string)

	// Create an instance of the PageVariable struct to hold data for the HTML template
	homepagevariables := PageVariable{
		AppointmentID: str,
	}

	// Render the payment.html template, passing invoice ID, total price, total amount, and appointment ID as template variables.
	c.HTML(http.StatusOK, "payment.html", gin.H{
		"invoiceID":     id,
		"totalPrice":    amountInPaisa / 100,
		"total":         amountInPaisa,
		"appointmentID": homepagevariables.AppointmentID,
	})
}

// generateUniqueID generates a unique ID using UUID (Universally Unique Identifier).
func generateUniqueID() string {
	// Generate a Version 4 (random) UUID
	id := uuid.New()
	return id.String()
}

// Function to display success page after successfull payment
func SuccessPage(c *gin.Context) {
	paymentID := c.Query("bookID")
	fmt.Println(paymentID)

	// Fetch the invoice corresponding to the provided payment ID from the database
	var invoice models.Invoice
	if err := configuration.DB.First(&invoice, paymentID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": "Failed to fetch the invoice",
		})
		return
	}
	fmt.Printf("%+v\n", invoice)

	if invoice.PaymentStatus == "Pending" {
		if err := configuration.DB.Model(&invoice).Update("payment_status", "Paid").Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Error": "Failed to update the invoice paymet status",
			})
			return
		}
	}

	// Update payment method to "online"
	if err := configuration.DB.Model(&invoice).Update("payment_method", "online").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": "Failed to update the payment method",
		})
		return
	}

	// Create a record of the RazorPay payment in the database
	razorPayment := models.RazorPay{
		InvoiceID:      uint(invoice.InvoiceID),
		RazorPaymentID: generateUniqueID(),
		AmountPaid:     invoice.TotalAmount,
	}

	if err := configuration.DB.Create(&razorPayment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": "Failed to create RazorPay Payment",
		})
	}

	// Update appointment status in appointment table
	if invoice.AppointmentID != 0 {
		if err := configuration.DB.Model(&models.Appointment{}).Where("appointment_id = ?", invoice.AppointmentID).Updates(map[string]interface{}{"booking_status": "confirmed", "payment_status": "paid"}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Error": "Failed to update the appointment status",
			})
			return
		}
	}

	// bookingID := c.Query("appointmentID")
	// fmt.Println("bookidi",bookingID)
	var booking models.Appointment
	if err := configuration.DB.First(&booking, invoice.AppointmentID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch booking details"})
		return
	}

	// Fetch doctor and patient details based on the booking
	var doctor models.Doctor
	if err := configuration.DB.First(&doctor, booking.DoctorID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch doctor details"})
		return
	}

	var patient models.Patient
	if err := configuration.DB.First(&patient, booking.PatientID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch patient details"})
		return
	}

	// Generate PDF invoice
	pdfInvoice, err := GeneratePaidPDFInvoice(booking, invoice, doctor, patient)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF invoice"})
		return
	}

	//Update corresponding appointment status
	var appointment models.Appointment
	if err := configuration.DB.Where("appointment_id = ?", invoice.AppointmentID).First(&appointment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch appointment"})
		return
	}

	// Send payment confirmation email with PDF invoice attached
	err = SendInvoiceEmail("Payment successful for invoice", appointment.PatientEmail, "invoice.pdf", pdfInvoice)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email"})
		return
	}

	// Render the success page template, passing payment ID, amount paid, and invoice ID as template variables
	c.HTML(http.StatusOK, "success.html", gin.H{
		"paymentID":  razorPayment.RazorPaymentID,
		"amountPaid": invoice.TotalAmount,
		"invoiceID":  invoice.InvoiceID,
	})
}

// generateDuePDFInvoice generates a professional PDF invoice for appointment dues
func GeneratePaidPDFInvoice(booking models.Appointment, invoice models.Invoice, doctor models.Doctor, patient models.Patient) ([]byte, error) {
	// Initialize PDF document
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.AddPage()

	// Set font and font size
	pdf.SetFont("Arial", "B", 14)

	// Title (Go - Doctor Appointment Booking)
	pdf.SetTextColor(128, 0, 128) // Dark purple color
	pdf.CellFormat(0, 10, "Go - Doctor Appointment Booking", "", 1, "C", false, 0, "")

	// Business details (GSTN: www.goworld.com)
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 7, "GSTN: www.goworld.com", "", 1, "C", false, 0, "")

	// Appointment details section
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(0, 0, 0) // Black color
	pdf.CellFormat(0, 10, "Invoice", "1", 1, "C", false, 0, "")
	add2Detail(pdf, "Invoice ID", fmt.Sprintf("%d", invoice.InvoiceID), true)
	add2Detail(pdf, "Doctor Name", doctor.Name, true)
	add2Detail(pdf, "Specialization", doctor.Specialization, true)
	add2Detail(pdf, "Patient Name", patient.Name, true)
	add2Detail(pdf, "Appointment ID", fmt.Sprintf("%d", booking.AppointmentID), true)
	add2Detail(pdf, "Appointment Date", booking.AppointmentDate.Format("2006-01-02"), true)
	add2Detail(pdf, "Time Slot", booking.AppointmentTimeSlot, true)

	// Invoice details section
	pdf.CellFormat(0, 10, "Invoice Details", "1", 1, "C", false, 0, "")
	add2Detail(pdf, "Booking Status", booking.BookingStatus, false)
	add2Detail(pdf, "Due date", invoice.PaymentDueDate.Format("2006-01-02"), false)
	add2Detail(pdf, "Paid date", invoice.UpdatedAt.Format("2006-01-02"), false)
	// CGST and SGST (0)
	add2Detail(pdf, "CGST (0%)", "0.00", false)
	add2Detail(pdf, "SGST (0%)", "0.00", false)
	pdf.SetFont("Arial", "B", 13)
	add2Detail(pdf, "Grand Total", fmt.Sprintf("%.2f", invoice.TotalAmount), true)
	pdf.SetTextColor(139, 128, 0) // Yellow color for total amount
	add2Detail(pdf, "Amount Paid", fmt.Sprintf("%.2f", invoice.TotalAmount), true)

	// Payment instructions
	pdf.SetTextColor(0, 0, 0) // Reset text color to black
	//pdf.CellFormat(0, 10, "Payment Instructions:", "", 1, "L", false, 0, "")
	pdf.MultiCell(0, 5, "Thank you for using our service.", "", "L", false)

	// Seal and signature section
	pdf.SetY(pdf.GetY() + 12) // Move down for seal and signature
	pdf.CellFormat(0, 10, "This is a computer generated invoice", "", 1, "R", false, 0, "")

	// Output PDF to buffer
	var pdfBuffer bytes.Buffer
	err := pdf.Output(&pdfBuffer)
	if err != nil {
		return nil, err
	}

	return pdfBuffer.Bytes(), nil
}

// addDetail adds a detail line to the PDF
func add2Detail(pdf *gofpdf.Fpdf, label, value string, isHeader bool) {
	if isHeader {
		pdf.SetFont("Arial", "B", 12)
		pdf.SetFillColor(255, 255, 255) // White color for header background
	} else {
		pdf.SetFont("Arial", "", 10)
		pdf.SetFillColor(240, 240, 240) // Light gray color for detail background
	}
	pdf.CellFormat(45, 10, label, "1", 0, "", false, 0, "")
	pdf.CellFormat(0, 10, value, "1", 1, "", false, 0, "")
}

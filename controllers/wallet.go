package controllers

import (
	"doc-connect/configuration"
	"doc-connect/models"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetUserWallet helps to get user wallet by user id
func Wallet(c *gin.Context) {
	userid := c.Param("userid")

	var wallet models.Wallet
	if err := configuration.DB.Where("user_id = ?", userid).First(&wallet).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "failed to find user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status":        "Success",
		"Wallet Amount": wallet.Amount,
	})

}

// CancelAppointment is a handler function for cancelling an appointment.
func CancelAppointment(c *gin.Context) {
	var appointment models.Appointment
	if err := configuration.DB.Where("appointment_id = ?", c.Param("id")).First(&appointment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
		return
	}

	if appointment.BookingStatus == "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Appointment has already been cancelled"})
		return
	}

	if appointment.BookingStatus == "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "This appointment has already been completed"})
		return
	}

	if appointment.BookingStatus != "confirmed" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Appointment cannot be cancelled as it is not confirmed"})
		return
	}

	var invoice models.Invoice
	if err := configuration.DB.Where("appointment_id = ?", c.Param("id")).First(&invoice).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
		return
	}

	if invoice.PaymentMethod == "online" {
		// Refund applicable for online payments
		refundAmount := invoice.TotalAmount * 0.95

		// Update payment status to refunded
		invoice.PaymentStatus = "refunded"

		// Update invoice in the database
		if err := configuration.DB.Save(&invoice).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to update invoice"})
			return
		}

		var wallet models.Wallet
		if err := configuration.DB.Where("user_id = ?", appointment.PatientID).First(&wallet).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Wallet doesn't exist, create a new one
				wallet = models.Wallet{
					UserID: appointment.PatientID,
					Amount: refundAmount,
				}
				if err := configuration.DB.Create(&wallet).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to create wallet"})
					return
				}
			} else {
				// Error occurred while fetching wallet
				c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to fetch wallet"})
				return
			}
		} else {
			// Wallet exists, update its amount
			wallet.Amount += refundAmount
			if err := configuration.DB.Save(&wallet).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to update wallet"})
				return
			}
		}

		appointment.BookingStatus = "cancelled"
		if err := configuration.DB.Save(&appointment).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to update appointment status"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("Appointment Cancelled. Refund amount: %.2f", refundAmount),
		})
	} else {
		// For offline payments, simply cancel the appointment without refunding
		appointment.BookingStatus = "cancelled"

		// Update appointment status in the database
		if err := configuration.DB.Save(&appointment).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to update appointment status"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Appointment Cancelled. Amount cannot be refunded as payment method was not online"})
	}
}

func PayFromWallet(c *gin.Context) {
	// Parse the JSON request body to extract the invoice ID
    var paymentRequest struct {
        InvoiceID uint `json:"invoice_id"`
    }

    if err := c.BindJSON(&paymentRequest); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
        return
    }

	// Fetch the invoice from the database based on the provided invoice ID
    var invoice models.Invoice
    if err := configuration.DB.Where("invoice_id = ?", paymentRequest.InvoiceID).First(&invoice).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
        return
    }

	// Check if the invoice has already been paid
    if invoice.PaymentStatus == "Paid" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invoice already paid"})
        return
    }

	// Fetch the wallet details of the patient from the database
    var wallet models.Wallet
    if err := configuration.DB.Where("user_id = ?", invoice.PatientID).First(&wallet).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch wallet details"})
        return
    }

	// Check if the wallet balance is sufficient to pay the invoice
    if wallet.Amount < invoice.TotalAmount {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance in wallet"})
        return
    }

    // Start a database transaction
    tx := configuration.DB.Begin()

    // Defer rollback if transaction fails
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    // Update invoice payment status
    invoice.PaymentStatus = "Paid"
    invoice.PaymentMethod = "online"
    if err := tx.Save(&invoice).Error; err != nil {
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment status"})
        return
    }

    // Deduct payment amount from wallet balance
    wallet.Amount -= invoice.TotalAmount
    if err := tx.Model(&models.Wallet{}).Where("user_id = ?", wallet.UserID).Update("amount", wallet.Amount).Error; err != nil {
        tx.Rollback()
		log.Println("Failed to update wallet balance:", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update wallet balance"})
        return
    }

    // Update appointment status
    var appointment models.Appointment
    if err := tx.Where("appointment_id = ?", invoice.AppointmentID).First(&appointment).Error; err != nil {
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch appointment"})
        return
    }

    appointment.BookingStatus = "confirmed"
    appointment.PaymentStatus = "paid"
    if err := tx.Save(&appointment).Error; err != nil {
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update appointment status"})
        return
    }

    // Commit the transaction
    tx.Commit()

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

    c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Payment from wallet successful"})
}

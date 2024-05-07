package models

import "time"

type Invoice struct {
	InvoiceID      uint      `gorm:"primaryKey"`
	DoctorID       uint      `gorm:"not null"`
	PatientID      uint      `gorm:"not null"`
	AppointmentID  uint      `gorm:"not null"`
	TotalAmount    float64    `gorm:"not null"`
	PaymentMethod  string    `json:"payment_method"`
	PaymentStatus  string    `gorm:"not null"`
	PaymentDueDate time.Time `gorm:"not null"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`
}

type RazorPay struct {
	RazorPaymentID  string `json:"razorpaymentID" gorm:"primaryKey;autoIncrement"`
	RazorPayorderID string `json:"razorpayorderID"`
	InvoiceID       uint   `json:"invoice_id"`
	AmountPaid      float64 `json:"amount_paid"`
}

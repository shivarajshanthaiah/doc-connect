package models

import "time"

type Appointment struct {
	AppointmentID       int       `gorm:"primaryKey"`
	PatientID           int       `json:"patient_id"`
	DoctorID            int       `json:"doctor_id"`
	PatientEmail        string    `json:"email"`
	AppointmentDate     time.Time `json:"appointment_date"`
	AppointmentTimeSlot string    `json:"appointment_time"`
	PatientHealthIssue  string    `json:"patient_health_issue"`
	PaymentStatus       string    `json:"payment_status"`
	BookingStatus       string    `json:"booking_status"`
}

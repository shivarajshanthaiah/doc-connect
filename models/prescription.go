package models

import (
	"gorm.io/gorm"
)

type Prescription struct {
	gorm.Model
	DoctorID         uint    `json:"doctor_id"`
	PatientID        uint    `json:"patient_id"`
	AppointmentID    uint    `json:"appointment_id"`
	HealthIssue      string `json:"health_issue"`
	PrescriptionText string `json:"prescription_text"`
}

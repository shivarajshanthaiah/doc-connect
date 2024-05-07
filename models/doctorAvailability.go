package models

import "time"

type DoctorAvailability struct {
	ID           uint      `gorm:"primaryKey"`
	DoctorID     uint      `json:"doctor_id"`
	Date         time.Time `json:"date"`
	AvilableTime string    `json:"available_time"`
}

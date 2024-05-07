package models

import "github.com/golang-jwt/jwt/v5"

type Doctor struct {
	DoctorID          uint   `gorm:"primaryKey"`
	Name              string `json:"name" gorm:"not null"`
	Age               int    `json:"age"`
	Gender            string `json:"gender" gorm:"not null"`
	Specialization    string `json:"specialization" gorm:"not null"`
	Experience        int    `json:"experience" gorm:"not null"`
	Email             string `json:"email" gorm:"unique"`
	Password          string `json:"password" gorm:"not null"`
	Phone             string `json:"phone" gorm:"not null"`
	LicenseNumber     string `json:"license_number" gorm:"not null"`
	ConsultancyCharge uint32 `json:"consultancy_charge"`
	Verified          string `json:"verified"`
	Approved          string `json:"approved"`
	HospitalID        uint   `json:"hospital_id" gorm:"not null"`
	Availabilities    []DoctorAvailability
}

type DoctorClaims struct {
	Id          uint   `json:"id"`
	DoctorEmail string `json:"email"`
	jwt.RegisteredClaims
}

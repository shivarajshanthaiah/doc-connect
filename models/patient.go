package models

import "github.com/dgrijalva/jwt-go"

type Patient struct {
	PatientID int    `gorm:"primaryKey"`
	Name      string `json:"name"`
	Age       string `json:"age"`
	Gender    string `json:"gender"`
	Phone     string `json:"phone" validate:"required"`
	Email     string `json:"email"`
	Address   string `json:"address"`
	Password  string `json:"password"`
}

type VerifyOTP struct {
	Phone string `json:"phone"`
	Otp   string `json:"otp"`
}

type PatientClaims struct {
	jwt.StandardClaims
	PatientID int `json:"patientID"`
	Phone string `json:"phone"`
}

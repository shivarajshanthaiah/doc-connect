package configuration

import (
	"doc-connect/models"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// hold connectioin to db
var DB *gorm.DB

// initializing db connection
func ConfigDB() {

	err1 := godotenv.Load(".env")
	if err1 != nil {
		log.Fatal("Error loading .env file")
	}
	dsn := os.Getenv("DB")
	var err error

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to the database")
	}

	DB.AutoMigrate(
		&models.Appointment{},
		&models.Doctor{},
		&models.Hospital{},
		&models.Patient{},
		&models.Invoice{},
		&models.RazorPay{},
		&models.Prescription{},
		&models.Admin{},
		&models.DoctorAvailability{},
		&models.Wallet{},
	)

}

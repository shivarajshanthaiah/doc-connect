package authentication

import (
	"log"
	"math/rand"
	"net/smtp"
	"os"
	"time"
)

// GenerateOTP
func GenerateOTP(length int) string {
	// Initializing the new random number generator
	rand.NewSource(time.Now().UnixNano())
	characters := "0123456789"
	// Create a byte slice to hold the OTP of the specified length.
	otp := make([]byte, length)

	for i := range otp {
		otp[i] = characters[rand.Intn(len(characters))]
	}
	return string(otp)
}


// SendOTPByEmail
func SendOTPByEmail(otp, email string) error {
	
	// Constructing the email message with the OTP included
	message := "Subject: WebPortal OTP\nHey Your OTP is " + otp

	SMTPemail := os.Getenv("Email")
	SMTPpass := os.Getenv("Password")

	// Authenticating with the SMTP server using the sender's credentials
	auth := smtp.PlainAuth("", SMTPemail, SMTPpass, "smtp.gmail.com")


	// Sending the email to the specified recipient's address
	err := smtp.SendMail("smtp.gmail.com:587", auth, SMTPemail, []string{email}, []byte(message))
	if err != nil {
		log.Println("Error sending email:", err)
		return err
	}

	return nil
}

// ValidateOTP
func ValidateOTP(otp, doctorOTP string) bool {
	return otp == doctorOTP
}

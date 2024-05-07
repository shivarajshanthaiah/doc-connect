package controllers

import (
	"fmt"
	"io"
	"os"

	"github.com/go-gomail/gomail"
)

// // SendEmail sends an email with an optional attachment
func SendEmail(msg, email, attachmentName string, attachmentData []byte) error {
	// SMTP server configuration
	senderEmail := os.Getenv("Email")
	senderPassword := os.Getenv("Password")

	// Compose email message
	m := gomail.NewMessage()
	m.SetHeader("From", senderEmail)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Payment Due mail")
	m.SetBody("text/plain", msg)

	// Add attachment
	m.Attach(attachmentName, gomail.SetCopyFunc(func(w io.Writer) error {
		// _, err := buf.WriteTo(w)
		_, err := w.Write(attachmentData)
		return err
	}))

	// Dial to SMTP server and send email
	d := gomail.NewDialer("smtp.gmail.com", 587, senderEmail, senderPassword)
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("error sending email: %v", err)
	}

	return nil
}

// // SendEmail sends an email with an optional attachment
func SendInvoiceEmail(msg, email, attachmentName string, attachmentData []byte) error {
	// SMTP server configuration
	senderEmail := os.Getenv("Email")
	senderPassword := os.Getenv("Password")

	// Compose email message
	m := gomail.NewMessage()
	m.SetHeader("From", senderEmail)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Payment Confimation mail")
	m.SetBody("text/plain", msg)

	// Add attachment
	m.Attach(attachmentName, gomail.SetCopyFunc(func(w io.Writer) error {
		// _, err := buf.WriteTo(w)
		_, err := w.Write(attachmentData)
		return err
	}))

	// Dial to SMTP server and send email
	d := gomail.NewDialer("smtp.gmail.com", 587, senderEmail, senderPassword)
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("error sending email: %v", err)
	}

	return nil
}

// // SendEmail sends an email with an optional attachment
func SendPrescriptionEmail(msg, email, attachmentName string, attachmentData []byte) error {
	// SMTP server configuration
	senderEmail := os.Getenv("Email")
	senderPassword := os.Getenv("Password")

	// Compose email message
	m := gomail.NewMessage()
	m.SetHeader("From", senderEmail)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Prescription e-mail")
	m.SetBody("text/plain", msg)

	// Add attachment
	m.Attach(attachmentName, gomail.SetCopyFunc(func(w io.Writer) error {
		// _, err := buf.WriteTo(w)
		_, err := w.Write(attachmentData)
		return err
	}))

	// Dial to SMTP server and send email
	d := gomail.NewDialer("smtp.gmail.com", 587, senderEmail, senderPassword)
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("error sending email: %v", err)
	}

	return nil
}

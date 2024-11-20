package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendMail(to string, subject string, body string) error {
	from := "rustygophers@gmail.com"
	password := os.Getenv("GMAIL_PASSWORD") // Ensure this is set correctly
	fmt.Println("Using password:", password)
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Properly format the email message to support HTML
	message := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" + // Set content type to HTML
		"\r\n" + // This empty line separates the headers from the body
		body + "\r\n")

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, message)
	if err != nil {
		fmt.Println("Email sending error:", err)
		return err
	}
	return nil
}

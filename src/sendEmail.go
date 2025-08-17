package main

import (
	"fmt"
	"net/smtp"
)

func sendEmail(toEmail string, subject string, body string) error {
	// SMTP configuration
	smtpHost := "localhost"
	smtpPort := "25"
	fromEmail := "noreply@cannan.dev"

	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n%s",
		fromEmail, toEmail, subject, body)

	// Send email
	err := smtp.SendMail(
		smtpHost+":"+smtpPort,
		nil, //no auth for localhost
		fromEmail,
		[]string{toEmail},
		[]byte(message),
	)
	return err
}

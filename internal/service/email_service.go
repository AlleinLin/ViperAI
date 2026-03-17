package service

import (
	"log"

	"viperai/internal/config"

	"gopkg.in/gomail.v2"
)

func sendEmail(to, code, message string) error {
	cfg := config.Get().Mail

	m := gomail.NewMessage()
	m.SetHeader("From", cfg.Address)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "ViperAI Notification")
	m.SetBody("text/plain", message+code)

	d := gomail.NewDialer(cfg.SMTPServer, cfg.SMTPPort, cfg.Address, cfg.AuthCode)

	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed to send email: %v", err)
		return err
	}

	return nil
}

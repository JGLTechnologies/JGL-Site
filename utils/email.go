package utils

import (
	"crypto/tls"
	"errors"
	"fmt"
	"html"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
	"time"
)

const (
	gmailSMTPHost    = "smtp.gmail.com"
	gmailSMTPAddress = gmailSMTPHost + ":587"
	smtpTimeout      = 10 * time.Second
)

func SendEmail(recipients []string, subject, body string) error {
	from := strings.TrimSpace(os.Getenv("GMAIL_SMTP_EMAIL"))
	password := os.Getenv("GMAIL_SMTP_APP_PASSWORD")

	if from == "" {
		return errors.New("GMAIL_SMTP_EMAIL environment variable is not set")
	}
	if password == "" {
		return errors.New("GMAIL_SMTP_APP_PASSWORD environment variable is not set")
	}
	if len(recipients) == 0 {
		return errors.New("at least one recipient is required")
	}

	if _, err := mail.ParseAddress(from); err != nil {
		return fmt.Errorf("invalid GMAIL_SMTP_EMAIL: %w", err)
	}

	to := make([]string, 0, len(recipients))
	for _, recipient := range recipients {
		recipient = strings.TrimSpace(recipient)
		address, err := mail.ParseAddress(recipient)
		if err != nil {
			return fmt.Errorf("invalid recipient %q: %w", recipient, err)
		}
		to = append(to, address.Address)
	}

	message := buildAnnouncementEmail("info@jgltechnologies.com", subject, body)

	auth := smtp.PlainAuth("", from, password, gmailSMTPHost)
	if err := sendMailWithTimeout(auth, "info@jgltechnologies.com", to, message); err != nil {
		return fmt.Errorf("send email through Gmail SMTP: %w", err)
	}

	return nil
}

func sendMailWithTimeout(auth smtp.Auth, from string, recipients []string, message []byte) error {
	connection, err := net.DialTimeout("tcp", gmailSMTPAddress, smtpTimeout)
	if err != nil {
		return err
	}
	defer connection.Close()
	if err := connection.SetDeadline(time.Now().Add(smtpTimeout)); err != nil {
		return err
	}

	client, err := smtp.NewClient(connection, gmailSMTPHost)
	if err != nil {
		return err
	}
	defer client.Close()
	if err := client.StartTLS(&tls.Config{ServerName: gmailSMTPHost, MinVersion: tls.VersionTLS12}); err != nil {
		return err
	}
	if err := client.Auth(auth); err != nil {
		return err
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	for _, recipient := range recipients {
		if err := client.Rcpt(recipient); err != nil {
			return err
		}
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(message); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func buildAnnouncementEmail(from, subject, body string) []byte {
	message := fmt.Sprintf(`From: JGL Technologies <%s>
To: undisclosed-recipients:;
Subject: JGL Announcement
MIME-Version: 1.0
Content-Type: text/html; charset=UTF-8

<!doctype html>
<html lang="en">
<body style="margin:0;padding:0;background:#f5f6f8;font-family:Arial,Helvetica,sans-serif;color:#111827;">
  <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="background:#f5f6f8;padding:40px 16px;">
    <tr><td align="center">
      <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="max-width:640px;background:#ffffff;border:1px solid #e5e7eb;border-radius:12px;box-shadow:0 4px 16px rgba(17,24,39,.06);">
        <tr><td style="padding:32px 36px 26px;background:#eef2ff;border-bottom:1px solid #dfe3f5;border-radius:12px 12px 0 0;">
          <img src="https://jgltechnologies.com/logo.png" width="240" alt="JGL Technologies" style="display:block;width:240px;max-width:100%%;height:auto;border:0;">
        </td></tr>
        <tr><td style="padding:34px 36px 40px;">
          <h1 style="margin:0 0 22px;font-size:30px;line-height:1.25;font-weight:700;color:#111827;">%s</h1>
          <div style="font-size:16px;line-height:1.75;color:#4b5563;white-space:pre-wrap;">%s</div>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`,
		from,
		html.EscapeString(subject),
		html.EscapeString(body),
	)

	return []byte(strings.ReplaceAll(message, "\n", "\r\n"))
}

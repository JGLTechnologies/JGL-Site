package utils

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"html/template"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
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

var announcementEmailTemplate = template.Must(template.New("announcement-email").Parse(`<!doctype html>
<html lang="en">
<body style="margin:0;padding:0;background:#f5f6f8;font-family:Arial,Helvetica,sans-serif;color:#111827;">
  <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="background:#f5f6f8;padding:40px 16px;">
    <tr><td align="center">
      <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="max-width:640px;background:#ffffff;border:1px solid #e5e7eb;border-radius:12px;box-shadow:0 4px 16px rgba(17,24,39,.06);">
        <tr><td style="padding:30px 36px 22px;border-bottom:1px solid #e5e7eb;">
          <img src="https://jgltechnologies.com/logo.png" width="64" alt="JGL Technologies" style="display:block;width:64px;height:auto;border:0;">
        </td></tr>
        <tr><td style="padding:34px 36px 40px;">
          <h1 style="margin:0 0 22px;font-size:30px;line-height:1.25;font-weight:700;color:#111827;">{{.Subject}}</h1>
          <div style="font-size:16px;line-height:1.75;color:#4b5563;white-space:pre-wrap;">{{.Body}}</div>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`))

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
	if hasNewline(subject) {
		return errors.New("email subject cannot contain a newline")
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

	message, err := buildAnnouncementEmail(from, subject, body)
	if err != nil {
		return fmt.Errorf("build email: %w", err)
	}

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

func buildAnnouncementEmail(from, subject, body string) ([]byte, error) {
	var htmlBody bytes.Buffer
	if err := announcementEmailTemplate.Execute(&htmlBody, struct {
		Subject string
		Body    string
	}{subject, body}); err != nil {
		return nil, err
	}

	var message bytes.Buffer
	multipartWriter := multipart.NewWriter(&message)
	fmt.Fprintf(&message, "From: JGL Technologies <%s>\r\n", from)
	fmt.Fprint(&message, "To: undisclosed-recipients:;\r\n")
	fmt.Fprintf(&message, "Subject: %s\r\n", mime.QEncoding.Encode("UTF-8", subject))
	fmt.Fprint(&message, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&message, "Content-Type: multipart/alternative; boundary=%q\r\n\r\n", multipartWriter.Boundary())

	plainHeader := make(map[string][]string)
	plainHeader["Content-Type"] = []string{"text/plain; charset=UTF-8"}
	plainHeader["Content-Transfer-Encoding"] = []string{"quoted-printable"}
	plainPart, err := multipartWriter.CreatePart(plainHeader)
	if err != nil {
		return nil, err
	}
	plainEncoder := quotedprintable.NewWriter(plainPart)
	if _, err := plainEncoder.Write([]byte(subject + "\n\n" + body)); err != nil {
		return nil, err
	}
	if err := plainEncoder.Close(); err != nil {
		return nil, err
	}

	htmlHeader := make(map[string][]string)
	htmlHeader["Content-Type"] = []string{"text/html; charset=UTF-8"}
	htmlHeader["Content-Transfer-Encoding"] = []string{"quoted-printable"}
	htmlPart, err := multipartWriter.CreatePart(htmlHeader)
	if err != nil {
		return nil, err
	}
	htmlEncoder := quotedprintable.NewWriter(htmlPart)
	if _, err := htmlEncoder.Write(htmlBody.Bytes()); err != nil {
		return nil, err
	}
	if err := htmlEncoder.Close(); err != nil {
		return nil, err
	}
	if err := multipartWriter.Close(); err != nil {
		return nil, err
	}

	return message.Bytes(), nil
}

func hasNewline(value string) bool {
	return strings.ContainsAny(value, "\r\n")
}

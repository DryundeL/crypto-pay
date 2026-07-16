package email

import (
	"fmt"
	"strconv"

	"github.com/wneessen/go-mail"
)

// Service handles email sending operations
type Service struct {
	host     string
	port     string
	username string
	password string
	from     string
}

// NewService creates a new email service
func NewService(host, port, username, password, from string) *Service {
	return &Service{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

// SendEmail sends an email using SMTP with github.com/wneessen/go-mail
func (s *Service) SendEmail(to, subject, body string) error {
	if s.host == "" || s.port == "" {
		return fmt.Errorf("SMTP configuration is not set")
	}
	// Parse port
	portInt, err := strconv.Atoi(s.port)
	if err != nil {
		return fmt.Errorf("invalid port: %w", err)
	}

	// Create new message
	msg := mail.NewMsg()
	if err := msg.From(s.from); err != nil {
		return fmt.Errorf("failed to set from address: %w", err)
	}
	if err := msg.To(to); err != nil {
		return fmt.Errorf("failed to set to address: %w", err)
	}
	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextPlain, body)

	// Create SMTP client options
	opts := []mail.Option{
		mail.WithPort(portInt),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(s.username),
		mail.WithPassword(s.password),
	}

	// For port 465, use SSL/TLS
	if s.port == "465" {
		opts = append(opts, mail.WithTLSPolicy(mail.TLSMandatory))
	} else {
		// For other ports (like 587), use STARTTLS
		opts = append(opts, mail.WithTLSPolicy(mail.TLSOpportunistic))
	}

	// Create client
	client, err := mail.NewClient(s.host, opts...)
	if err != nil {
		return fmt.Errorf("failed to create mail client: %w", err)
	}
	defer client.Close()

	// Send email
	if err := client.DialAndSend(msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

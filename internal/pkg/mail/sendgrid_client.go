package mail

import (
	"fmt"
	"net/smtp"

	"github.com/adf-code/beta-book-api/config"
	"github.com/adf-code/beta-book-api/internal/entity"
	"github.com/rs/zerolog"
)

type EmailClient interface {
	SendBookCreatedEmail(book entity.Book) error
}

type SendGridClient struct {
	senderName  string
	senderEmail string
	smtpHost    string
	smtpPort    int
	logger      zerolog.Logger
}

func NewSendGridClient(cfg *config.AppConfig, logger zerolog.Logger) *SendGridClient {
	return &SendGridClient{
		senderName:  "Beta Book API",
		senderEmail: cfg.SendGridSenderEmail,
		smtpHost:    "localhost",
		smtpPort:    1025,
		logger:      logger,
	}
}

// Tetap ada, tidak diubah
func (s *SendGridClient) InitSendGrid() *SendGridClient {
	return s
}

func (s *SendGridClient) SendBookCreatedEmail(book entity.Book) error {

	to := "arief.dfaltah@gmail.com" // bisa kamu buat dynamic
	subject := fmt.Sprintf("New Book Created: %s", book.Title)

	htmlBody := fmt.Sprintf(`
		<h1>📚 New Book Created</h1>
		<p>
			<strong>Title:</strong> %s<br>
			<strong>Author:</strong> %s<br>
			<strong>Year:</strong> %d
		</p>
	`, book.Title, book.Author, book.Year)

	message := fmt.Sprintf(
		"From: %s <%s>\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n"+
			"%s",
		s.senderName,
		s.senderEmail,
		to,
		subject,
		htmlBody,
	)

	addr := fmt.Sprintf("%s:%d", s.smtpHost, s.smtpPort)

	err := smtp.SendMail(
		addr,
		nil, // Mailpit tidak perlu auth
		s.senderEmail,
		[]string{to},
		[]byte(message),
	)

	if err != nil {
		s.logger.Error().Err(err).Msg("❌ Failed to send email via Mailpit")
		return err
	}

	s.logger.Info().Msg("✅ Email sent (Mailpit localhost:8025)")
	return nil
}

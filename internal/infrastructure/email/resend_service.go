package email

import (
	"context"
	"fmt"

	"github.com/purushothdl/gochat-backend/internal/config"
	"github.com/resend/resend-go/v2"
)

type ResendService struct {
	client *resend.Client
	config *config.ResendConfig
}

func NewResendService(cfg *config.ResendConfig) *ResendService {
	return &ResendService{
		client: resend.NewClient(cfg.APIKey),
		config: cfg,
	}
}

func (s *ResendService) SendPasswordResetEmail(ctx context.Context, recipientEmail, recipientName, resetLink string) error {
	subject := "Reset Your GoChat Password"
	htmlBody := fmt.Sprintf(`
		<p>Hi %s,</p>
		<p>You requested a password reset. Please click the link below to set a new password. This link is valid for 1 hour.</p>
		<p><a href="%s">Reset Password</a></p>
		<p>If you did not request this, please ignore this email.</p>
		<p>Thanks,<br>The GoChat Team</p>
	`, recipientName, resetLink)

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail),
		To:      []string{recipientEmail},
		Subject: subject,
		Html:    htmlBody,
	}

	_, err := s.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("resend: failed to send password reset email: %w", err)
	}

	return nil
}
package mailer

import (
	"context"
	"fmt"
	"net/smtp"
)

type emailTemplate struct {
	body    string
	subject string
}

var verificationTemplates = map[string]emailTemplate{
	"en": {
		subject: "Verify your BeeTrack account",
		body:    "Hi %s,\n\nPlease verify your email address by clicking the link below:\n\n%s\n\nThis link expires in 24 hours.\n\nIf you did not create a BeeTrack account, you can ignore this email.",
	},
	"pl": {
		subject: "Zweryfikuj swoje konto BeeTrack",
		body:    "Cześć %s,\n\nPotwierdź swój adres email, klikając w poniższy link:\n\n%s\n\nLink wygasa za 24 godziny.\n\nJeśli nie zakładałeś konta w BeeTrack, zignoruj tę wiadomość.",
	},
}

var resetTemplates = map[string]emailTemplate{
	"en": {
		subject: "Reset your BeeTrack password",
		body:    "Hi %s,\n\nClick the link below to reset your password:\n\n%s\n\nThis link expires in 1 hour.\n\nIf you did not request a password reset, you can ignore this email.",
	},
	"pl": {
		subject: "Resetowanie hasła BeeTrack",
		body:    "Cześć %s,\n\nKliknij poniższy link, aby zresetować swoje hasło:\n\n%s\n\nLink wygasa za 1 godzinę.\n\nJeśli nie prosiłeś o reset hasła, zignoruj tę wiadomość.",
	},
}

// Mailer sends transactional emails via SMTP.
type Mailer struct {
	addr string
	auth smtp.Auth
	from string
}

// New creates a Mailer. If user is empty, no SMTP authentication is used (suitable for MailPit).
func New(host, port, user, pass, from string) *Mailer {
	addr := host + ":" + port
	var auth smtp.Auth
	if user != "" {
		auth = smtp.PlainAuth("", user, pass, host)
	}
	return &Mailer{addr: addr, auth: auth, from: from}
}

// SendVerificationEmail sends an account verification email in the requested language,
// falling back to English for unsupported languages.
func (m *Mailer) SendVerificationEmail(_ context.Context, to, name, verificationURL, lang string) error {
	tmpl := resolveTemplate(verificationTemplates, lang)
	return m.send(to, tmpl.subject, fmt.Sprintf(tmpl.body, name, verificationURL))
}

// SendPasswordResetEmail sends a password reset email in the requested language,
// falling back to English for unsupported languages.
func (m *Mailer) SendPasswordResetEmail(_ context.Context, to, name, resetURL, lang string) error {
	tmpl := resolveTemplate(resetTemplates, lang)
	return m.send(to, tmpl.subject, fmt.Sprintf(tmpl.body, name, resetURL))
}

func resolveTemplate(templates map[string]emailTemplate, lang string) emailTemplate {
	if tmpl, ok := templates[lang]; ok {
		return tmpl
	}
	return templates["en"]
}

func (m *Mailer) send(to, subject, body string) error {
	msg := []byte(
		"From: " + m.from + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/plain; charset=UTF-8\r\n" +
			"\r\n" +
			body,
	)
	return smtp.SendMail(m.addr, m.auth, m.from, []string{to}, msg)
}

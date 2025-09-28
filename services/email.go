package services

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"os"

	"github.com/alphabatem/common/context"
	log "github.com/sirupsen/logrus"
)

type EmailService struct {
	context.DefaultService

	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromEmail    string
	fromName     string
	baseURL      string

	templates map[string]*template.Template
}

const EMAIL_SVC = "email_svc"

func (svc EmailService) Id() string {
	return EMAIL_SVC
}

func (svc *EmailService) Configure(ctx *context.Context) error {
	svc.smtpHost = os.Getenv("SMTP_HOST")
	svc.smtpPort = os.Getenv("SMTP_PORT")
	svc.smtpUsername = os.Getenv("SMTP_USERNAME")
	svc.smtpPassword = os.Getenv("SMTP_PASSWORD")
	svc.fromEmail = os.Getenv("FROM_EMAIL")
	svc.fromName = os.Getenv("FROM_NAME")
	svc.baseURL = os.Getenv("BASE_URL")

	// Set defaults if not provided
	if svc.smtpPort == "" {
		svc.smtpPort = "587"
	}
	if svc.fromName == "" {
		svc.fromName = "TechYouth"
	}
	if svc.baseURL == "" {
		svc.baseURL = "http://localhost:8000"
	}

	svc.templates = make(map[string]*template.Template)

	return svc.DefaultService.Configure(ctx)
}

func (svc *EmailService) Start() error {
	// Load email templates
	err := svc.loadTemplates()
	if err != nil {
		log.WithError(err).Error("Failed to load email templates")
		// Don't fail startup, just log the error
	}

	return nil
}

// Email templates
const verificationEmailHTML = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Verify Your Email - {{.AppName}}</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4F46E5; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .button { display: inline-block; padding: 12px 24px; background-color: #4F46E5; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { padding: 20px; text-align: center; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome to {{.AppName}}!</h1>
        </div>
        <div class="content">
            <h2>Hi {{.Username}},</h2>
            <p>Thank you for registering with {{.AppName}}! To complete your registration, please verify your email address by clicking the button below:</p>
            <a href="{{.VerificationURL}}" class="button">Verify Email Address</a>
            <p>If the button doesn't work, you can also copy and paste this link into your browser:</p>
            <p><a href="{{.VerificationURL}}">{{.VerificationURL}}</a></p>
            <p>This verification link will expire in 24 hours.</p>
            <p>If you didn't create an account with {{.AppName}}, you can safely ignore this email.</p>
        </div>
        <div class="footer">
            <p>&copy; 2025 {{.AppName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

const passwordResetEmailHTML = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Reset Your Password - {{.AppName}}</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #DC2626; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .button { display: inline-block; padding: 12px 24px; background-color: #DC2626; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { padding: 20px; text-align: center; color: #666; font-size: 12px; }
        .warning { background-color: #FEF2F2; border-left: 4px solid #DC2626; padding: 10px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Password Reset Request</h1>
        </div>
        <div class="content">
            <h2>Hi {{.Username}},</h2>
            <p>We received a request to reset your password for your {{.AppName}} account.</p>
            <a href="{{.ResetURL}}" class="button">Reset Password</a>
            <p>If the button doesn't work, you can also copy and paste this link into your browser:</p>
            <p><a href="{{.ResetURL}}">{{.ResetURL}}</a></p>
            <div class="warning">
                <strong>Important:</strong> This password reset link will expire in 1 hour for security reasons.
            </div>
            <p>If you didn't request a password reset, you can safely ignore this email. Your password will remain unchanged.</p>
        </div>
        <div class="footer">
            <p>&copy; 2025 {{.AppName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

const loginNotificationEmailHTML = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>New Login Detected - {{.AppName}}</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #059669; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .info-box { background-color: #F0FDF4; border-left: 4px solid #059669; padding: 15px; margin: 20px 0; }
        .footer { padding: 20px; text-align: center; color: #666; font-size: 12px; }
        .details { background-color: white; padding: 15px; border-radius: 5px; margin: 15px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>New Login Detected</h1>
        </div>
        <div class="content">
            <h2>Hi {{.Username}},</h2>
            <p>We detected a new login to your {{.AppName}} account. Here are the details:</p>
            
            <div class="details">
                <strong>Login Details:</strong><br>
                <strong>Time:</strong> {{.LoginTime}}<br>
                <strong>IP Address:</strong> {{.IP}}<br>
                <strong>Device:</strong> {{.Device}}<br>
                <strong>Location:</strong> {{.Location}}
            </div>
            
            <div class="info-box">
                <strong>Was this you?</strong> If you recognize this login, no action is needed.
            </div>
            
            <p>If you don't recognize this login, please:</p>
            <ul>
                <li>Change your password immediately</li>
                <li>Review your account activity</li>
                <li>Contact our support team if you need assistance</li>
            </ul>
        </div>
        <div class="footer">
            <p>&copy; 2025 {{.AppName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

// Template data structures
type VerificationEmailData struct {
	AppName         string
	Username        string
	VerificationURL string
}

type PasswordResetEmailData struct {
	AppName  string
	Username string
	ResetURL string
}

type LoginNotificationEmailData struct {
	AppName   string
	Username  string
	LoginTime string
	IP        string
	Device    string
	Location  string
}

// Load email templates
func (svc *EmailService) loadTemplates() error {
	var err error

	// Load verification email template
	svc.templates["verification"], err = template.New("verification").Parse(verificationEmailHTML)
	if err != nil {
		return fmt.Errorf("failed to parse verification email template: %v", err)
	}

	// Load password reset email template
	svc.templates["password_reset"], err = template.New("password_reset").Parse(passwordResetEmailHTML)
	if err != nil {
		return fmt.Errorf("failed to parse password reset email template: %v", err)
	}

	// Load login notification email template
	svc.templates["login_notification"], err = template.New("login_notification").Parse(loginNotificationEmailHTML)
	if err != nil {
		return fmt.Errorf("failed to parse login notification email template: %v", err)
	}

	return nil
}

// Send verification email
func (svc *EmailService) SendVerificationEmail(email, username, token string) error {
	if svc.smtpHost == "" {
		log.Warn("SMTP not configured, skipping verification email")
		return nil
	}

	verificationURL := fmt.Sprintf("%s/api/v1/verify-email?token=%s", svc.baseURL, token)

	data := VerificationEmailData{
		AppName:         "TechYouth",
		Username:        username,
		VerificationURL: verificationURL,
	}

	subject := "Verify Your Email Address - TechYouth"
	return svc.sendTemplateEmail(email, subject, "verification", data)
}

// Send password reset email
func (svc *EmailService) SendPasswordResetEmail(email, username, token string) error {
	if svc.smtpHost == "" {
		log.Warn("SMTP not configured, skipping password reset email")
		return nil
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", svc.baseURL, token)

	data := PasswordResetEmailData{
		AppName:  "TechYouth",
		Username: username,
		ResetURL: resetURL,
	}

	subject := "Reset Your Password - TechYouth"
	return svc.sendTemplateEmail(email, subject, "password_reset", data)
}

// Send login notification email
func (svc *EmailService) SendLoginNotificationEmail(email, username, loginTime, ip, device, location string) error {
	if svc.smtpHost == "" {
		log.Warn("SMTP not configured, skipping login notification email")
		return nil
	}

	data := LoginNotificationEmailData{
		AppName:   "TechYouth",
		Username:  username,
		LoginTime: loginTime,
		IP:        ip,
		Device:    device,
		Location:  location,
	}

	subject := "New Login Detected - TechYouth"
	return svc.sendTemplateEmail(email, subject, "login_notification", data)
}

// Send template email
func (svc *EmailService) sendTemplateEmail(to, subject, templateName string, data interface{}) error {
	tmpl, exists := svc.templates[templateName]
	if !exists {
		return fmt.Errorf("template %s not found", templateName)
	}

	var body bytes.Buffer
	err := tmpl.Execute(&body, data)
	if err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	return svc.sendEmail(to, subject, body.String())
}

// Send email using SMTP
func (svc *EmailService) sendEmail(to, subject, body string) error {
	if svc.smtpHost == "" {
		return fmt.Errorf("SMTP not configured")
	}

	// Setup authentication
	auth := smtp.PlainAuth("", svc.smtpUsername, svc.smtpPassword, svc.smtpHost)

	// Compose message
	msg := []byte(fmt.Sprintf(
		"From: %s <%s>\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"\r\n"+
			"%s",
		svc.fromName, svc.fromEmail, to, subject, body))

	// Send email
	err := smtp.SendMail(
		svc.smtpHost+":"+svc.smtpPort,
		auth,
		svc.fromEmail,
		[]string{to},
		msg,
	)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{"to": to, "subject": subject}).Error("Failed to send plain email")
		return fmt.Errorf("failed to send email: %v", err)
	}

	log.WithFields(log.Fields{"to": to, "subject": subject}).Info("Plain email sent successfully")
	return nil
}

// Test email configuration
func (svc *EmailService) TestEmailConfig() error {
	if svc.smtpHost == "" {
		return fmt.Errorf("SMTP host not configured")
	}

	testEmail := svc.fromEmail
	if testEmail == "" {
		return fmt.Errorf("from email not configured")
	}

	subject := "Test Email Configuration - TechYouth"
	body := "This is a test email to verify SMTP configuration."

	return svc.SendPlainEmail(testEmail, subject, body)
}

// Send plain text email (for simple notifications)
func (svc *EmailService) SendPlainEmail(to, subject, body string) error {
	if svc.smtpHost == "" {
		log.Warn("SMTP not configured, skipping email")
		return nil
	}

	// Setup authentication
	auth := smtp.PlainAuth("", svc.smtpUsername, svc.smtpPassword, svc.smtpHost)

	// Compose message
	msg := []byte(fmt.Sprintf(
		"From: %s <%s>\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"Content-Type: text/plain; charset=UTF-8\r\n"+
			"\r\n"+
			"%s",
		svc.fromName, svc.fromEmail, to, subject, body))

	// Send email
	err := smtp.SendMail(
		svc.smtpHost+":"+svc.smtpPort,
		auth,
		svc.fromEmail,
		[]string{to},
		msg,
	)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{"to": to, "subject": subject}).Error("Failed to send plain email")
		return fmt.Errorf("failed to send email: %v", err)
	}

	log.WithFields(log.Fields{"to": to, "subject": subject}).Info("Plain email sent successfully")
	return nil
}

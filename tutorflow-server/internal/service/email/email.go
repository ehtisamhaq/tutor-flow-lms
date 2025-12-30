package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"

	"github.com/tutorflow/tutorflow-server/internal/pkg/config"
)

// Service handles email sending
type Service struct {
	cfg       config.EmailConfig
	templates map[string]*template.Template
}

// NewService creates a new email service
func NewService(cfg config.EmailConfig) *Service {
	svc := &Service{
		cfg:       cfg,
		templates: make(map[string]*template.Template),
	}
	svc.loadTemplates()
	return svc
}

// EmailData common email data
type EmailData struct {
	RecipientName string
	SupportEmail  string
	CompanyName   string
	Year          int
	Data          map[string]interface{}
}

// Send sends an email
func (s *Service) Send(to, subject, body string) error {
	if s.cfg.SMTPHost == "" {
		// Skip if email not configured
		return nil
	}

	from := s.cfg.FromEmail
	msg := s.buildMessage(from, to, subject, body)

	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPassword, s.cfg.SMTPHost)
	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)

	return smtp.SendMail(addr, auth, from, []string{to}, msg)
}

// SendHTML sends an HTML email
func (s *Service) SendHTML(to, subject, htmlBody string) error {
	if s.cfg.SMTPHost == "" {
		return nil
	}

	from := s.cfg.FromEmail
	msg := s.buildHTMLMessage(from, to, subject, htmlBody)

	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPassword, s.cfg.SMTPHost)
	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)

	return smtp.SendMail(addr, auth, from, []string{to}, msg)
}

func (s *Service) buildMessage(from, to, subject, body string) []byte {
	msg := fmt.Sprintf("From: %s <%s>\r\n", s.cfg.FromName, from)
	msg += fmt.Sprintf("To: %s\r\n", to)
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += "MIME-Version: 1.0\r\n"
	msg += "Content-Type: text/plain; charset=\"UTF-8\"\r\n"
	msg += "\r\n"
	msg += body

	return []byte(msg)
}

func (s *Service) buildHTMLMessage(from, to, subject, htmlBody string) []byte {
	msg := fmt.Sprintf("From: %s <%s>\r\n", s.cfg.FromName, from)
	msg += fmt.Sprintf("To: %s\r\n", to)
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += "MIME-Version: 1.0\r\n"
	msg += "Content-Type: text/html; charset=\"UTF-8\"\r\n"
	msg += "\r\n"
	msg += htmlBody

	return []byte(msg)
}

// loadTemplates loads email templates
func (s *Service) loadTemplates() {
	// Templates are embedded as strings for simplicity
	s.templates["welcome"] = template.Must(template.New("welcome").Parse(welcomeTemplate))
	s.templates["password_reset"] = template.Must(template.New("password_reset").Parse(passwordResetTemplate))
	s.templates["enrollment"] = template.Must(template.New("enrollment").Parse(enrollmentTemplate))
	s.templates["payment"] = template.Must(template.New("payment").Parse(paymentTemplate))
	s.templates["certificate"] = template.Must(template.New("certificate").Parse(certificateTemplate))
	s.templates["assignment"] = template.Must(template.New("assignment").Parse(assignmentTemplate))
	s.templates["grade"] = template.Must(template.New("grade").Parse(gradeTemplate))
}

// --- Pre-built Email Methods ---

// SendWelcome sends welcome email to new user
func (s *Service) SendWelcome(to, name string) error {
	data := map[string]interface{}{
		"Name":        name,
		"CompanyName": s.cfg.FromName,
	}
	body, err := s.renderTemplate("welcome", data)
	if err != nil {
		return err
	}
	return s.SendHTML(to, "Welcome to TutorFlow!", body)
}

// SendPasswordReset sends password reset email
func (s *Service) SendPasswordReset(to, name, resetURL string) error {
	data := map[string]interface{}{
		"Name":        name,
		"ResetURL":    resetURL,
		"CompanyName": s.cfg.FromName,
	}
	body, err := s.renderTemplate("password_reset", data)
	if err != nil {
		return err
	}
	return s.SendHTML(to, "Reset Your Password", body)
}

// SendEnrollmentConfirmation sends enrollment confirmation
func (s *Service) SendEnrollmentConfirmation(to, name, courseName, courseURL string) error {
	data := map[string]interface{}{
		"Name":        name,
		"CourseName":  courseName,
		"CourseURL":   courseURL,
		"CompanyName": s.cfg.FromName,
	}
	body, err := s.renderTemplate("enrollment", data)
	if err != nil {
		return err
	}
	return s.SendHTML(to, fmt.Sprintf("You're enrolled in %s!", courseName), body)
}

// SendPaymentReceipt sends payment receipt
func (s *Service) SendPaymentReceipt(to, name, orderNumber string, amount float64, items []string) error {
	data := map[string]interface{}{
		"Name":        name,
		"OrderNumber": orderNumber,
		"Amount":      fmt.Sprintf("$%.2f", amount),
		"Items":       items,
		"CompanyName": s.cfg.FromName,
	}
	body, err := s.renderTemplate("payment", data)
	if err != nil {
		return err
	}
	return s.SendHTML(to, fmt.Sprintf("Payment Receipt - Order %s", orderNumber), body)
}

// SendCertificate sends certificate notification
func (s *Service) SendCertificate(to, name, courseName, certNumber, verifyURL string) error {
	data := map[string]interface{}{
		"Name":            name,
		"CourseName":      courseName,
		"CertificateNum":  certNumber,
		"VerificationURL": verifyURL,
		"CompanyName":     s.cfg.FromName,
	}
	body, err := s.renderTemplate("certificate", data)
	if err != nil {
		return err
	}
	return s.SendHTML(to, fmt.Sprintf("Congratulations! Certificate for %s", courseName), body)
}

// SendAssignmentDue sends assignment reminder
func (s *Service) SendAssignmentDue(to, name, assignmentTitle, courseName, dueDate string) error {
	data := map[string]interface{}{
		"Name":            name,
		"AssignmentTitle": assignmentTitle,
		"CourseName":      courseName,
		"DueDate":         dueDate,
		"CompanyName":     s.cfg.FromName,
	}
	body, err := s.renderTemplate("assignment", data)
	if err != nil {
		return err
	}
	return s.SendHTML(to, fmt.Sprintf("Assignment Due: %s", assignmentTitle), body)
}

// SendGradePosted sends grade notification
func (s *Service) SendGradePosted(to, name, itemTitle, courseName string, score float64, maxScore float64) error {
	data := map[string]interface{}{
		"Name":        name,
		"ItemTitle":   itemTitle,
		"CourseName":  courseName,
		"Score":       fmt.Sprintf("%.1f", score),
		"MaxScore":    fmt.Sprintf("%.1f", maxScore),
		"Percentage":  fmt.Sprintf("%.1f%%", (score/maxScore)*100),
		"CompanyName": s.cfg.FromName,
	}
	body, err := s.renderTemplate("grade", data)
	if err != nil {
		return err
	}
	return s.SendHTML(to, fmt.Sprintf("Grade Posted: %s", itemTitle), body)
}

// renderTemplate renders an email template
func (s *Service) renderTemplate(name string, data map[string]interface{}) (string, error) {
	tmpl, ok := s.templates[name]
	if !ok {
		return "", fmt.Errorf("template %s not found", name)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// --- Email Templates ---

const baseTemplate = `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; }
    .container { max-width: 600px; margin: 0 auto; padding: 20px; }
    .header { background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%); color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0; }
    .content { background: #fff; padding: 30px; border: 1px solid #e5e7eb; }
    .button { display: inline-block; background: #6366f1; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin: 20px 0; }
    .footer { background: #f9fafb; padding: 20px; text-align: center; font-size: 12px; color: #6b7280; border-radius: 0 0 8px 8px; }
  </style>
</head>
<body>
  <div class="container">
    {{CONTENT}}
  </div>
</body>
</html>
`

const welcomeTemplate = `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background: #f3f4f6; }
    .container { max-width: 600px; margin: 0 auto; padding: 20px; }
    .header { background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%); color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0; }
    .content { background: #fff; padding: 30px; border: 1px solid #e5e7eb; }
    .button { display: inline-block; background: #6366f1; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin: 20px 0; }
    .footer { background: #f9fafb; padding: 20px; text-align: center; font-size: 12px; color: #6b7280; border-radius: 0 0 8px 8px; border: 1px solid #e5e7eb; border-top: none; }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1>Welcome to TutorFlow!</h1>
    </div>
    <div class="content">
      <h2>Hi {{.Name}},</h2>
      <p>Welcome to TutorFlow! We're excited to have you join our learning community.</p>
      <p>Here's what you can do:</p>
      <ul>
        <li>Browse thousands of courses from expert instructors</li>
        <li>Track your learning progress</li>
        <li>Earn certificates upon completion</li>
        <li>Join discussions and connect with other learners</li>
      </ul>
      <a href="#" class="button">Start Learning</a>
      <p>If you have any questions, feel free to reach out to our support team.</p>
    </div>
    <div class="footer">
      <p>© 2024 {{.CompanyName}}. All rights reserved.</p>
    </div>
  </div>
</body>
</html>
`

const passwordResetTemplate = `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background: #f3f4f6; }
    .container { max-width: 600px; margin: 0 auto; padding: 20px; }
    .header { background: #ef4444; color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0; }
    .content { background: #fff; padding: 30px; border: 1px solid #e5e7eb; }
    .button { display: inline-block; background: #ef4444; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin: 20px 0; }
    .footer { background: #f9fafb; padding: 20px; text-align: center; font-size: 12px; color: #6b7280; border-radius: 0 0 8px 8px; border: 1px solid #e5e7eb; border-top: none; }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1>Password Reset</h1>
    </div>
    <div class="content">
      <h2>Hi {{.Name}},</h2>
      <p>We received a request to reset your password. Click the button below to set a new password:</p>
      <a href="{{.ResetURL}}" class="button">Reset Password</a>
      <p><small>This link will expire in 1 hour. If you didn't request this, you can safely ignore this email.</small></p>
    </div>
    <div class="footer">
      <p>© 2024 {{.CompanyName}}. All rights reserved.</p>
    </div>
  </div>
</body>
</html>
`

const enrollmentTemplate = `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background: #f3f4f6; }
    .container { max-width: 600px; margin: 0 auto; padding: 20px; }
    .header { background: linear-gradient(135deg, #10b981 0%, #059669 100%); color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0; }
    .content { background: #fff; padding: 30px; border: 1px solid #e5e7eb; }
    .button { display: inline-block; background: #10b981; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin: 20px 0; }
    .footer { background: #f9fafb; padding: 20px; text-align: center; font-size: 12px; color: #6b7280; border-radius: 0 0 8px 8px; border: 1px solid #e5e7eb; border-top: none; }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1>You're Enrolled!</h1>
    </div>
    <div class="content">
      <h2>Hi {{.Name}},</h2>
      <p>Great news! You've been successfully enrolled in:</p>
      <h3 style="color: #10b981;">{{.CourseName}}</h3>
      <p>You now have full access to all course materials, quizzes, and discussions.</p>
      <a href="{{.CourseURL}}" class="button">Start Learning</a>
    </div>
    <div class="footer">
      <p>© 2024 {{.CompanyName}}. All rights reserved.</p>
    </div>
  </div>
</body>
</html>
`

const paymentTemplate = `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background: #f3f4f6; }
    .container { max-width: 600px; margin: 0 auto; padding: 20px; }
    .header { background: #1f2937; color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0; }
    .content { background: #fff; padding: 30px; border: 1px solid #e5e7eb; }
    .footer { background: #f9fafb; padding: 20px; text-align: center; font-size: 12px; color: #6b7280; border-radius: 0 0 8px 8px; border: 1px solid #e5e7eb; border-top: none; }
    .receipt { background: #f9fafb; padding: 20px; border-radius: 6px; margin: 20px 0; }
    .total { font-size: 24px; font-weight: bold; color: #10b981; }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1>Payment Receipt</h1>
    </div>
    <div class="content">
      <h2>Hi {{.Name}},</h2>
      <p>Thank you for your purchase! Here's your receipt:</p>
      <div class="receipt">
        <p><strong>Order Number:</strong> {{.OrderNumber}}</p>
        <p><strong>Items:</strong></p>
        <ul>
          {{range .Items}}<li>{{.}}</li>{{end}}
        </ul>
        <p class="total">Total: {{.Amount}}</p>
      </div>
      <p>You can access your courses from your dashboard.</p>
    </div>
    <div class="footer">
      <p>© 2024 {{.CompanyName}}. All rights reserved.</p>
    </div>
  </div>
</body>
</html>
`

const certificateTemplate = `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background: #f3f4f6; }
    .container { max-width: 600px; margin: 0 auto; padding: 20px; }
    .header { background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%); color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0; }
    .content { background: #fff; padding: 30px; border: 1px solid #e5e7eb; }
    .button { display: inline-block; background: #f59e0b; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin: 20px 0; }
    .footer { background: #f9fafb; padding: 20px; text-align: center; font-size: 12px; color: #6b7280; border-radius: 0 0 8px 8px; border: 1px solid #e5e7eb; border-top: none; }
    .cert-box { background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%); padding: 20px; border-radius: 8px; text-align: center; margin: 20px 0; border: 2px solid #f59e0b; }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1>🎉 Congratulations!</h1>
    </div>
    <div class="content">
      <h2>Hi {{.Name}},</h2>
      <p>You've successfully completed:</p>
      <div class="cert-box">
        <h3>{{.CourseName}}</h3>
        <p>Certificate Number: <strong>{{.CertificateNum}}</strong></p>
      </div>
      <p>Your certificate is ready to download and share!</p>
      <a href="{{.VerificationURL}}" class="button">View Certificate</a>
    </div>
    <div class="footer">
      <p>© 2024 {{.CompanyName}}. All rights reserved.</p>
    </div>
  </div>
</body>
</html>
`

const assignmentTemplate = `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background: #f3f4f6; }
    .container { max-width: 600px; margin: 0 auto; padding: 20px; }
    .header { background: #3b82f6; color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0; }
    .content { background: #fff; padding: 30px; border: 1px solid #e5e7eb; }
    .button { display: inline-block; background: #3b82f6; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin: 20px 0; }
    .footer { background: #f9fafb; padding: 20px; text-align: center; font-size: 12px; color: #6b7280; border-radius: 0 0 8px 8px; border: 1px solid #e5e7eb; border-top: none; }
    .due-box { background: #fef2f2; border-left: 4px solid #ef4444; padding: 15px; margin: 20px 0; }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1>Assignment Reminder</h1>
    </div>
    <div class="content">
      <h2>Hi {{.Name}},</h2>
      <p>This is a reminder about an upcoming assignment:</p>
      <div class="due-box">
        <h3>{{.AssignmentTitle}}</h3>
        <p><strong>Course:</strong> {{.CourseName}}</p>
        <p><strong>Due Date:</strong> {{.DueDate}}</p>
      </div>
      <p>Don't forget to submit before the deadline!</p>
      <a href="#" class="button">View Assignment</a>
    </div>
    <div class="footer">
      <p>© 2024 {{.CompanyName}}. All rights reserved.</p>
    </div>
  </div>
</body>
</html>
`

const gradeTemplate = `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background: #f3f4f6; }
    .container { max-width: 600px; margin: 0 auto; padding: 20px; }
    .header { background: #8b5cf6; color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0; }
    .content { background: #fff; padding: 30px; border: 1px solid #e5e7eb; }
    .footer { background: #f9fafb; padding: 20px; text-align: center; font-size: 12px; color: #6b7280; border-radius: 0 0 8px 8px; border: 1px solid #e5e7eb; border-top: none; }
    .grade-box { background: #f5f3ff; padding: 20px; border-radius: 8px; text-align: center; margin: 20px 0; }
    .score { font-size: 36px; font-weight: bold; color: #8b5cf6; }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1>Grade Posted</h1>
    </div>
    <div class="content">
      <h2>Hi {{.Name}},</h2>
      <p>Your grade has been posted for:</p>
      <div class="grade-box">
        <h3>{{.ItemTitle}}</h3>
        <p>{{.CourseName}}</p>
        <p class="score">{{.Score}} / {{.MaxScore}}</p>
        <p>({{.Percentage}})</p>
      </div>
      <p>Keep up the great work!</p>
    </div>
    <div class="footer">
      <p>© 2024 {{.CompanyName}}. All rights reserved.</p>
    </div>
  </div>
</body>
</html>
`

package domain

import (
	"time"

	"github.com/google/uuid"
)

// Certificate represents a completion certificate
type Certificate struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EnrollmentID      uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"enrollment_id"`
	CertificateNumber string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"certificate_number"`
	IssuedAt          time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"issued_at"`
	PDFURL            *string   `gorm:"type:varchar(500)" json:"pdf_url,omitempty"`

	Enrollment *Enrollment `gorm:"foreignKey:EnrollmentID" json:"enrollment,omitempty"`
}

// GenerateCertificateNumber creates a unique certificate number
func GenerateCertificateNumber() string {
	return "TF-" + time.Now().Format("20060102") + "-" + uuid.New().String()[:8]
}

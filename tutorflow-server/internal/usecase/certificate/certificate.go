package certificate

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

// UseCase defines certificate business logic
type UseCase struct {
	certRepo       repository.CertificateRepository
	enrollmentRepo repository.EnrollmentRepository
	courseRepo     repository.CourseRepository
}

// NewUseCase creates a new certificate use case
func NewUseCase(
	certRepo repository.CertificateRepository,
	enrollmentRepo repository.EnrollmentRepository,
	courseRepo repository.CourseRepository,
) *UseCase {
	return &UseCase{
		certRepo:       certRepo,
		enrollmentRepo: enrollmentRepo,
		courseRepo:     courseRepo,
	}
}

// GetCertificate returns certificate by ID
func (uc *UseCase) GetCertificate(ctx context.Context, id uuid.UUID) (*domain.Certificate, error) {
	return uc.certRepo.GetByID(ctx, id)
}

// VerifyCertificate verifies a certificate by its number (public endpoint)
func (uc *UseCase) VerifyCertificate(ctx context.Context, certificateNumber string) (*CertificateVerification, error) {
	cert, err := uc.certRepo.GetByNumber(ctx, certificateNumber)
	if err != nil || cert == nil {
		return nil, fmt.Errorf("certificate not found")
	}

	var userName, courseName string
	if cert.Enrollment != nil {
		if cert.Enrollment.User != nil {
			userName = cert.Enrollment.User.FirstName + " " + cert.Enrollment.User.LastName
		}
		if cert.Enrollment.Course != nil {
			courseName = cert.Enrollment.Course.Title
		}
	}

	return &CertificateVerification{
		Valid:             true,
		CertificateNumber: cert.CertificateNumber,
		HolderName:        userName,
		CourseName:        courseName,
		IssuedAt:          cert.IssuedAt,
	}, nil
}

// CertificateVerification response for public verification
type CertificateVerification struct {
	Valid             bool        `json:"valid"`
	CertificateNumber string      `json:"certificate_number"`
	HolderName        string      `json:"holder_name"`
	CourseName        string      `json:"course_name"`
	IssuedAt          interface{} `json:"issued_at"`
}

// GetMyCertificates returns user's certificates
func (uc *UseCase) GetMyCertificates(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.Certificate, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	return uc.certRepo.GetByUser(ctx, userID, page, limit)
}

// GetCertificateForEnrollment returns certificate for an enrollment
func (uc *UseCase) GetCertificateForEnrollment(ctx context.Context, enrollmentID uuid.UUID) (*domain.Certificate, error) {
	return uc.certRepo.GetByEnrollment(ctx, enrollmentID)
}

// IssueCertificate issues a certificate for a completed enrollment
func (uc *UseCase) IssueCertificate(ctx context.Context, enrollmentID uuid.UUID) (*domain.Certificate, error) {
	// Check if certificate already exists
	existing, _ := uc.certRepo.GetByEnrollment(ctx, enrollmentID)
	if existing != nil {
		return existing, nil
	}

	// Verify enrollment is completed
	enrollment, err := uc.enrollmentRepo.GetByID(ctx, enrollmentID)
	if err != nil || enrollment == nil {
		return nil, fmt.Errorf("enrollment not found")
	}

	if enrollment.Status != domain.EnrollmentStatusCompleted {
		return nil, fmt.Errorf("enrollment is not completed")
	}

	// Create certificate
	cert := &domain.Certificate{
		EnrollmentID:      enrollmentID,
		CertificateNumber: domain.GenerateCertificateNumber(),
	}

	if err := uc.certRepo.Create(ctx, cert); err != nil {
		return nil, err
	}

	// Return with full data
	return uc.certRepo.GetByID(ctx, cert.ID)
}

// RequestCertificate allows user to request certificate for their enrollment
func (uc *UseCase) RequestCertificate(ctx context.Context, userID, courseID uuid.UUID) (*domain.Certificate, error) {
	enrollment, err := uc.enrollmentRepo.GetByUserAndCourse(ctx, userID, courseID)
	if err != nil || enrollment == nil {
		return nil, fmt.Errorf("enrollment not found")
	}

	return uc.IssueCertificate(ctx, enrollment.ID)
}

// UpdatePDFURL updates the PDF URL for a certificate
func (uc *UseCase) UpdatePDFURL(ctx context.Context, certID uuid.UUID, pdfURL string) error {
	cert, err := uc.certRepo.GetByID(ctx, certID)
	if err != nil || cert == nil {
		return fmt.Errorf("certificate not found")
	}

	cert.PDFURL = &pdfURL
	return uc.certRepo.Update(ctx, cert)
}

// CertificateData for PDF generation
type CertificateData struct {
	CertificateNumber string `json:"certificate_number"`
	HolderName        string `json:"holder_name"`
	CourseName        string `json:"course_name"`
	InstructorName    string `json:"instructor_name"`
	IssuedAt          string `json:"issued_at"`
	VerificationURL   string `json:"verification_url"`
}

// GetCertificateData returns data for PDF generation
func (uc *UseCase) GetCertificateData(ctx context.Context, certID uuid.UUID, baseURL string) (*CertificateData, error) {
	cert, err := uc.certRepo.GetByID(ctx, certID)
	if err != nil || cert == nil {
		return nil, fmt.Errorf("certificate not found")
	}

	var holderName, courseName, instructorName string
	if cert.Enrollment != nil {
		if cert.Enrollment.User != nil {
			holderName = cert.Enrollment.User.FirstName + " " + cert.Enrollment.User.LastName
		}
		if cert.Enrollment.Course != nil {
			courseName = cert.Enrollment.Course.Title
			if cert.Enrollment.Course.Instructor != nil {
				instructorName = cert.Enrollment.Course.Instructor.FirstName + " " + cert.Enrollment.Course.Instructor.LastName
			}
		}
	}

	return &CertificateData{
		CertificateNumber: cert.CertificateNumber,
		HolderName:        holderName,
		CourseName:        courseName,
		InstructorName:    instructorName,
		IssuedAt:          cert.IssuedAt.Format("January 2, 2006"),
		VerificationURL:   fmt.Sprintf("%s/verify/%s", baseURL, cert.CertificateNumber),
	}, nil
}

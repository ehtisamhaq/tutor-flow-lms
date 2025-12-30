package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

// CertificateRepository
type certificateRepository struct {
	db *gorm.DB
}

func NewCertificateRepository(db *gorm.DB) repository.CertificateRepository {
	return &certificateRepository{db: db}
}

func (r *certificateRepository) Create(ctx context.Context, cert *domain.Certificate) error {
	return r.db.WithContext(ctx).Create(cert).Error
}

func (r *certificateRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Certificate, error) {
	var cert domain.Certificate
	err := r.db.WithContext(ctx).
		Preload("Enrollment.User").
		Preload("Enrollment.Course").
		Where("id = ?", id).
		First(&cert).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cert, nil
}

func (r *certificateRepository) GetByNumber(ctx context.Context, number string) (*domain.Certificate, error) {
	var cert domain.Certificate
	err := r.db.WithContext(ctx).
		Preload("Enrollment.User").
		Preload("Enrollment.Course").
		Where("certificate_number = ?", number).
		First(&cert).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cert, nil
}

func (r *certificateRepository) GetByEnrollment(ctx context.Context, enrollmentID uuid.UUID) (*domain.Certificate, error) {
	var cert domain.Certificate
	err := r.db.WithContext(ctx).
		Preload("Enrollment.User").
		Preload("Enrollment.Course").
		Where("enrollment_id = ?", enrollmentID).
		First(&cert).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cert, nil
}

func (r *certificateRepository) Update(ctx context.Context, cert *domain.Certificate) error {
	return r.db.WithContext(ctx).Save(cert).Error
}

func (r *certificateRepository) GetByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.Certificate, int64, error) {
	var certs []domain.Certificate
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Certificate{}).
		Joins("JOIN enrollments ON enrollments.id = certificates.enrollment_id").
		Where("enrollments.user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := r.db.WithContext(ctx).
		Preload("Enrollment.Course").
		Preload("Enrollment.User").
		Joins("JOIN enrollments ON enrollments.id = certificates.enrollment_id").
		Where("enrollments.user_id = ?", userID).
		Order("certificates.issued_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&certs).Error

	return certs, total, err
}

func (r *certificateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Certificate{}, "id = ?", id).Error
}

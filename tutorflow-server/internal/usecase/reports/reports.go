package reports

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
	"github.com/tutorflow/tutorflow-server/internal/service/export"
)

// UseCase defines reports and recently viewed business logic
type UseCase struct {
	reportRepo repository.ScheduledReportRepository
	rvRepo     repository.RecentlyViewedRepository
	courseRepo repository.CourseRepository
	exportSvc  *export.Service
}

// NewUseCase creates a new reports use case
func NewUseCase(
	reportRepo repository.ScheduledReportRepository,
	rvRepo repository.RecentlyViewedRepository,
	courseRepo repository.CourseRepository,
	exportSvc *export.Service,
) *UseCase {
	return &UseCase{
		reportRepo: reportRepo,
		rvRepo:     rvRepo,
		courseRepo: courseRepo,
		exportSvc:  exportSvc,
	}
}

// Recently Viewed

func (uc *UseCase) TrackView(ctx context.Context, userID, courseID uuid.UUID) error {
	return uc.rvRepo.Track(ctx, userID, courseID)
}

func (uc *UseCase) GetRecentlyViewed(ctx context.Context, userID uuid.UUID, limit int) ([]domain.RecentlyViewed, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	return uc.rvRepo.GetByUser(ctx, userID, limit)
}

func (uc *UseCase) ClearRecentlyViewed(ctx context.Context, userID uuid.UUID) error {
	return uc.rvRepo.Clear(ctx, userID)
}

// Exports

func (uc *UseCase) ExportData(ctx context.Context, req domain.ExportRequest) (*export.ExportResult, error) {
	return uc.exportSvc.GenerateReport(ctx, req)
}

// Scheduled Reports

func (uc *UseCase) CreateScheduledReport(ctx context.Context, userID uuid.UUID, report *domain.ScheduledReport) (*domain.ScheduledReport, error) {
	report.UserID = userID
	report.NextRunAt = calculateNextRun(report.Schedule, time.Now())

	if err := uc.reportRepo.Create(ctx, report); err != nil {
		return nil, err
	}
	return report, nil
}

func (uc *UseCase) GetMyScheduledReports(ctx context.Context, userID uuid.UUID) ([]domain.ScheduledReport, error) {
	return uc.reportRepo.GetByUser(ctx, userID)
}

func (uc *UseCase) UpdateScheduledReport(ctx context.Context, userID uuid.UUID, id uuid.UUID, update *domain.ScheduledReport) (*domain.ScheduledReport, error) {
	report, err := uc.reportRepo.GetByID(ctx, id)
	if err != nil || report == nil {
		return nil, fmt.Errorf("report not found")
	}
	if report.UserID != userID {
		return nil, fmt.Errorf("forbidden")
	}

	report.Name = update.Name
	report.Schedule = update.Schedule
	report.Format = update.Format
	report.RecipientEmail = update.RecipientEmail
	report.IsActive = update.IsActive
	report.NextRunAt = calculateNextRun(report.Schedule, time.Now())

	if err := uc.reportRepo.Update(ctx, report); err != nil {
		return nil, err
	}
	return report, nil
}

func (uc *UseCase) DeleteScheduledReport(ctx context.Context, userID, id uuid.UUID) error {
	report, err := uc.reportRepo.GetByID(ctx, id)
	if err != nil || report == nil {
		return fmt.Errorf("report not found")
	}
	if report.UserID != userID {
		return fmt.Errorf("forbidden")
	}
	return uc.reportRepo.Delete(ctx, id)
}

func calculateNextRun(schedule string, from time.Time) *time.Time {
	var next time.Time
	switch schedule {
	case domain.ScheduleDaily:
		next = from.AddDate(0, 0, 1)
	case domain.ScheduleWeekly:
		next = from.AddDate(0, 0, 7)
	case domain.ScheduleMonthly:
		next = from.AddDate(0, 1, 0)
	default:
		next = from.AddDate(0, 0, 1)
	}
	return &next
}

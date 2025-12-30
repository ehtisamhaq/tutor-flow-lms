package export

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/tutorflow/tutorflow-server/internal/domain"
)

// Service handles report generation and export
type Service struct {
	db *gorm.DB
}

// NewService creates a new export service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ExportResult contains the generated file
type ExportResult struct {
	Filename    string
	ContentType string
	Data        []byte
}

// GenerateReport generates a report based on type and format
func (s *Service) GenerateReport(ctx context.Context, req domain.ExportRequest) (*ExportResult, error) {
	switch req.ReportType {
	case domain.ReportTypeRevenue:
		return s.generateRevenueReport(ctx, req)
	case domain.ReportTypeEnrollments:
		return s.generateEnrollmentsReport(ctx, req)
	case domain.ReportTypeUsers:
		return s.generateUsersReport(ctx, req)
	case domain.ReportTypeCourses:
		return s.generateCoursesReport(ctx, req)
	case domain.ReportTypeInstructors:
		return s.generateInstructorsReport(ctx, req)
	default:
		return nil, fmt.Errorf("unknown report type: %s", req.ReportType)
	}
}

// Revenue Report
func (s *Service) generateRevenueReport(ctx context.Context, req domain.ExportRequest) (*ExportResult, error) {
	type RevenueRow struct {
		Date        string
		OrderNumber string
		Customer    string
		CourseName  string
		Amount      float64
		Status      string
	}

	query := s.db.WithContext(ctx).
		Table("orders").
		Select(`
			TO_CHAR(orders.created_at, 'YYYY-MM-DD') as date,
			orders.order_number,
			CONCAT(users.first_name, ' ', users.last_name) as customer,
			courses.title as course_name,
			order_items.price as amount,
			orders.status
		`).
		Joins("JOIN users ON users.id = orders.user_id").
		Joins("LEFT JOIN order_items ON order_items.order_id = orders.id").
		Joins("LEFT JOIN courses ON courses.id = order_items.course_id")

	if req.DateFrom != nil {
		query = query.Where("orders.created_at >= ?", req.DateFrom)
	}
	if req.DateTo != nil {
		query = query.Where("orders.created_at <= ?", req.DateTo)
	}

	var rows []RevenueRow
	if err := query.Order("orders.created_at DESC").Scan(&rows).Error; err != nil {
		return nil, err
	}

	// Generate CSV
	headers := []string{"Date", "Order Number", "Customer", "Course", "Amount", "Status"}
	data := make([][]string, len(rows))
	for i, r := range rows {
		data[i] = []string{r.Date, r.OrderNumber, r.Customer, r.CourseName, fmt.Sprintf("%.2f", r.Amount), r.Status}
	}

	return s.generateCSV("revenue_report", headers, data)
}

// Enrollments Report
func (s *Service) generateEnrollmentsReport(ctx context.Context, req domain.ExportRequest) (*ExportResult, error) {
	type EnrollmentRow struct {
		Date       string
		Student    string
		Email      string
		Course     string
		Instructor string
		Status     string
		Progress   float64
	}

	query := s.db.WithContext(ctx).
		Table("enrollments").
		Select(`
			TO_CHAR(enrollments.enrolled_at, 'YYYY-MM-DD') as date,
			CONCAT(u.first_name, ' ', u.last_name) as student,
			u.email,
			courses.title as course,
			CONCAT(i.first_name, ' ', i.last_name) as instructor,
			enrollments.status,
			enrollments.progress
		`).
		Joins("JOIN users u ON u.id = enrollments.user_id").
		Joins("JOIN courses ON courses.id = enrollments.course_id").
		Joins("LEFT JOIN users i ON i.id = courses.instructor_id")

	if req.DateFrom != nil {
		query = query.Where("enrollments.enrolled_at >= ?", req.DateFrom)
	}
	if req.DateTo != nil {
		query = query.Where("enrollments.enrolled_at <= ?", req.DateTo)
	}

	var rows []EnrollmentRow
	if err := query.Order("enrollments.enrolled_at DESC").Scan(&rows).Error; err != nil {
		return nil, err
	}

	headers := []string{"Date", "Student", "Email", "Course", "Instructor", "Status", "Progress %"}
	data := make([][]string, len(rows))
	for i, r := range rows {
		data[i] = []string{r.Date, r.Student, r.Email, r.Course, r.Instructor, r.Status, fmt.Sprintf("%.1f", r.Progress)}
	}

	return s.generateCSV("enrollments_report", headers, data)
}

// Users Report
func (s *Service) generateUsersReport(ctx context.Context, req domain.ExportRequest) (*ExportResult, error) {
	type UserRow struct {
		Date     string
		Name     string
		Email    string
		Role     string
		IsActive string
	}

	query := s.db.WithContext(ctx).
		Table("users").
		Select(`
			TO_CHAR(created_at, 'YYYY-MM-DD') as date,
			CONCAT(first_name, ' ', last_name) as name,
			email,
			role,
			CASE WHEN is_active THEN 'Yes' ELSE 'No' END as is_active
		`)

	if req.DateFrom != nil {
		query = query.Where("created_at >= ?", req.DateFrom)
	}
	if req.DateTo != nil {
		query = query.Where("created_at <= ?", req.DateTo)
	}

	var rows []UserRow
	if err := query.Order("created_at DESC").Scan(&rows).Error; err != nil {
		return nil, err
	}

	headers := []string{"Registered Date", "Name", "Email", "Role", "Active"}
	data := make([][]string, len(rows))
	for i, r := range rows {
		data[i] = []string{r.Date, r.Name, r.Email, r.Role, r.IsActive}
	}

	return s.generateCSV("users_report", headers, data)
}

// Courses Report
func (s *Service) generateCoursesReport(ctx context.Context, req domain.ExportRequest) (*ExportResult, error) {
	type CourseRow struct {
		Title      string
		Instructor string
		Category   string
		Level      string
		Price      float64
		Students   int
		Rating     float64
		Status     string
		Created    string
	}

	query := s.db.WithContext(ctx).
		Table("courses").
		Select(`
			courses.title,
			CONCAT(users.first_name, ' ', users.last_name) as instructor,
			COALESCE(categories.name, 'Uncategorized') as category,
			courses.level,
			courses.price,
			courses.total_students as students,
			courses.rating,
			courses.status,
			TO_CHAR(courses.created_at, 'YYYY-MM-DD') as created
		`).
		Joins("LEFT JOIN users ON users.id = courses.instructor_id").
		Joins("LEFT JOIN categories ON categories.id = courses.category_id")

	var rows []CourseRow
	if err := query.Order("courses.created_at DESC").Scan(&rows).Error; err != nil {
		return nil, err
	}

	headers := []string{"Title", "Instructor", "Category", "Level", "Price", "Students", "Rating", "Status", "Created"}
	data := make([][]string, len(rows))
	for i, r := range rows {
		data[i] = []string{r.Title, r.Instructor, r.Category, r.Level, fmt.Sprintf("%.2f", r.Price),
			fmt.Sprintf("%d", r.Students), fmt.Sprintf("%.1f", r.Rating), r.Status, r.Created}
	}

	return s.generateCSV("courses_report", headers, data)
}

// Instructors Report
func (s *Service) generateInstructorsReport(ctx context.Context, req domain.ExportRequest) (*ExportResult, error) {
	type InstructorRow struct {
		Name          string
		Email         string
		TotalCourses  int
		TotalStudents int64
		AvgRating     float64
	}

	query := s.db.WithContext(ctx).
		Table("users").
		Select(`
			CONCAT(users.first_name, ' ', users.last_name) as name,
			users.email,
			COUNT(DISTINCT courses.id) as total_courses,
			COALESCE(SUM(courses.total_students), 0) as total_students,
			COALESCE(AVG(courses.rating), 0) as avg_rating
		`).
		Joins("LEFT JOIN courses ON courses.instructor_id = users.id AND courses.status = ?", domain.CourseStatusPublished).
		Where("users.role = ?", domain.RoleTutor).
		Group("users.id")

	var rows []InstructorRow
	if err := query.Order("total_students DESC").Scan(&rows).Error; err != nil {
		return nil, err
	}

	headers := []string{"Name", "Email", "Total Courses", "Total Students", "Avg Rating"}
	data := make([][]string, len(rows))
	for i, r := range rows {
		data[i] = []string{r.Name, r.Email, fmt.Sprintf("%d", r.TotalCourses),
			fmt.Sprintf("%d", r.TotalStudents), fmt.Sprintf("%.2f", r.AvgRating)}
	}

	return s.generateCSV("instructors_report", headers, data)
}

// generateCSV creates a CSV file from headers and data
func (s *Service) generateCSV(name string, headers []string, data [][]string) (*ExportResult, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write headers
	if err := writer.Write(headers); err != nil {
		return nil, err
	}

	// Write data rows
	for _, row := range data {
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	timestamp := time.Now().Format("20060102_150405")
	return &ExportResult{
		Filename:    fmt.Sprintf("%s_%s.csv", name, timestamp),
		ContentType: "text/csv",
		Data:        buf.Bytes(),
	}, nil
}

// GetScheduledReportNextRun calculates the next run time
func GetScheduledReportNextRun(schedule string, from time.Time) time.Time {
	switch schedule {
	case domain.ScheduleDaily:
		return from.AddDate(0, 0, 1).Truncate(24 * time.Hour).Add(6 * time.Hour) // 6 AM next day
	case domain.ScheduleWeekly:
		return from.AddDate(0, 0, 7).Truncate(24 * time.Hour).Add(6 * time.Hour) // 6 AM in 7 days
	case domain.ScheduleMonthly:
		return from.AddDate(0, 1, 0).Truncate(24 * time.Hour).Add(6 * time.Hour) // 6 AM next month
	default:
		return from.AddDate(0, 0, 1)
	}
}

// ProcessScheduledReports processes due scheduled reports
func (s *Service) ProcessScheduledReports(ctx context.Context, reports []domain.ScheduledReport, emailFn func(email string, data []byte, filename string) error) error {
	for _, report := range reports {
		req := domain.ExportRequest{
			ReportType: report.ReportType,
			Format:     report.Format,
		}

		result, err := s.GenerateReport(ctx, req)
		if err != nil {
			continue
		}

		// Send email if recipient specified
		if report.RecipientEmail != nil && *report.RecipientEmail != "" && emailFn != nil {
			_ = emailFn(*report.RecipientEmail, result.Data, result.Filename)
		}
	}
	return nil
}

package admin

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tutorflow/tutorflow-server/internal/domain"
)

// UseCase defines admin dashboard business logic
type UseCase struct {
	db *gorm.DB
}

// NewUseCase creates a new admin use case
func NewUseCase(db *gorm.DB) *UseCase {
	return &UseCase{db: db}
}

// DashboardStats contains main dashboard metrics
type DashboardStats struct {
	Users       UserStats       `json:"users"`
	Courses     CourseStats     `json:"courses"`
	Revenue     RevenueStats    `json:"revenue"`
	Enrollments EnrollmentStats `json:"enrollments"`
}

type UserStats struct {
	TotalUsers    int64 `json:"total_users"`
	NewUsersToday int64 `json:"new_users_today"`
	NewUsersWeek  int64 `json:"new_users_week"`
	NewUsersMonth int64 `json:"new_users_month"`
	ActiveUsers   int64 `json:"active_users"`
	TotalStudents int64 `json:"total_students"`
	TotalTutors   int64 `json:"total_tutors"`
}

type CourseStats struct {
	TotalCourses     int64   `json:"total_courses"`
	PublishedCourses int64   `json:"published_courses"`
	DraftCourses     int64   `json:"draft_courses"`
	TotalLessons     int64   `json:"total_lessons"`
	AvgRating        float64 `json:"avg_rating"`
}

type RevenueStats struct {
	TotalRevenue   float64 `json:"total_revenue"`
	RevenueToday   float64 `json:"revenue_today"`
	RevenueWeek    float64 `json:"revenue_week"`
	RevenueMonth   float64 `json:"revenue_month"`
	TotalOrders    int64   `json:"total_orders"`
	AvgOrderValue  float64 `json:"avg_order_value"`
	PendingPayouts float64 `json:"pending_payouts"`
}

type EnrollmentStats struct {
	TotalEnrollments  int64   `json:"total_enrollments"`
	ActiveEnrollments int64   `json:"active_enrollments"`
	CompletedCourses  int64   `json:"completed_courses"`
	CompletionRate    float64 `json:"completion_rate"`
	EnrollmentsToday  int64   `json:"enrollments_today"`
	EnrollmentsWeek   int64   `json:"enrollments_week"`
}

// GetDashboardStats returns main dashboard statistics
func (uc *UseCase) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	stats := &DashboardStats{}

	// User stats
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekAgo := today.AddDate(0, 0, -7)
	monthAgo := today.AddDate(0, -1, 0)

	uc.db.WithContext(ctx).Model(&domain.User{}).Count(&stats.Users.TotalUsers)
	uc.db.WithContext(ctx).Model(&domain.User{}).Where("created_at >= ?", today).Count(&stats.Users.NewUsersToday)
	uc.db.WithContext(ctx).Model(&domain.User{}).Where("created_at >= ?", weekAgo).Count(&stats.Users.NewUsersWeek)
	uc.db.WithContext(ctx).Model(&domain.User{}).Where("created_at >= ?", monthAgo).Count(&stats.Users.NewUsersMonth)
	uc.db.WithContext(ctx).Model(&domain.User{}).Where("role = ?", domain.RoleStudent).Count(&stats.Users.TotalStudents)
	uc.db.WithContext(ctx).Model(&domain.User{}).Where("role = ?", domain.RoleTutor).Count(&stats.Users.TotalTutors)

	// Course stats
	uc.db.WithContext(ctx).Model(&domain.Course{}).Count(&stats.Courses.TotalCourses)
	uc.db.WithContext(ctx).Model(&domain.Course{}).Where("status = ?", domain.CourseStatusPublished).Count(&stats.Courses.PublishedCourses)
	uc.db.WithContext(ctx).Model(&domain.Course{}).Where("status = ?", domain.CourseStatusDraft).Count(&stats.Courses.DraftCourses)
	uc.db.WithContext(ctx).Model(&domain.Lesson{}).Count(&stats.Courses.TotalLessons)

	var avgRating struct{ Avg float64 }
	uc.db.WithContext(ctx).Model(&domain.Course{}).Select("AVG(rating) as avg").Where("status = ?", domain.CourseStatusPublished).Scan(&avgRating)
	stats.Courses.AvgRating = avgRating.Avg

	// Revenue stats
	var totalRevenue struct{ Sum float64 }
	uc.db.WithContext(ctx).Model(&domain.Order{}).Select("COALESCE(SUM(total_amount), 0) as sum").Where("status = ?", domain.OrderStatusCompleted).Scan(&totalRevenue)
	stats.Revenue.TotalRevenue = totalRevenue.Sum

	var revenueToday struct{ Sum float64 }
	uc.db.WithContext(ctx).Model(&domain.Order{}).Select("COALESCE(SUM(total_amount), 0) as sum").Where("status = ? AND created_at >= ?", domain.OrderStatusCompleted, today).Scan(&revenueToday)
	stats.Revenue.RevenueToday = revenueToday.Sum

	var revenueWeek struct{ Sum float64 }
	uc.db.WithContext(ctx).Model(&domain.Order{}).Select("COALESCE(SUM(total_amount), 0) as sum").Where("status = ? AND created_at >= ?", domain.OrderStatusCompleted, weekAgo).Scan(&revenueWeek)
	stats.Revenue.RevenueWeek = revenueWeek.Sum

	var revenueMonth struct{ Sum float64 }
	uc.db.WithContext(ctx).Model(&domain.Order{}).Select("COALESCE(SUM(total_amount), 0) as sum").Where("status = ? AND created_at >= ?", domain.OrderStatusCompleted, monthAgo).Scan(&revenueMonth)
	stats.Revenue.RevenueMonth = revenueMonth.Sum

	uc.db.WithContext(ctx).Model(&domain.Order{}).Where("status = ?", domain.OrderStatusCompleted).Count(&stats.Revenue.TotalOrders)
	if stats.Revenue.TotalOrders > 0 {
		stats.Revenue.AvgOrderValue = stats.Revenue.TotalRevenue / float64(stats.Revenue.TotalOrders)
	}

	// Enrollment stats
	uc.db.WithContext(ctx).Model(&domain.Enrollment{}).Count(&stats.Enrollments.TotalEnrollments)
	uc.db.WithContext(ctx).Model(&domain.Enrollment{}).Where("status = ?", domain.EnrollmentStatusActive).Count(&stats.Enrollments.ActiveEnrollments)
	uc.db.WithContext(ctx).Model(&domain.Enrollment{}).Where("status = ?", domain.EnrollmentStatusCompleted).Count(&stats.Enrollments.CompletedCourses)
	uc.db.WithContext(ctx).Model(&domain.Enrollment{}).Where("enrolled_at >= ?", today).Count(&stats.Enrollments.EnrollmentsToday)
	uc.db.WithContext(ctx).Model(&domain.Enrollment{}).Where("enrolled_at >= ?", weekAgo).Count(&stats.Enrollments.EnrollmentsWeek)

	if stats.Enrollments.TotalEnrollments > 0 {
		stats.Enrollments.CompletionRate = float64(stats.Enrollments.CompletedCourses) / float64(stats.Enrollments.TotalEnrollments) * 100
	}

	return stats, nil
}

// RevenueChartData for time series charts
type RevenueChartData struct {
	Labels []string  `json:"labels"`
	Values []float64 `json:"values"`
}

// GetRevenueChart returns revenue over time
func (uc *UseCase) GetRevenueChart(ctx context.Context, period string) (*RevenueChartData, error) {
	var days int
	var format string
	switch period {
	case "week":
		days = 7
		format = "Mon"
	case "month":
		days = 30
		format = "Jan 02"
	case "year":
		days = 365
		format = "Jan"
	default:
		days = 30
		format = "Jan 02"
	}

	data := &RevenueChartData{
		Labels: make([]string, days),
		Values: make([]float64, days),
	}

	today := time.Now()
	for i := days - 1; i >= 0; i-- {
		date := today.AddDate(0, 0, -i)
		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDay := startOfDay.AddDate(0, 0, 1)

		var revenue struct{ Sum float64 }
		uc.db.WithContext(ctx).Model(&domain.Order{}).
			Select("COALESCE(SUM(total_amount), 0) as sum").
			Where("status = ? AND created_at >= ? AND created_at < ?", domain.OrderStatusCompleted, startOfDay, endOfDay).
			Scan(&revenue)

		idx := days - 1 - i
		data.Labels[idx] = date.Format(format)
		data.Values[idx] = revenue.Sum
	}

	return data, nil
}

// TopCourse represents a top performing course
type TopCourse struct {
	ID            uuid.UUID `json:"id"`
	Title         string    `json:"title"`
	Instructor    string    `json:"instructor"`
	TotalStudents int64     `json:"total_students"`
	Revenue       float64   `json:"revenue"`
	Rating        float64   `json:"rating"`
}

// GetTopCourses returns top performing courses
func (uc *UseCase) GetTopCourses(ctx context.Context, limit int, sortBy string) ([]TopCourse, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}

	orderClause := "total_students DESC"
	switch sortBy {
	case "revenue":
		orderClause = "revenue DESC"
	case "rating":
		orderClause = "rating DESC"
	case "enrollments":
		orderClause = "total_students DESC"
	}

	var courses []TopCourse
	err := uc.db.WithContext(ctx).
		Table("courses").
		Select(`
			courses.id,
			courses.title,
			CONCAT(users.first_name, ' ', users.last_name) as instructor,
			courses.total_students,
			courses.rating,
			COALESCE(SUM(order_items.price), 0) as revenue
		`).
		Joins("LEFT JOIN users ON users.id = courses.instructor_id").
		Joins("LEFT JOIN order_items ON order_items.course_id = courses.id").
		Joins("LEFT JOIN orders ON orders.id = order_items.order_id AND orders.status = ?", domain.OrderStatusCompleted).
		Where("courses.status = ?", domain.CourseStatusPublished).
		Group("courses.id, users.first_name, users.last_name").
		Order(orderClause).
		Limit(limit).
		Scan(&courses).Error

	return courses, err
}

// TopInstructor represents a top performing instructor
type TopInstructor struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	TotalCourses  int64     `json:"total_courses"`
	TotalStudents int64     `json:"total_students"`
	TotalRevenue  float64   `json:"total_revenue"`
	AvgRating     float64   `json:"avg_rating"`
}

// GetTopInstructors returns top performing instructors
func (uc *UseCase) GetTopInstructors(ctx context.Context, limit int) ([]TopInstructor, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}

	var instructors []TopInstructor
	err := uc.db.WithContext(ctx).
		Table("users").
		Select(`
			users.id,
			CONCAT(users.first_name, ' ', users.last_name) as name,
			COUNT(DISTINCT courses.id) as total_courses,
			COALESCE(SUM(courses.total_students), 0) as total_students,
			COALESCE(AVG(courses.rating), 0) as avg_rating
		`).
		Joins("JOIN courses ON courses.instructor_id = users.id AND courses.status = ?", domain.CourseStatusPublished).
		Where("users.role = ?", domain.RoleTutor).
		Group("users.id").
		Order("total_students DESC").
		Limit(limit).
		Scan(&instructors).Error

	return instructors, err
}

// RecentOrder for recent orders list
type RecentOrder struct {
	ID          uuid.UUID `json:"id"`
	OrderNumber string    `json:"order_number"`
	UserName    string    `json:"user_name"`
	UserEmail   string    `json:"user_email"`
	TotalAmount float64   `json:"total_amount"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// GetRecentOrders returns recent orders
func (uc *UseCase) GetRecentOrders(ctx context.Context, limit int) ([]RecentOrder, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}

	var orders []RecentOrder
	err := uc.db.WithContext(ctx).
		Table("orders").
		Select(`
			orders.id,
			orders.order_number,
			CONCAT(users.first_name, ' ', users.last_name) as user_name,
			users.email as user_email,
			orders.total_amount,
			orders.status,
			orders.created_at
		`).
		Joins("JOIN users ON users.id = orders.user_id").
		Order("orders.created_at DESC").
		Limit(limit).
		Scan(&orders).Error

	return orders, err
}

// RecentUser for recent registrations
type RecentUser struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// GetRecentUsers returns recently registered users
func (uc *UseCase) GetRecentUsers(ctx context.Context, limit int) ([]RecentUser, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}

	var users []RecentUser
	err := uc.db.WithContext(ctx).
		Table("users").
		Select(`
			id,
			CONCAT(first_name, ' ', last_name) as name,
			email,
			role,
			created_at
		`).
		Order("created_at DESC").
		Limit(limit).
		Scan(&users).Error

	return users, err
}

// SystemHealth for system monitoring
type SystemHealth struct {
	DatabaseConnection bool   `json:"database_connection"`
	ActiveConnections  int    `json:"active_connections"`
	ServerUptime       string `json:"server_uptime"`
}

// GetSystemHealth returns system health info
func (uc *UseCase) GetSystemHealth(ctx context.Context) (*SystemHealth, error) {
	health := &SystemHealth{
		DatabaseConnection: true,
	}

	// Check DB connection
	sqlDB, err := uc.db.DB()
	if err != nil {
		health.DatabaseConnection = false
		return health, nil
	}

	stats := sqlDB.Stats()
	health.ActiveConnections = stats.OpenConnections

	return health, nil
}

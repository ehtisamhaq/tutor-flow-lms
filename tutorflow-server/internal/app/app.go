package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"

	_ "github.com/tutorflow/tutorflow-server/docs"
	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/handler"
	appMiddleware "github.com/tutorflow/tutorflow-server/internal/middleware"
	"github.com/tutorflow/tutorflow-server/internal/pkg/config"
	"github.com/tutorflow/tutorflow-server/internal/pkg/database"
	"github.com/tutorflow/tutorflow-server/internal/pkg/jwt"
	"github.com/tutorflow/tutorflow-server/internal/repository/postgres"
	"github.com/tutorflow/tutorflow-server/internal/service/email"
	"github.com/tutorflow/tutorflow-server/internal/service/export"
	"github.com/tutorflow/tutorflow-server/internal/service/payment"
	"github.com/tutorflow/tutorflow-server/internal/service/push"
	"github.com/tutorflow/tutorflow-server/internal/service/storage"
	"github.com/tutorflow/tutorflow-server/internal/usecase/admin"
	"github.com/tutorflow/tutorflow-server/internal/usecase/announcement"
	"github.com/tutorflow/tutorflow-server/internal/usecase/auth"
	"github.com/tutorflow/tutorflow-server/internal/usecase/cart"
	"github.com/tutorflow/tutorflow-server/internal/usecase/certificate"
	"github.com/tutorflow/tutorflow-server/internal/usecase/course"
	"github.com/tutorflow/tutorflow-server/internal/usecase/discussion"
	"github.com/tutorflow/tutorflow-server/internal/usecase/enrollment"
	"github.com/tutorflow/tutorflow-server/internal/usecase/learningpath"
	"github.com/tutorflow/tutorflow-server/internal/usecase/message"
	"github.com/tutorflow/tutorflow-server/internal/usecase/notification"
	"github.com/tutorflow/tutorflow-server/internal/usecase/order"
	"github.com/tutorflow/tutorflow-server/internal/usecase/quiz"
	"github.com/tutorflow/tutorflow-server/internal/usecase/reports"
	"github.com/tutorflow/tutorflow-server/internal/usecase/review"
	"github.com/tutorflow/tutorflow-server/internal/usecase/search"
	"github.com/tutorflow/tutorflow-server/internal/usecase/user"
)

// App is the main application struct
type App struct {
	cfg    *config.Config
	logger *zap.SugaredLogger
	echo   *echo.Echo
}

// New creates a new application instance
func New(cfg *config.Config, logger *zap.SugaredLogger) *App {
	return &App{
		cfg:    cfg,
		logger: logger,
		echo:   echo.New(),
	}
}

// Run starts the application
func (a *App) Run() error {
	// Connect to database
	db, err := database.Connect(a.cfg.Database)
	if err != nil {
		return err
	}
	a.logger.Info("Connected to database")

	// Run migrations
	if err := database.AutoMigrate(db); err != nil {
		return err
	}
	a.logger.Info("Database migrations completed")

	// Middleware
	a.echo.HideBanner = true
	a.echo.Use(middleware.RequestID())
	a.echo.Use(middleware.Recover())
	a.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: a.cfg.Server.AllowedOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "X-Session-ID"},
	}))
	a.echo.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"time":"${time_rfc3339}","id":"${id}","method":"${method}","uri":"${uri}","status":${status},"latency":"${latency_human}"}` + "\n",
	}))

	// Custom Error Handler
	a.echo.HTTPErrorHandler = appMiddleware.ErrorHandler(a.logger)

	// Serve static uploads
	a.echo.Static("/uploads", a.cfg.Storage.LocalPath)

	// Health check
	a.echo.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "healthy",
			"version": a.cfg.Server.Version,
		})
	})

	// Initialize JWT Manager
	jwtManager := jwt.NewManager(a.cfg.JWT)

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	refreshTokenRepo := postgres.NewRefreshTokenRepository(db)
	tutorProfileRepo := postgres.NewTutorProfileRepository(db)
	courseRepo := postgres.NewCourseRepository(db)
	categoryRepo := postgres.NewCategoryRepository(db)
	moduleRepo := postgres.NewModuleRepository(db)
	lessonRepo := postgres.NewLessonRepository(db)
	enrollmentRepo := postgres.NewEnrollmentRepository(db)
	progressRepo := postgres.NewLessonProgressRepository(db)
	notificationRepo := postgres.NewNotificationRepository(db)
	cartRepo := postgres.NewCartRepository(db)
	wishlistRepo := postgres.NewWishlistRepository(db)
	couponRepo := postgres.NewCouponRepository(db)
	orderRepo := postgres.NewOrderRepository(db)
	quizRepo := postgres.NewQuizRepository(db)
	attemptRepo := postgres.NewQuizAttemptRepository(db)
	assignmentRepo := postgres.NewAssignmentRepository(db)
	submissionRepo := postgres.NewSubmissionRepository(db)
	reviewRepo := postgres.NewReviewRepository(db)
	discussionRepo := postgres.NewDiscussionRepository(db)
	certRepo := postgres.NewCertificateRepository(db)
	searchRepo := postgres.NewSearchRepository(db)
	announcementRepo := postgres.NewAnnouncementRepository(db)
	pushRepo := postgres.NewPushSubscriptionRepository(db)
	rvRepo := postgres.NewRecentlyViewedRepository(db)
	scheduledReportRepo := postgres.NewScheduledReportRepository(db)
	messageRepo := postgres.NewMessageRepository(db)
	learningPathRepo := postgres.NewLearningPathRepository(db)

	// Initialize services
	storageSvc := storage.NewService(a.cfg.Storage)
	paymentSvc := payment.NewService(a.cfg.Stripe)
	emailSvc := email.NewService(a.cfg.Email)
	_ = emailSvc
	pushSvc := push.NewService(push.Config{
		VAPIDPublicKey:  a.cfg.Push.VAPIDPublicKey,
		VAPIDPrivateKey: a.cfg.Push.VAPIDPrivateKey,
		VAPIDSubject:    a.cfg.Push.VAPIDSubject,
	}, pushRepo)
	exportSvc := export.NewService(db)

	// Initialize use cases
	authUC := auth.NewUseCase(userRepo, refreshTokenRepo, jwtManager)
	userUC := user.NewUseCase(userRepo, tutorProfileRepo)
	courseUC := course.NewUseCase(courseRepo, categoryRepo, moduleRepo, lessonRepo, enrollmentRepo)
	enrollmentUC := enrollment.NewUseCase(enrollmentRepo, progressRepo, courseRepo, notificationRepo)
	cartUC := cart.NewUseCase(cartRepo, wishlistRepo, courseRepo, enrollmentRepo)
	orderUC := order.NewUseCase(orderRepo, cartRepo, couponRepo, enrollmentRepo, courseRepo, paymentSvc)
	quizUC := quiz.NewUseCase(quizRepo, attemptRepo, assignmentRepo, submissionRepo, enrollmentRepo, progressRepo)
	reviewUC := review.NewUseCase(reviewRepo, enrollmentRepo, courseRepo, notificationRepo)
	notificationUC := notification.NewUseCase(notificationRepo, enrollmentRepo)
	discussionUC := discussion.NewUseCase(discussionRepo, enrollmentRepo, courseRepo)
	certificateUC := certificate.NewUseCase(certRepo, enrollmentRepo, courseRepo)
	reportUC := reports.NewUseCase(scheduledReportRepo, rvRepo, courseRepo, exportSvc)
	searchUC := search.NewUseCase(searchRepo, courseRepo, categoryRepo)
	adminUC := admin.NewUseCase(db)
	announcementUC := announcement.NewUseCase(announcementRepo, courseRepo, enrollmentRepo, notificationRepo)
	messageUC := message.NewUseCase(messageRepo, userRepo)
	learningPathUC := learningpath.NewUseCase(learningPathRepo, enrollmentRepo, certRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authUC)
	userHandler := handler.NewUserHandler(userUC)
	courseHandler := handler.NewCourseHandler(courseUC)
	enrollmentHandler := handler.NewEnrollmentHandler(enrollmentUC)
	uploadHandler := handler.NewUploadHandler(storageSvc)
	cartHandler := handler.NewCartHandler(cartUC)
	orderHandler := handler.NewOrderHandler(orderUC, paymentSvc)
	quizHandler := handler.NewQuizHandler(quizUC)
	reviewHandler := handler.NewReviewHandler(reviewUC)
	notificationHandler := handler.NewNotificationHandler(notificationUC)
	discussionHandler := handler.NewDiscussionHandler(discussionUC)
	certificateHandler := handler.NewCertificateHandler(certificateUC)
	searchHandler := handler.NewSearchHandler(searchUC)
	adminHandler := handler.NewAdminHandler(adminUC)
	announcementHandler := handler.NewAnnouncementHandler(announcementUC)
	messageHandler := handler.NewMessageHandler(messageUC)
	pushHandler := handler.NewPushHandler(pushSvc)
	learningPathHandler := handler.NewLearningPathHandler(learningPathUC)
	reportHandler := handler.NewReportHandler(reportUC)

	// Middleware functions
	authMW := appMiddleware.AuthMiddleware(jwtManager)
	optionalAuthMW := appMiddleware.OptionalAuthMiddleware(jwtManager)
	adminMW := appMiddleware.RequireAdmin()
	managerMW := appMiddleware.RequireAdminOrManager()
	tutorMW := appMiddleware.RequireRole(domain.RoleAdmin, domain.RoleManager, domain.RoleTutor)

	// API v1 routes
	api := a.echo.Group("/api/v1")

	// Register routes
	authHandler.RegisterRoutes(api.Group("/auth"), authMW)
	userHandler.RegisterRoutes(api.Group("/users"), authMW, adminMW, managerMW)
	courseHandler.RegisterRoutes(api.Group("/courses"), authMW, optionalAuthMW, tutorMW, adminMW)
	enrollmentHandler.RegisterRoutes(api.Group("/enrollments"), authMW, managerMW)
	uploadHandler.RegisterRoutes(api.Group("/uploads"), authMW)
	cartHandler.RegisterRoutes(api, authMW, optionalAuthMW)
	orderHandler.RegisterRoutes(api.Group("/orders"), authMW, adminMW)
	quizHandler.RegisterRoutes(api, authMW, tutorMW)
	reviewHandler.RegisterRoutes(api, authMW, tutorMW)
	notificationHandler.RegisterRoutes(api, authMW)
	discussionHandler.RegisterRoutes(api, authMW)
	certificateHandler.RegisterRoutes(api, authMW)
	searchHandler.RegisterRoutes(api, optionalAuthMW)
	adminHandler.RegisterRoutes(api, authMW, adminMW)
	announcementHandler.RegisterRoutes(api, authMW, tutorMW)
	messageHandler.RegisterRoutes(api, authMW)
	pushHandler.RegisterRoutes(api, authMW)
	learningPathHandler.RegisterRoutes(api, authMW, optionalAuthMW, adminMW)
	reportHandler.RegisterRoutes(api, authMW, adminMW)

	// Swagger route
	a.echo.GET("/swagger/*", echoSwagger.WrapHandler)

	// Background worker for scheduled reports
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			ctx := context.Background()
			reportsDue, err := scheduledReportRepo.GetDueReports(ctx)
			if err != nil {
				continue
			}

			_ = exportSvc.ProcessScheduledReports(ctx, reportsDue, func(email string, data []byte, filename string) error {
				a.logger.Infof("Sending scheduled report %s to %s", filename, email)
				return nil
			})

			for _, r := range reportsDue {
				now := time.Now()
				next := export.GetScheduledReportNextRun(r.Schedule, now)
				_ = scheduledReportRepo.UpdateLastRun(ctx, r.ID, now, next)
			}
		}
	}()

	// Start server
	go func() {
		addr := ":" + a.cfg.Server.Port
		a.logger.Infof("Starting server on %s", addr)
		if err := a.echo.Start(addr); err != nil && err != http.ErrServerClosed {
			a.logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	a.logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.echo.Shutdown(ctx); err != nil {
		a.logger.Fatalf("Server forced to shutdown: %v", err)
	}

	a.logger.Info("Server exited properly")
	return nil
}

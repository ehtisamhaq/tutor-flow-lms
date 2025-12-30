package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/handler"
	appMiddleware "github.com/tutorflow/tutorflow-server/internal/middleware"
	"github.com/tutorflow/tutorflow-server/internal/pkg/config"
	"github.com/tutorflow/tutorflow-server/internal/pkg/database"
	"github.com/tutorflow/tutorflow-server/internal/pkg/jwt"
	"github.com/tutorflow/tutorflow-server/internal/repository/postgres"
	"github.com/tutorflow/tutorflow-server/internal/service/email"
	"github.com/tutorflow/tutorflow-server/internal/service/payment"
	"github.com/tutorflow/tutorflow-server/internal/service/storage"
	"github.com/tutorflow/tutorflow-server/internal/usecase/admin"
	"github.com/tutorflow/tutorflow-server/internal/usecase/auth"
	"github.com/tutorflow/tutorflow-server/internal/usecase/cart"
	"github.com/tutorflow/tutorflow-server/internal/usecase/certificate"
	"github.com/tutorflow/tutorflow-server/internal/usecase/course"
	"github.com/tutorflow/tutorflow-server/internal/usecase/discussion"
	"github.com/tutorflow/tutorflow-server/internal/usecase/enrollment"
	"github.com/tutorflow/tutorflow-server/internal/usecase/notification"
	"github.com/tutorflow/tutorflow-server/internal/usecase/order"
	"github.com/tutorflow/tutorflow-server/internal/usecase/quiz"
	"github.com/tutorflow/tutorflow-server/internal/usecase/review"
	"github.com/tutorflow/tutorflow-server/internal/usecase/search"
	"github.com/tutorflow/tutorflow-server/internal/usecase/user"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	if os.Getenv("APP_ENV") == "development" {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		sugar.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := database.Connect(cfg.Database)
	if err != nil {
		sugar.Fatalf("Failed to connect to database: %v", err)
	}
	sugar.Info("Connected to database")

	// Run migrations
	if err := database.AutoMigrate(db); err != nil {
		sugar.Fatalf("Failed to run migrations: %v", err)
	}
	sugar.Info("Database migrations completed")

	// Initialize Echo
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: cfg.Server.AllowedOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "X-Session-ID"},
	}))
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"time":"${time_rfc3339}","id":"${id}","method":"${method}","uri":"${uri}","status":${status},"latency":"${latency_human}"}` + "\n",
	}))

	// Serve static uploads
	e.Static("/uploads", cfg.Storage.LocalPath)

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "healthy",
			"version": cfg.Server.Version,
		})
	})

	// Initialize JWT Manager
	jwtManager := jwt.NewManager(cfg.JWT)

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

	// Initialize services
	storageSvc := storage.NewService(cfg.Storage)
	paymentSvc := payment.NewService(cfg.Stripe)
	emailSvc := email.NewService(cfg.Email)
	_ = emailSvc // Email service available for use in use cases

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
	searchUC := search.NewUseCase(searchRepo, courseRepo, categoryRepo)
	adminUC := admin.NewUseCase(db)

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

	// Middleware functions
	authMW := appMiddleware.AuthMiddleware(jwtManager)
	optionalAuthMW := appMiddleware.OptionalAuthMiddleware(jwtManager)
	adminMW := appMiddleware.RequireAdmin()
	managerMW := appMiddleware.RequireAdminOrManager()
	tutorMW := appMiddleware.RequireRole(domain.RoleAdmin, domain.RoleManager, domain.RoleTutor)

	// API v1 routes
	api := e.Group("/api/v1")

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

	// Start server
	go func() {
		addr := ":" + cfg.Server.Port
		sugar.Infof("Starting server on %s", addr)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			sugar.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	sugar.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		sugar.Fatalf("Server forced to shutdown: %v", err)
	}

	sugar.Info("Server exited properly")
}

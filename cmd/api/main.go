package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/4H1R/zoora/docs"
	"github.com/4H1R/zoora/internal/admin"
	"github.com/4H1R/zoora/internal/attendance"
	"github.com/4H1R/zoora/internal/auth"
	"github.com/4H1R/zoora/internal/platform/authz"
	"github.com/4H1R/zoora/internal/calendar"
	"github.com/4H1R/zoora/internal/chat"
	"github.com/4H1R/zoora/internal/classes"
	"github.com/4H1R/zoora/internal/gradebook"
	"github.com/4H1R/zoora/internal/config"
	"github.com/4H1R/zoora/internal/livesessions"
	"github.com/4H1R/zoora/internal/media"
	"github.com/4H1R/zoora/internal/middleware"
	"github.com/4H1R/zoora/internal/offlines"
	"github.com/4H1R/zoora/internal/organizations"
	"github.com/4H1R/zoora/internal/platform/cache"
	"github.com/4H1R/zoora/internal/platform/database"
	"github.com/4H1R/zoora/internal/platform/health"
	"github.com/4H1R/zoora/internal/platform/httpx"
	lk "github.com/4H1R/zoora/internal/platform/livekit"
	"github.com/4H1R/zoora/internal/platform/logger"
	"github.com/4H1R/zoora/internal/platform/queue"
	"github.com/4H1R/zoora/internal/platform/storage"
	"github.com/4H1R/zoora/internal/polls"
	"github.com/4H1R/zoora/internal/practices"
	"github.com/4H1R/zoora/internal/questionbanks"
	"github.com/4H1R/zoora/internal/quizzes"
	"github.com/4H1R/zoora/internal/roles"
	"github.com/4H1R/zoora/internal/users"
	// "github.com/4H1R/zoora/internal/websocket"
)

// @title Zoora API
// @version 1.0
// @description REST API for the Zoora education platform.
//
// @host localhost:8080
// @BasePath /api/v1
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your JWT token with the "Bearer " prefix, e.g. "Bearer eyJhbGci..."
func main() {
	log := logger.New(false)

	cfg, err := config.Load()
	if err != nil {
		log.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if err := httpx.RegisterValidators(); err != nil {
		log.Error("failed to register validators", "error", err)
		os.Exit(1)
	}

	if cfg.IsDevelopment() {
		log = logger.New(true)
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	db, err := database.NewConnection(cfg.DatabaseURL, log)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	redisClient, err := cache.NewRedisClient(cfg.RedisURL, log)
	if err != nil {
		log.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}

	storageClient, err := storage.NewClient(cfg, log)
	if err != nil {
		log.Error("failed to initialize storage", "error", err)
		os.Exit(1)
	}

	queueClient, err := queue.NewClient(cfg.RedisURL, log)
	if err != nil {
		log.Error("failed to initialize queue client", "error", err)
		os.Exit(1)
	}

	jwtService := auth.NewJWTService(cfg)

	userRepo := users.NewRepository(db)
	orgRepo := organizations.NewRepository(db)
	roleRepo := roles.NewRoleRepository(db)
	permRepo := roles.NewPermissionRepository(db)
	classRepo := classes.NewRepository(db)
	classSessionRepo := classes.NewSessionRepository(db)
	classMemberRepo := classes.NewMemberRepository(db)

	questionBankRepo := questionbanks.NewRepository(db)
	questionRepo := questionbanks.NewQuestionRepository(db)
	quizRepo := quizzes.NewRepository(db)
	quizRuleRepo := quizzes.NewRuleRepository(db)
	quizRoomRepo := quizzes.NewRoomRepository(db)
	quizSubmissionRepo := quizzes.NewSubmissionRepository(db)
	mediaRepo := media.NewRepository(db)
	liveRoomRepo := livesessions.NewRoomRepository(db)
	liveParticipantRepo := livesessions.NewParticipantRepository(db)
	liveRecordingRepo := livesessions.NewRecordingRepository(db)
	offlineRoomRepo := offlines.NewRoomRepository(db)
	offlineViewRepo := offlines.NewViewRepository(db)
	practiceRoomRepo := practices.NewRoomRepository(db)
	practiceSubRepo := practices.NewSubmissionRepository(db)
	pollRepo := polls.NewRepository(db)
	pollAnswerRepo := polls.NewAnswerRepository(db)
	chatRepo := chat.NewChatRepository(db)
	chatMemberRepo := chat.NewMemberRepository(db)
	chatMessageRepo := chat.NewMessageRepository(db)
	chatReactionRepo := chat.NewReactionRepository(db)

	authMiddleware := auth.Middleware(jwtService, redisClient, roleRepo, userRepo)

	authzResolver := authz.NewResolver(classMemberRepo)

	userService := users.NewService(userRepo, roleRepo, log)
	orgService := organizations.NewService(orgRepo, userRepo, log)
	classService := classes.NewService(classRepo, classSessionRepo, classMemberRepo, log)
	questionBankService := questionbanks.NewService(questionBankRepo, questionRepo, mediaRepo, log)
	quizService := quizzes.NewService(quizRepo, quizRuleRepo, quizRoomRepo, quizSubmissionRepo, questionRepo, classRepo, classMemberRepo, log)
	transactor := database.NewTransactor(db)

	// Reconcile the permissions table + preset-role grants with the code-defined
	// source of truth so renaming/removing a permission constant takes effect on
	// an existing DB without a destructive reseed.
	if err := roles.SyncPermissions(context.Background(), db, transactor, roleRepo, permRepo, redisClient, log); err != nil {
		log.Error("failed to sync permissions", "error", err)
		os.Exit(1)
	}

	roleService := roles.NewService(roleRepo, permRepo, transactor, redisClient, log)
	authBusinessService := auth.NewAuthService(userRepo, jwtService, redisClient, log)
	mediaService := media.NewService(mediaRepo, storageClient, log)
	livekitClient := lk.NewClient(cfg, log)
	chatService := chat.NewService(chatRepo, chatMemberRepo, chatMessageRepo, chatReactionRepo, transactor, log)
	liveSessionService := livesessions.NewService(
		liveRoomRepo, liveParticipantRepo, liveRecordingRepo,
		classSessionRepo, classRepo, classMemberRepo,
		chatService, transactor,
		livekitClient, log,
	)
	offlineService := offlines.NewService(offlineRoomRepo, offlineViewRepo, classSessionRepo, classRepo, classMemberRepo, log)
	practiceService := practices.NewService(practiceRoomRepo, practiceSubRepo, classSessionRepo, classRepo, classMemberRepo, log)
	pollService := polls.NewService(pollRepo, pollAnswerRepo, log)

	attendanceRepo := attendance.NewRepository(db)
	attendanceService := attendance.NewService(
		attendanceRepo, classRepo, classSessionRepo, classMemberRepo,
		liveRoomRepo, liveParticipantRepo, offlineViewRepo, offlineRoomRepo,
		authzResolver, log,
	)

	// wsHub := websocket.NewHub(log)
	// go wsHub.Run()

	healthChecker := health.NewChecker(db, redisClient, storageClient)

	router := gin.New()
	// Behind Traefik in the container network; trust private ranges so
	// X-Forwarded-For / client IP resolve correctly.
	if err := router.SetTrustedProxies([]string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	}); err != nil {
		log.Error("failed to set trusted proxies", "error", err)
		os.Exit(1)
	}
	router.Use(
		middleware.Recovery(log),
		middleware.ErrorHandler(log),
		middleware.Logging(log),
		middleware.CORS(cfg.CORSAllowedOrigins),
		middleware.GlobalRateLimit(redisClient),
	)

	router.GET("/healthz", healthChecker.LivenessHandler)
	router.GET("/readyz", healthChecker.ReadinessHandler)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")

	authHandler := auth.NewHandler(authBusinessService)
	authHandler.RegisterRoutes(v1, middleware.AuthRateLimit(redisClient))

	perm := auth.RequirePermission
	permAny := auth.RequireAnyPermission

	orgHandler := organizations.NewHandler(orgService)
	orgHandler.RegisterRoutes(v1, authMiddleware, perm)

	userHandler := users.NewHandler(userService)
	userHandler.RegisterRoutes(v1, authMiddleware, perm)

	roleHandler := roles.NewHandler(roleService, permRepo)
	roleHandler.RegisterRoutes(v1, authMiddleware, perm)

	classHandler := classes.NewHandler(classService)
	classHandler.RegisterRoutes(v1, authMiddleware, perm)

	questionBankHandler := questionbanks.NewHandler(questionBankService)
	questionBankHandler.RegisterRoutes(v1, authMiddleware, perm)

	quizHandler := quizzes.NewHandler(quizService)
	quizHandler.RegisterRoutes(v1, authMiddleware, perm, permAny)
	mediaHandler := media.NewHandler(mediaService)
	mediaHandler.RegisterRoutes(v1, authMiddleware, perm)

	liveSessionHandler := livesessions.NewHandler(liveSessionService)
	liveSessionHandler.RegisterRoutes(v1, authMiddleware, perm)

	calendarRepo := calendar.NewRepository(db)
	calendarService := calendar.NewService(calendarRepo, log)
	calendarHandler := calendar.NewHandler(calendarService)
	calendarHandler.RegisterRoutes(v1, authMiddleware)

	offlineHandler := offlines.NewHandler(offlineService)
	offlineHandler.RegisterRoutes(v1, authMiddleware, perm)

	practiceHandler := practices.NewHandler(practiceService)
	practiceHandler.RegisterRoutes(v1, authMiddleware, perm)

	pollHandler := polls.NewHandler(pollService)
	pollHandler.RegisterRoutes(v1, authMiddleware, perm)

	chatHandler := chat.NewHandler(chatService)
	chatHandler.RegisterRoutes(v1, authMiddleware, perm)

	attendanceHandler := attendance.NewHandler(attendanceService)
	attendanceHandler.RegisterRoutes(v1, authMiddleware, perm)

	gradebookColRepo := gradebook.NewColumnRepository(db)
	gradebookCellRepo := gradebook.NewCellRepository(db)
	gradebookService := gradebook.NewService(
		gradebookColRepo, gradebookCellRepo,
		classRepo, classMemberRepo,
		attendanceRepo, practiceSubRepo, quizSubmissionRepo,
		authzResolver, log,
	)
	gradebookHandler := gradebook.NewHandler(gradebookService)
	gradebookHandler.RegisterRoutes(v1, authMiddleware, perm)

	// Admin route tree: /api/v1/admin/*
	adminUserHandler := users.NewAdminHandler(userService, authBusinessService)
	adminOrgHandler := organizations.NewAdminHandler(orgService)
	adminClassHandler := classes.NewAdminHandler(classService)

	adminQuestionBankHandler := questionbanks.NewAdminHandler(questionBankService)
	adminQuizHandler := quizzes.NewAdminHandler(quizService)
	adminLiveSessionHandler := livesessions.NewAdminHandler(liveSessionService)
	adminOfflineHandler := offlines.NewAdminHandler(offlineService)
	adminPracticeHandler := practices.NewAdminHandler(practiceService)
	adminPollHandler := polls.NewAdminHandler(pollService)
	adminRoleHandler := roles.NewAdminHandler(roleService)
	adminAttendanceHandler := attendance.NewAdminHandler(attendanceService)

	adminGroup := v1.Group("/admin", authMiddleware, auth.RequireAdmin())
	admin.RegisterRoutes(adminGroup, adminUserHandler, adminOrgHandler, adminClassHandler, adminQuestionBankHandler, adminQuizHandler, adminLiveSessionHandler, adminOfflineHandler, adminPracticeHandler, adminPollHandler, adminRoleHandler, adminAttendanceHandler)

	// router.GET("/ws/:room", websocket.HandleWebSocket(wsHub, jwtService, log))

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("starting API server", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server shutdown error", "error", err)
	}
	// wsHub.Shutdown()
	queueClient.Close()

	sqlDB, _ := db.DB()
	sqlDB.Close()
	redisClient.Close()

	log.Info("server stopped")
}

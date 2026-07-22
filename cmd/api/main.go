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
	"github.com/4H1R/zoora/internal/audit"
	"github.com/4H1R/zoora/internal/auth"
	"github.com/4H1R/zoora/internal/billing"
	"github.com/4H1R/zoora/internal/calendar"
	"github.com/4H1R/zoora/internal/changelog"
	"github.com/4H1R/zoora/internal/chat"
	"github.com/4H1R/zoora/internal/chathub"
	"github.com/4H1R/zoora/internal/classes"
	"github.com/4H1R/zoora/internal/config"
	"github.com/4H1R/zoora/internal/connectors"
	"github.com/4H1R/zoora/internal/conversations"
	"github.com/4H1R/zoora/internal/customfields"
	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/entitlements"
	"github.com/4H1R/zoora/internal/gradebook"
	"github.com/4H1R/zoora/internal/imports"
	"github.com/4H1R/zoora/internal/leads"
	"github.com/4H1R/zoora/internal/livesessions"
	"github.com/4H1R/zoora/internal/media"
	"github.com/4H1R/zoora/internal/middleware"
	"github.com/4H1R/zoora/internal/notifications"
	"github.com/4H1R/zoora/internal/offlines"
	"github.com/4H1R/zoora/internal/organizations"
	"github.com/4H1R/zoora/internal/orgsettings"
	"github.com/4H1R/zoora/internal/platform/authz"
	"github.com/4H1R/zoora/internal/platform/cache"
	"github.com/4H1R/zoora/internal/platform/database"
	"github.com/4H1R/zoora/internal/platform/health"
	"github.com/4H1R/zoora/internal/platform/httpx"
	lk "github.com/4H1R/zoora/internal/platform/livekit"
	"github.com/4H1R/zoora/internal/platform/logger"
	"github.com/4H1R/zoora/internal/platform/observability"
	"github.com/4H1R/zoora/internal/platform/payment"
	"github.com/4H1R/zoora/internal/platform/queue"
	"github.com/4H1R/zoora/internal/platform/sms"
	"github.com/4H1R/zoora/internal/platform/storage"
	"github.com/4H1R/zoora/internal/polls"
	"github.com/4H1R/zoora/internal/practices"
	"github.com/4H1R/zoora/internal/qa"
	"github.com/4H1R/zoora/internal/questionbanks"
	"github.com/4H1R/zoora/internal/quizzes"
	"github.com/4H1R/zoora/internal/roles"
	"github.com/4H1R/zoora/internal/tickets"
	"github.com/4H1R/zoora/internal/tutorials"
	"github.com/4H1R/zoora/internal/users"
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
	log := logger.New(false, "")

	cfg, err := config.Load()
	if err != nil {
		log.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if err := httpx.RegisterValidators(); err != nil {
		log.Error("failed to register validators", "error", err)
		os.Exit(1)
	}

	log = logger.New(cfg.IsDevelopment(), cfg.LogLevel)
	if cfg.IsDevelopment() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	db, err := database.NewConnection(cfg.DatabaseURL, cfg.DatabaseReplicaURL, database.PoolConfig{
		MaxOpenConns:    cfg.DBMaxOpenConns,
		MaxIdleConns:    cfg.DBMaxIdleConns,
		ConnMaxLifetime: cfg.DBConnMaxLifetime,
		ConnMaxIdleTime: cfg.DBConnMaxIdleTime,
	}, log, cfg.IsDevelopment())
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	// Cache-role Redis: auth, tenant resolve, rate-limit, entity caches. Falls
	// back to the shared instance until REDIS_CACHE_URL is set.
	redisClient, err := cache.NewRedisClient(cfg.CacheRedisURL(), log)
	if err != nil {
		log.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}

	// Pub/sub-role Redis: WebSocket fan-out bridge + presence. Split from the
	// cache instance via REDIS_PUBSUB_URL when realtime traffic needs isolation.
	pubsubRedisClient, err := cache.NewRedisClient(cfg.PubSubRedisURL(), log)
	if err != nil {
		log.Error("failed to connect to pub/sub redis", "error", err)
		os.Exit(1)
	}

	storageClient, err := storage.NewClient(cfg, log)
	if err != nil {
		log.Error("failed to initialize storage", "error", err)
		os.Exit(1)
	}

	queueClient, err := queue.NewClient(cfg.QueueRedisURL(), log)
	if err != nil {
		log.Error("failed to initialize queue client", "error", err)
		os.Exit(1)
	}

	jwtService := auth.NewJWTService(cfg)

	userRepo := users.NewRepository(db)
	customFieldRepo := customfields.NewRepository(db)
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
	changelogRepo := changelog.NewRepository(db)
	tutorialRepo := tutorials.NewRepository(db)
	liveRoomRepo := livesessions.NewRoomRepository(db)
	liveParticipantRepo := livesessions.NewParticipantRepository(db)
	liveRecordingRepo := livesessions.NewRecordingRepository(db)
	liveWhiteboardRepo := livesessions.NewWhiteboardRepository(db)
	offlineRoomRepo := offlines.NewRoomRepository(db)
	offlineViewRepo := offlines.NewViewRepository(db)
	practiceRoomRepo := practices.NewRoomRepository(db)
	practiceSubRepo := practices.NewSubmissionRepository(db)
	pollRepo := polls.NewRepository(db)
	pollAnswerRepo := polls.NewAnswerRepository(db)
	qaRepo := qa.NewRepository(db)
	qaVoteRepo := qa.NewVoteRepository(db)
	chatRepo := chat.NewChatRepository(db)
	chatMessageRepo := chat.NewMessageRepository(db)
	convRepo := conversations.NewConversationRepository(db)
	convMemberRepo := conversations.NewMemberRepository(db)
	convMessageRepo := conversations.NewMessageRepository(db)
	convReactionRepo := conversations.NewReactionRepository(db)
	convMentionRepo := conversations.NewMentionRepository(db)
	entitlementRepo := entitlements.NewRepository(db)
	entitlementService := entitlements.NewService(entitlementRepo)

	importRepo := imports.NewRepository(db)
	importService := imports.NewService(
		importRepo, userRepo, roleRepo, classRepo, classMemberRepo, mediaRepo,
		entitlementService, storageClient, queueClient,
		imports.NewRedisResultStore(redisClient), log,
	)

	authMiddleware := auth.Middleware(jwtService, redisClient, roleRepo, userRepo, entitlementRepo)
	tenantMiddleware := middleware.Tenant(redisClient, orgRepo, cfg.BaseDomain, cfg.AdminSubdomain)

	authzResolver := authz.NewResolver(classMemberRepo)

	orgSettingsRepo := orgsettings.NewRepository(db)
	orgSettingsService := orgsettings.NewService(orgSettingsRepo, log)

	sessionManager := auth.NewSessionManager(jwtService, redisClient)
	customFieldService := customfields.NewService(customFieldRepo, log)
	orgService := organizations.NewService(orgRepo, userRepo, orgSettingsRepo, redisClient, queueClient, log)
	transactor := database.NewTransactor(db)

	leadRepo := leads.NewRepository(db)
	leadService := leads.NewService(leadRepo, orgRepo, orgSettingsRepo, userRepo, roleRepo, transactor, log)

	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo, log)

	questionBankService := questionbanks.NewService(questionBankRepo, questionRepo, mediaRepo, queueClient, transactor, auditService, log)
	quizService := quizzes.NewService(quizRepo, quizRuleRepo, quizRoomRepo, quizSubmissionRepo, questionRepo, classRepo, classMemberRepo, queueClient, transactor, auditService, log)

	userService := users.NewService(userRepo, roleRepo, entitlementService, redisClient, sessionManager, transactor, auditService, log)

	// Reconcile the permissions table + preset-role grants with the code-defined
	// source of truth so renaming/removing a permission constant takes effect on
	// an existing DB without a destructive reseed.
	if err := roles.SyncPermissions(context.Background(), db, transactor, roleRepo, permRepo, redisClient, log); err != nil {
		log.Error("failed to sync permissions", "error", err)
		os.Exit(1)
	}

	roleService := roles.NewService(roleRepo, permRepo, transactor, auditService, redisClient, log)
	authBusinessService := auth.NewAuthService(userRepo, jwtService, redisClient, log)
	mediaService := media.NewService(mediaRepo, storageClient, entitlementService, entitlementRepo, log)
	changelogService := changelog.NewServiceWithMedia(changelogRepo, mediaRepo, storageClient, redisClient, log)
	tutorialService := tutorials.NewService(tutorialRepo, log)
	livekitClient := lk.NewClient(cfg, log)
	modelAuthorizer := livesessions.NewModelAuthorizer(liveRoomRepo, classSessionRepo, classRepo, classMemberRepo)
	chatService := chat.NewService(chatRepo, chatMessageRepo, transactor, log, livekitClient, liveRoomRepo, modelAuthorizer)

	convHubMembership := conversations.NewHubMembership(convMemberRepo)
	convHub := chathub.NewHub(convHubMembership, log)
	convBridge := chathub.NewBridge(convHub, pubsubRedisClient, log)
	convPresence := chathub.NewPresence(pubsubRedisClient, chathub.PresenceTTL)
	go convBridge.Run(context.Background())

	pollModelAuthorizer := polls.NewModelAuthorizer(liveRoomRepo, classSessionRepo, classRepo, classMemberRepo)
	pollService := polls.NewService(pollRepo, pollAnswerRepo, pollModelAuthorizer, log)
	qaBroadcaster := qa.NewBroadcaster(livekitClient, liveRoomRepo, log)
	qaService := qa.NewService(qaRepo, qaVoteRepo, modelAuthorizer, log, qaBroadcaster)
	liveSessionService := livesessions.NewService(
		liveRoomRepo, liveParticipantRepo, liveRecordingRepo, liveWhiteboardRepo,
		classSessionRepo, classRepo, classMemberRepo,
		chatService, pollService, transactor,
		livekitClient, storageClient, queueClient, entitlementService, cfg.LiveRoomHostGracePeriod, cfg.LiveKitEgressMaxConcurrent, log,
	)
	offlineService := offlines.NewService(offlineRoomRepo, offlineViewRepo, classSessionRepo, classRepo, classMemberRepo, queueClient, transactor, auditService, log)
	practiceService := practices.NewService(practiceRoomRepo, practiceSubRepo, classSessionRepo, classRepo, classMemberRepo, transactor, auditService, log)

	attendanceRepo := attendance.NewRepository(db)
	attendanceService := attendance.NewService(
		attendanceRepo, classRepo, classSessionRepo, classMemberRepo,
		liveRoomRepo, liveParticipantRepo, offlineViewRepo, offlineRoomRepo,
		orgSettingsService, authzResolver, log,
	)

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
	// Load-testing escape hatch; never honored in production.
	rateLimitDisabled := cfg.RateLimitDisabled && !cfg.IsProduction()
	if cfg.RateLimitDisabled && cfg.IsProduction() {
		log.Warn("RATE_LIMIT_DISABLED is set but ignored in production")
	}
	if rateLimitDisabled {
		log.Warn("rate limiting is DISABLED (RATE_LIMIT_DISABLED=true)")
	}

	// Optional Sentry error reporting. No-op (empty handlers) when SENTRY_DSN is
	// unset, so the app runs unchanged until keys are added. Registered after
	// Recovery so panics are captured and then re-raised for Recovery to answer.
	sentryFlush, sentryHandlers := observability.InitSentry(cfg, log)
	defer sentryFlush()

	router.Use(middleware.RequestID(), middleware.RequestInfo(), middleware.Recovery(log))
	router.Use(sentryHandlers...)
	router.Use(
		middleware.ErrorHandler(log),
		middleware.Logging(log),
		middleware.CORS(cfg.CORSAllowedOrigins),
		middleware.GlobalRateLimit(redisClient, rateLimitDisabled),
	)

	router.GET("/healthz", healthChecker.LivenessHandler)
	router.GET("/readyz", healthChecker.ReadinessHandler)
	// Caddy on-demand TLS gate: only mint certs for real tenant subdomains.
	router.GET("/internal/tls-check", middleware.OnDemandTLSCheck(redisClient, orgRepo, cfg.BaseDomain, cfg.AdminSubdomain))
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// LiveKit server webhooks: unauthenticated + non-tenant-scoped. Auth is the
	// signed Authorization header (verified in the handler); rooms resolve by
	// globally-unique LiveKit room name. Must sit outside the tenant group.
	liveWebhookHandler := livesessions.NewWebhookHandler(livekitClient, liveSessionService, log)
	liveWebhookHandler.RegisterRoutes(router.Group("/webhooks"))

	v1 := router.Group("/api/v1", tenantMiddleware, middleware.AuditDenied(auditService, log))

	authHandler := auth.NewHandler(authBusinessService)
	authHandler.RegisterRoutes(v1, middleware.AuthRateLimit(redisClient, rateLimitDisabled))

	// Public, unauthenticated lead capture from the marketing site. Mounted on
	// v1 (tenant middleware only injects context, never rejects), so it works
	// from the apex host. Rate-limited + honeypot-gated against abuse.
	leadHandler := leads.NewHandler(leadService)
	leadHandler.RegisterRoutes(v1, middleware.LeadRateLimit(redisClient, rateLimitDisabled))

	perm := auth.RequirePermission
	permAny := auth.RequireAnyPermission

	orgHandler := organizations.NewHandler(orgService)
	orgHandler.RegisterRoutes(v1, authMiddleware, perm)

	orgSettingsHandler := orgsettings.NewHandler(orgSettingsService)
	orgSettingsHandler.RegisterRoutes(v1, authMiddleware, perm)

	userHandler := users.NewHandler(userService)
	userHandler.RegisterRoutes(v1, authMiddleware, perm)

	customFieldHandler := customfields.NewHandler(customFieldService)
	customFieldHandler.RegisterRoutes(v1, authMiddleware, perm)

	roleHandler := roles.NewHandler(roleService, permRepo)
	roleHandler.RegisterRoutes(v1, authMiddleware, perm)

	importHandler := imports.NewHandler(importService)
	importHandler.RegisterRoutes(v1, authMiddleware, perm)

	questionBankHandler := questionbanks.NewHandler(questionBankService)
	questionBankHandler.RegisterRoutes(v1, authMiddleware, perm)

	quizHandler := quizzes.NewHandler(quizService)
	quizHandler.RegisterRoutes(v1, authMiddleware, perm, permAny)
	mediaHandler := media.NewHandler(mediaService)
	mediaHandler.RegisterRoutes(v1, authMiddleware, perm)

	changelogHandler := changelog.NewHandler(changelogService)
	changelogHandler.RegisterRoutes(v1, authMiddleware)

	tutorialHandler := tutorials.NewHandler(tutorialService)
	tutorialHandler.RegisterRoutes(v1, authMiddleware)

	// SMS channel is optional. When unconfigured, OTP linking and SMS delivery
	// are disabled (nil sender surfaces a validation error / skips the channel).
	var smsSender domain.SMSSender
	if cfg.KavenegarAPIKey != "" {
		smsSender = sms.NewKavenegar(sms.Config{
			APIKey:      cfg.KavenegarAPIKey,
			Sender:      cfg.KavenegarSender,
			OTPTemplate: cfg.KavenegarOTPTemplate,
		}, log)
	}

	connectorRepo := connectors.NewRepository(db)
	connectorService := connectors.NewService(connectorRepo, userRepo, orgRepo, redisClient, smsSender, connectors.BotLinkConfig{
		TelegramBotUsername: cfg.TelegramBotUsername,
		BaleBotUsername:     cfg.BaleBotUsername,
	}, log)
	connectorHandler := connectors.NewHandler(connectorService)
	connectorHandler.RegisterRoutes(v1, authMiddleware)

	notificationRepo := notifications.NewRepository(db)
	// API only enqueues; the worker owns bot/push sends. SMS sender is wired
	// here too so a manual re-enqueue path could use it, but delivery runs
	// worker-side.
	notificationService := notifications.NewService(
		notificationRepo, classRepo, connectorRepo, orgSettingsService, orgRepo,
		queueClient, notifications.Senders{SMS: smsSender}, cfg.NotificationSendRatePerHour, redisClient, log,
	)
	notificationHandler := notifications.NewHandler(notificationService)
	notificationHandler.RegisterRoutes(v1, authMiddleware)

	convNotifier := conversations.NewNotifier(notificationService)
	convUserLookup := conversations.NewUserOrgLookup(userRepo)
	conversationService := conversations.NewService(
		convRepo, convMemberRepo, convMessageRepo, convReactionRepo, convMentionRepo,
		transactor, log,
		convBridge,     // broadcaster (implements ToConversation/ToUser/ToUsers)
		convNotifier,   // notifier (SendSystem fan-out)
		convUserLookup, // userDirectory (cross-org DM/member guard, required)
		conversations.NewAttachmentValidator(mediaRepo), // attachmentValidator (required)
		presenceReaderAdapter{p: convPresence},          // presenceReader (batch online/last-seen)
		queueClient,                                     // enqueuer (attachment cleanup on delete)
	)

	// classService depends on the conversations service (ClassChatProvisioner) to
	// provision a class's group/channel chat, so it is constructed after it.
	classService := classes.NewService(classRepo, classSessionRepo, classMemberRepo, conversationService, transactor, auditService, log)
	classHandler := classes.NewHandler(classService)
	classHandler.RegisterRoutes(v1, authMiddleware, perm)

	// --- billing ---
	zpBase := "https://payment.zarinpal.com"
	if cfg.ZarinpalSandbox {
		zpBase = "https://sandbox.zarinpal.com"
	}
	gatewayRegistry := payment.NewRegistry(
		payment.NewZarinpal(payment.ZarinpalConfig{
			MerchantID: cfg.ZarinpalMerchantID,
			BaseURL:    zpBase,
		}),
	)
	billingIssuer := billing.IssuerConfig{
		Name:       cfg.InvoiceIssuerName,
		EconomicID: cfg.InvoiceIssuerEconomicID,
		Address:    cfg.InvoiceIssuerAddress,
		Phone:      cfg.InvoiceIssuerPhone,
	}
	billingRepo := billing.NewRepository(db)
	billingPDF := billing.NewPDFRenderer(storageClient, orgRepo, billingIssuer, cfg.ChromeRemoteURL)
	billingSvc := billing.NewService(
		billingRepo,
		orgRepo, // domain.OrganizationRepository
		orgRepo, // planActivator (UpdatePlan)
		billing.NewEntitlementsCacheBuster(redisClient), // entitlementsCacheBuster
		gatewayRegistry,
		storageClient,                         // objectStorage (presign)
		billing.NewQueueEnqueuer(queueClient), // enqueuer
		notificationService,                   // systemNotifier (SendSystem)
		billingPDF,
		billing.BillingConfig{
			CallbackBaseURL: cfg.ZarinpalCallbackBaseURL,
			AppURLTemplate:  cfg.AppURLTemplate,
			Issuer:          billingIssuer,
		},
		log,
	)
	billingAdminSvc := billing.NewAdminService(billingSvc)
	billingHandler := billing.NewHandler(billingSvc, cfg.AppURLTemplate)
	billingHandler.RegisterRoutes(v1, authMiddleware, perm)
	billingAdminHandler := billing.NewAdminHandler(billingAdminSvc)

	liveSessionHandler := livesessions.NewHandler(liveSessionService)
	liveSessionHandler.RegisterRoutes(v1, authMiddleware, perm)

	calendarRepo := calendar.NewRepository(db)
	calendarService := calendar.NewService(calendarRepo, redisClient, log)
	calendarHandler := calendar.NewHandler(calendarService)
	calendarHandler.RegisterRoutes(v1, authMiddleware)

	offlineHandler := offlines.NewHandler(offlineService)
	offlineHandler.RegisterRoutes(v1, authMiddleware, perm)

	practiceHandler := practices.NewHandler(practiceService)
	practiceHandler.RegisterRoutes(v1, authMiddleware, perm)

	pollHandler := polls.NewHandler(pollService)
	pollHandler.RegisterRoutes(v1, authMiddleware, perm)

	qaHandler := qa.NewHandler(qaService)
	qaHandler.RegisterRoutes(v1, authMiddleware, perm)

	chatHandler := chat.NewHandler(chatService)
	chatHandler.RegisterRoutes(v1, authMiddleware, perm)

	conversationHandler := conversations.NewHandler(conversationService)
	conversationHandler.RegisterRoutes(v1, authMiddleware, perm)
	v1.GET("/ws", chathub.HandleWS(convHub, convBridge, convPresence, jwtService, redisClient, cfg.CORSAllowedOrigins, log))

	attendanceHandler := attendance.NewHandler(attendanceService)
	attendanceHandler.RegisterRoutes(v1, authMiddleware, perm)

	auditHandler := audit.NewHandler(auditService)
	auditHandler.RegisterRoutes(v1, authMiddleware, perm)

	gradebookColRepo := gradebook.NewColumnRepository(db)
	gradebookCellRepo := gradebook.NewCellRepository(db)
	gradebookService := gradebook.NewService(
		gradebookColRepo, gradebookCellRepo,
		classRepo, classMemberRepo,
		attendanceRepo, practiceSubRepo, quizSubmissionRepo,
		quizRepo, practiceRoomRepo, classSessionRepo,
		authzResolver, transactor, auditService, log,
	)
	gradebookHandler := gradebook.NewHandler(gradebookService)
	gradebookHandler.RegisterRoutes(v1, authMiddleware, perm)

	ticketRepo := tickets.NewRepository(db)
	ticketMessageRepo := tickets.NewMessageRepository(db)
	ticketNotifier := tickets.NewNotifier(notificationService)
	ticketService := tickets.NewService(
		ticketRepo, ticketMessageRepo,
		classRepo,        // classLookup
		classMemberRepo,  // memberLookup
		classSessionRepo, // sessionLookup (quiz-room -> class validation)
		quizRoomRepo,     // quizRoomLookup
		gradebookColRepo, // columnLookup
		mediaRepo,        // mediaLookup (attachment validation)
		transactor, ticketNotifier, log,
	)
	ticketHandler := tickets.NewHandler(ticketService)
	ticketHandler.RegisterRoutes(v1, authMiddleware, perm)

	adminUserHandler := users.NewAdminHandler(userService, authBusinessService)
	adminOrgHandler := organizations.NewAdminHandler(orgService)
	adminClassHandler := classes.NewAdminHandler(classService)

	adminQuestionBankHandler := questionbanks.NewAdminHandler(questionBankService)
	adminQuizHandler := quizzes.NewAdminHandler(quizService)
	adminLiveSessionHandler := livesessions.NewAdminHandler(liveSessionService)
	adminOfflineHandler := offlines.NewAdminHandler(offlineService)
	adminPracticeHandler := practices.NewAdminHandler(practiceService)
	adminPollHandler := polls.NewAdminHandler(pollService)
	adminQAHandler := qa.NewAdminHandler(qaService)
	adminRoleHandler := roles.NewAdminHandler(roleService)
	adminAttendanceHandler := attendance.NewAdminHandler(attendanceService)
	adminChangelogHandler := changelog.NewAdminHandler(changelogService)
	adminTutorialHandler := tutorials.NewAdminHandler(tutorialService)
	adminOrgSettingsHandler := orgsettings.NewAdminHandler(orgSettingsService)
	adminLeadHandler := leads.NewAdminHandler(leadService)

	adminGroup := v1.Group("/admin", authMiddleware, auth.RequireAdmin())
	admin.RegisterRoutes(adminGroup, adminUserHandler, adminOrgHandler, adminClassHandler, adminQuestionBankHandler, adminQuizHandler, adminLiveSessionHandler, adminOfflineHandler, adminPracticeHandler, adminPollHandler, adminQAHandler, adminRoleHandler, adminAttendanceHandler, adminChangelogHandler, adminTutorialHandler, adminOrgSettingsHandler, billingAdminHandler, adminLeadHandler)

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
	queueClient.Close()

	sqlDB, _ := db.DB()
	sqlDB.Close()
	redisClient.Close()
	pubsubRedisClient.Close()

	log.Info("server stopped")
}

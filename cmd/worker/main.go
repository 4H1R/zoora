package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"

	"github.com/4H1R/zoora/internal/ai"
	"github.com/4H1R/zoora/internal/attendance"
	"github.com/4H1R/zoora/internal/audit"
	"github.com/4H1R/zoora/internal/billing"
	"github.com/4H1R/zoora/internal/chat"
	"github.com/4H1R/zoora/internal/classes"
	"github.com/4H1R/zoora/internal/config"
	"github.com/4H1R/zoora/internal/connectors"
	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/entitlements"
	"github.com/4H1R/zoora/internal/imports"
	"github.com/4H1R/zoora/internal/livesessions"
	"github.com/4H1R/zoora/internal/media"
	"github.com/4H1R/zoora/internal/notifications"
	"github.com/4H1R/zoora/internal/offlines"
	"github.com/4H1R/zoora/internal/organizations"
	"github.com/4H1R/zoora/internal/orgsettings"
	"github.com/4H1R/zoora/internal/platform/authz"
	"github.com/4H1R/zoora/internal/platform/bots"
	"github.com/4H1R/zoora/internal/platform/cache"
	"github.com/4H1R/zoora/internal/platform/database"
	lk "github.com/4H1R/zoora/internal/platform/livekit"
	"github.com/4H1R/zoora/internal/platform/llm"
	"github.com/4H1R/zoora/internal/platform/logger"
	"github.com/4H1R/zoora/internal/platform/payment"
	"github.com/4H1R/zoora/internal/platform/push"
	"github.com/4H1R/zoora/internal/platform/queue"
	"github.com/4H1R/zoora/internal/platform/sms"
	"github.com/4H1R/zoora/internal/platform/storage"
	"github.com/4H1R/zoora/internal/polls"
	"github.com/4H1R/zoora/internal/questionbanks"
	"github.com/4H1R/zoora/internal/quizzes"
	"github.com/4H1R/zoora/internal/roles"
	"github.com/4H1R/zoora/internal/users"
)

func main() {
	log := logger.New(false, "")

	cfg, err := config.Load()
	if err != nil {
		log.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	log = logger.New(cfg.IsDevelopment(), cfg.LogLevel)

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

	queueServer, err := queue.NewServer(cfg.QueueRedisURL(), log)
	if err != nil {
		log.Error("failed to initialize queue server", "error", err)
		os.Exit(1)
	}

	queueClient, err := queue.NewClient(cfg.QueueRedisURL(), log)
	if err != nil {
		log.Error("failed to initialize queue client", "error", err)
		os.Exit(1)
	}

	transactor := database.NewTransactor(db)
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo, log)
	chatRepo := chat.NewChatRepository(db)
	chatMessageRepo := chat.NewMessageRepository(db)
	// Worker has no LiveKit client; realtime chat broadcast is API-only (nil deps = no-op).
	// modelAuth is nil too: the worker only calls FindChatByRoom/ArchiveByRoom on the
	// chat service (already room-authorized upstream), never the authz-gated methods.
	chatSvc := chat.NewService(chatRepo, chatMessageRepo, transactor, log, nil, nil, nil)

	liveRoomRepo := livesessions.NewRoomRepository(db)
	liveParticipantRepo := livesessions.NewParticipantRepository(db)
	liveRecordingRepo := livesessions.NewRecordingRepository(db)
	liveWhiteboardRepo := livesessions.NewWhiteboardRepository(db)
	classSessionRepo := classes.NewSessionRepository(db)
	classRepo := classes.NewRepository(db)
	classMemberRepo := classes.NewMemberRepository(db)
	livekitClient := lk.NewClient(cfg, log)
	pollRepo := polls.NewRepository(db)
	pollAnswerRepo := polls.NewAnswerRepository(db)
	pollModelAuthorizer := polls.NewModelAuthorizer(liveRoomRepo, classSessionRepo, classRepo, classMemberRepo)
	pollSvc := polls.NewService(pollRepo, pollAnswerRepo, pollModelAuthorizer, transactor, auditService, log)
	liveSessionService := livesessions.NewService(
		liveRoomRepo, liveParticipantRepo, liveRecordingRepo, liveWhiteboardRepo,
		classSessionRepo, classRepo, classMemberRepo,
		chatSvc, pollSvc, transactor, auditService,
		livekitClient, nil, queueClient, nil, cfg.LiveRoomHostGracePeriod, 0, log,
	)
	queueServer.HandleFunc(domain.TypeLiveSessionAutoClose, livesessions.NewAutoCloseHandler(liveSessionService))
	queueServer.HandleFunc(domain.TypeLiveSessionCloseIfNoHost, livesessions.NewCloseIfNoHostHandler(liveSessionService))

	attendanceRepo := attendance.NewRepository(db)
	offlineRoomRepo := offlines.NewRoomRepository(db)
	offlineViewRepo := offlines.NewViewRepository(db)
	orgSettingsRepo := orgsettings.NewRepository(db)
	orgSettingsService := orgsettings.NewService(orgSettingsRepo, transactor, auditService, log)
	authzResolver := authz.NewResolver(classMemberRepo)
	attendanceService := attendance.NewService(
		attendanceRepo, classRepo, classSessionRepo, classMemberRepo,
		liveRoomRepo, liveParticipantRepo, offlineViewRepo, offlineRoomRepo,
		orgSettingsService, authzResolver, transactor, auditService, log,
	)
	queueServer.HandleFunc(domain.TypeAttendanceAutoMark, attendance.NewAutoMarkHandler(attendanceService))

	// --- notification delivery channels (all optional; empty disables one) ---
	var telegramBot, baleBot *bots.Client
	if cfg.TelegramBotToken != "" {
		telegramBot, err = bots.NewClient(bots.Config{BaseURL: "https://api.telegram.org", Token: cfg.TelegramBotToken, ProxyURL: cfg.TelegramProxyURL}, log)
		if err != nil {
			log.Error("telegram bot init failed", "error", err)
			os.Exit(1)
		}
	}
	if cfg.BaleBotToken != "" {
		baleBot, err = bots.NewClient(bots.Config{BaseURL: "https://tapi.bale.ai", Token: cfg.BaleBotToken, ProxyURL: cfg.BaleProxyURL}, log)
		if err != nil {
			log.Error("bale bot init failed", "error", err)
			os.Exit(1)
		}
	}
	var smsSender domain.SMSSender
	if cfg.KavenegarAPIKey != "" {
		smsSender = sms.NewKavenegar(sms.Config{APIKey: cfg.KavenegarAPIKey, Sender: cfg.KavenegarSender, OTPTemplate: cfg.KavenegarOTPTemplate}, log)
	}
	var pushSender domain.PushSender
	if cfg.FCMCredentialsFile != "" {
		pushSender, err = push.NewFCM(context.Background(), cfg.FCMCredentialsFile, log)
		if err != nil {
			log.Error("fcm init failed", "error", err)
			os.Exit(1)
		}
	}

	// Interface-nil pitfall: assigning a nil *bots.Client into an interface
	// field yields a non-nil interface. Assign bot senders conditionally.
	senders := notifications.Senders{SMS: smsSender, Push: pushSender}
	if telegramBot != nil {
		senders.Telegram = telegramBot
	}
	if baleBot != nil {
		senders.Bale = baleBot
	}

	// Redis (cache role) is created up front: fan-out invalidates recipients'
	// cached unread counts, and the bot pollers below need link tokens.
	redisClient, err := cache.NewRedisClient(cfg.CacheRedisURL(), log)
	if err != nil {
		log.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}

	connectorRepo := connectors.NewRepository(db)
	orgRepo := organizations.NewRepository(db)
	notificationRepo := notifications.NewRepository(db)
	notificationService := notifications.NewService(
		notificationRepo, classRepo, connectorRepo, orgSettingsService, orgRepo,
		queueClient, senders, 0, redisClient, log,
	)
	queueServer.HandleFunc(domain.TypeNotificationFanout, notifications.NewFanoutHandler(notificationService))
	queueServer.HandleFunc(domain.TypeNotificationDeliverBot, notifications.NewDeliverBotHandler(notificationService))
	queueServer.HandleFunc(domain.TypeNotificationDeliverSMS, notifications.NewDeliverSMSHandler(notificationService))
	queueServer.HandleFunc(domain.TypeNotificationDeliverPush, notifications.NewDeliverPushHandler(notificationService))

	userRepo := users.NewRepository(db)
	connectorService := connectors.NewService(connectorRepo, userRepo, orgRepo, redisClient, smsSender, connectors.BotLinkConfig{
		TelegramBotUsername: cfg.TelegramBotUsername,
		BaleBotUsername:     cfg.BaleBotUsername,
	}, transactor, auditService, log)
	pollCtx, pollCancel := context.WithCancel(context.Background())
	defer pollCancel()
	if telegramBot != nil {
		go connectors.NewPoller(telegramBot, connectorService, domain.ConnectorTelegram, log).Run(pollCtx)
	}
	if baleBot != nil {
		go connectors.NewPoller(baleBot, connectorService, domain.ConnectorBale, log).Run(pollCtx)
	}

	storageClient, err := storage.NewClient(cfg, log)
	if err != nil {
		log.Error("failed to initialize storage client", "error", err)
		os.Exit(1)
	}
	mediaRepo := media.NewRepository(db)
	// usage reader is nil — the worker never serves the files "by owner" view.
	mediaService := media.NewService(mediaRepo, storageClient, nil, nil, transactor, auditService, log)
	queueServer.HandleFunc(domain.TypeMediaCleanup, media.NewCleanupHandler(mediaService))
	queueServer.HandleFunc(domain.TypeOrganizationCleanup, organizations.NewCleanupHandler(storageClient))

	// Anti-cheat: render question/option text to distorted images (server-side).
	questionRepo := questionbanks.NewQuestionRepository(db)
	questionImageRenderer := questionbanks.NewImageRenderer(questionRepo, mediaRepo, storageClient, log)
	queueServer.HandleFunc(domain.TypeQuestionRenderImages, questionbanks.NewRenderImagesHandler(questionImageRenderer))

	// --- AI quiz grading (worker path): one Asynq task grades one submission's
	// descriptive answers on the rate-limited `ai` queue. Disabled (no-op) when
	// LLM_API_KEY is empty — llmClient stays a nil interface. ---
	quizRepo := quizzes.NewRepository(db)
	quizRuleRepo := quizzes.NewRuleRepository(db)
	quizRoomRepo := quizzes.NewRoomRepository(db)
	quizSubmissionRepo := quizzes.NewSubmissionRepository(db)
	aiUsageRepo := ai.NewUsageRepository(db)
	aiJobRepo := ai.NewJobRepository(db)
	var llmClient domain.LLM
	if built, err := llm.New(llm.AdapterConfig{
		APIKey:    cfg.LLMAPIKey,
		Model:     cfg.LLMModel,
		BaseURL:   cfg.LLMBaseURL,
		MaxTokens: cfg.LLMMaxTokens,
		Timeout:   cfg.LLMTimeout,
	}, cfg.LLMProvider, aiUsageRepo, cfg.LLMAIQueueConcurrency); err != nil {
		log.Error("llm init failed", "error", err)
		os.Exit(1)
	} else if built != nil {
		llmClient = built
	}
	aiGradingWorker := quizzes.NewAIGradingWorker(
		quizRepo, quizRuleRepo, quizRoomRepo, quizSubmissionRepo, questionRepo,
		classRepo, classMemberRepo, queueClient, llmClient, aiJobRepo, transactor, auditService, log,
	)
	queueServer.HandleFunc(domain.TypeQuizAIGradeSubmission, quizzes.NewAIGradeSubmissionHandler(aiGradingWorker))

	// Share-code redeems: clone a bank (questions + media) into the redeemer's org.
	questionBankRepo := questionbanks.NewRepository(db)
	bankCopier := questionbanks.NewBankCopier(questionBankRepo, questionRepo, mediaRepo, mediaService, storageClient, log)
	queueServer.HandleFunc(domain.TypeQuestionBankCopy, questionbanks.NewCopyBankHandler(bankCopier))

	// --- bulk imports: service isn't used to enqueue here (only the API does),
	// but the constructor requires a queue client + result store regardless. ---
	roleRepo := roles.NewRoleRepository(db)
	entitlementRepo := entitlements.NewRepository(db)
	entitlementService := entitlements.NewService(entitlementRepo)
	importRepo := imports.NewRepository(db)
	importService := imports.NewService(
		importRepo, userRepo, roleRepo, classRepo, classMemberRepo, mediaRepo,
		entitlementService, storageClient, queueClient,
		imports.NewRedisResultStore(redisClient), transactor, auditService, log,
	)
	queueServer.HandleFunc(domain.TypeImportProcess, imports.NewProcessHandler(importService))

	retentionSweeper := livesessions.NewRetentionSweeper(livesessions.NewRetentionRepository(db), storageClient, log)
	queueServer.HandleFunc(domain.TypeRecordingRetentionSweep, livesessions.NewRetentionSweepHandler(retentionSweeper))

	// --- billing (worker: PDF generation + reminder/expire sweeps only; no
	// gateway calls, so the payment registry is empty). orgRepo is built above
	// for the connector link greeting. ---
	billingRepo := billing.NewRepository(db)
	billingIssuer := billing.IssuerConfig{
		Name:       cfg.InvoiceIssuerName,
		EconomicID: cfg.InvoiceIssuerEconomicID,
		Address:    cfg.InvoiceIssuerAddress,
		Phone:      cfg.InvoiceIssuerPhone,
	}
	billingPDF := billing.NewPDFRenderer(storageClient, orgRepo, billingIssuer, cfg.ChromeRemoteURL)
	billingSvc := billing.NewService(
		billingRepo,
		orgRepo,
		orgRepo,
		billing.NewEntitlementsCacheBuster(redisClient),
		payment.NewRegistry(),
		storageClient,
		billing.NewQueueEnqueuer(queueClient),
		notificationService,
		billingPDF,
		auditService,
		billing.BillingConfig{
			AppURLTemplate: cfg.AppURLTemplate,
			Issuer:         billingIssuer,
		},
		log,
	)
	queueServer.HandleFunc(domain.TypeInvoiceGeneratePDF, billing.NewGeneratePDFHandler(billingSvc))
	queueServer.HandleFunc(domain.TypeBillingReminderSweep, billing.NewReminderSweepHandler(billingSvc))
	queueServer.HandleFunc(domain.TypeBillingExpireSweep, billing.NewExpireSweepHandler(billingSvc))

	// Periodic queue self-inspection: warns when tasks pile up in the archived
	// (dead-letter) / retry sets so exhausted-retry failures don't sit unnoticed.
	queueServer.HandleFunc(domain.TypeQueueHealthCheck, queue.NewHealthCheckHandler(queueClient, log))

	// Periodic safety net for missed LiveKit webhooks: re-scan for active rooms
	// whose host went stale and close the ones LiveKit confirms are host-less.
	// The event-driven webhook path is primary; this catches dropped events.
	redisOpt, err := asynq.ParseRedisURI(cfg.QueueRedisURL())
	if err != nil {
		log.Error("failed to parse redis URI for scheduler", "error", err)
		os.Exit(1)
	}
	scheduler := asynq.NewScheduler(redisOpt, nil)
	if _, err := scheduler.Register("@every 5m", asynq.NewTask(domain.TypeLiveSessionAutoClose, nil)); err != nil {
		log.Error("failed to register auto-close schedule", "error", err)
		os.Exit(1)
	}
	if _, err := scheduler.Register("@every 24h", asynq.NewTask(domain.TypeRecordingRetentionSweep, nil), asynq.Queue(domain.QueueMedia)); err != nil {
		log.Error("failed to register recording-retention schedule", "error", err)
		os.Exit(1)
	}
	if _, err := scheduler.Register("@every 24h", asynq.NewTask(domain.TypeBillingReminderSweep, nil)); err != nil {
		log.Error("failed to register billing reminder schedule", "error", err)
		os.Exit(1)
	}
	if _, err := scheduler.Register("@every 1h", asynq.NewTask(domain.TypeBillingExpireSweep, nil)); err != nil {
		log.Error("failed to register billing expire schedule", "error", err)
		os.Exit(1)
	}
	if _, err := scheduler.Register("@every 15m", asynq.NewTask(domain.TypeQueueHealthCheck, nil)); err != nil {
		log.Error("failed to register queue health schedule", "error", err)
		os.Exit(1)
	}
	go func() {
		log.Info("starting asynq scheduler")
		if err := scheduler.Run(); err != nil {
			log.Error("scheduler error", "error", err)
			os.Exit(1)
		}
	}()

	go func() {
		log.Info("starting worker server")
		if err := queueServer.Run(); err != nil {
			log.Error("worker server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down worker...")

	pollCancel()
	scheduler.Shutdown()
	queueServer.Shutdown()

	sqlDB, _ := db.DB()
	sqlDB.Close()
	redisClient.Close()

	log.Info("worker stopped")
}

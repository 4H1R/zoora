package quizzes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/queue"
)

type service struct {
	repo        domain.QuizRepository
	rules       domain.QuizRuleRepository
	rooms       domain.QuizRoomRepository
	submissions domain.QuizSubmissionRepository
	questions   domain.QuestionRepository
	classes     domain.ClassRepository
	members     domain.ClassMemberRepository
	queue       *queue.Client
	llm         domain.LLM // may be nil when AI is disabled
	aiJobs      domain.AIGradingJobRepository
	logger      *slog.Logger
}

func NewService(
	repo domain.QuizRepository,
	rules domain.QuizRuleRepository,
	rooms domain.QuizRoomRepository,
	submissions domain.QuizSubmissionRepository,
	questions domain.QuestionRepository,
	classes domain.ClassRepository,
	members domain.ClassMemberRepository,
	queueClient *queue.Client,
	llmClient domain.LLM,
	aiJobs domain.AIGradingJobRepository,
	logger *slog.Logger,
) domain.QuizService {
	return newService(repo, rules, rooms, submissions, questions, classes, members, queueClient, llmClient, aiJobs, logger)
}

// NewAIGradingWorker builds the concrete service for the worker's AI-grade task
// handler, which needs the concrete *service (NewService returns the interface).
func NewAIGradingWorker(
	repo domain.QuizRepository,
	rules domain.QuizRuleRepository,
	rooms domain.QuizRoomRepository,
	submissions domain.QuizSubmissionRepository,
	questions domain.QuestionRepository,
	classes domain.ClassRepository,
	members domain.ClassMemberRepository,
	queueClient *queue.Client,
	llmClient domain.LLM,
	aiJobs domain.AIGradingJobRepository,
	logger *slog.Logger,
) *service {
	return newService(repo, rules, rooms, submissions, questions, classes, members, queueClient, llmClient, aiJobs, logger)
}

func newService(
	repo domain.QuizRepository,
	rules domain.QuizRuleRepository,
	rooms domain.QuizRoomRepository,
	submissions domain.QuizSubmissionRepository,
	questions domain.QuestionRepository,
	classes domain.ClassRepository,
	members domain.ClassMemberRepository,
	queueClient *queue.Client,
	llmClient domain.LLM,
	aiJobs domain.AIGradingJobRepository,
	logger *slog.Logger,
) *service {
	return &service{
		repo:        repo,
		rules:       rules,
		rooms:       rooms,
		submissions: submissions,
		questions:   questions,
		classes:     classes,
		members:     members,
		queue:       queueClient,
		llm:         llmClient,
		aiJobs:      aiJobs,
		logger:      logger,
	}
}

// enqueueQuestionRender schedules anti-cheat image rendering for a single
// question, mirroring questionbanks' enqueue: media queue, question-scoped
// TaskID so repeat enqueues coalesce onto one pending task. Best-effort — a
// failure is logged, not surfaced; the take gate catches not-ready questions.
func (s *service) enqueueQuestionRender(ctx context.Context, questionID uuid.UUID) {
	if s.queue == nil {
		return
	}
	payload, err := json.Marshal(domain.QuestionRenderImagesPayload{QuestionID: questionID})
	if err != nil {
		s.logger.Error("marshal render-images payload", "question_id", questionID.String(), "error", err)
		return
	}
	task := asynq.NewTask(domain.TypeQuestionRenderImages, payload)
	_, err = s.queue.Enqueue(task,
		asynq.Queue(domain.QueueMedia),
		asynq.TaskID("question-render-"+questionID.String()),
	)
	if err != nil && !errors.Is(err, asynq.ErrTaskIDConflict) {
		s.logger.Error("enqueue render-images", "question_id", questionID.String(), "error", err)
	}
}

// enqueueQuizImageRenders resolves every question the quiz can draw (manual
// rules' explicit questions plus every question in each random rule's bank) and
// enqueues a render for any that has not been rendered yet (status 'none'). It
// marks those questions pending first so the take gate blocks until they are
// ready. Called after a render-as-image quiz's rules or flag change. No-op when
// the quiz does not render as image.
func (s *service) enqueueQuizImageRenders(ctx context.Context, quiz *domain.Quiz) {
	if !quiz.RenderAsImage {
		return
	}
	ids, err := s.candidateQuestionIDs(ctx, quiz.ID)
	if err != nil {
		s.logger.Error("resolve quiz image candidates", "quiz_id", quiz.ID.String(), "error", err)
		return
	}
	for _, id := range ids {
		q, err := s.questions.FindByID(ctx, id)
		if err != nil {
			s.logger.Error("load candidate question", "question_id", id.String(), "error", err)
			continue
		}
		if q.ImageRenderStatus != domain.ImageRenderStatusNone {
			continue // already pending/ready/failed — leave its cache alone
		}
		q.ImageRenderStatus = domain.ImageRenderStatusPending
		if err := s.questions.Update(ctx, q); err != nil {
			s.logger.Error("mark question pending", "question_id", id.String(), "error", err)
			continue
		}
		s.enqueueQuestionRender(ctx, id)
	}
}

// candidateQuestionIDs returns the de-duplicated set of question ids a quiz can
// draw: manual rules contribute their explicit ids; random rules contribute
// every question in the referenced bank (any of them may be drawn per student).
func (s *service) candidateQuestionIDs(ctx context.Context, quizID uuid.UUID) ([]uuid.UUID, error) {
	rules, _, err := s.rules.ListByQuiz(ctx, quizID, domain.ListParams{Page: 1, PageSize: 10000})
	if err != nil {
		return nil, err
	}
	seen := make(map[uuid.UUID]struct{})
	out := make([]uuid.UUID, 0)
	add := func(id uuid.UUID) {
		if _, dup := seen[id]; dup {
			return
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	for _, r := range rules {
		switch r.Type {
		case domain.QuizRuleTypeManual:
			for _, id := range r.QuestionIDs {
				add(id)
			}
		case domain.QuizRuleTypeRandom:
			if r.BankID == nil {
				continue
			}
			qs, err := s.questions.ListAllByBank(ctx, *r.BankID)
			if err != nil {
				return nil, err
			}
			for i := range qs {
				add(qs[i].ID)
			}
		}
	}
	return out, nil
}

// quizImageRenderStatus aggregates the render status of a quiz's candidate
// questions into one quiz-level status for the authoring UI and take gate:
// failed if any failed, pending if any not-yet-ready, else ready. Returns
// 'none' when the quiz does not render as image or draws no questions.
func (s *service) quizImageRenderStatus(ctx context.Context, quiz *domain.Quiz) domain.ImageRenderStatus {
	if !quiz.RenderAsImage {
		return domain.ImageRenderStatusNone
	}
	ids, err := s.candidateQuestionIDs(ctx, quiz.ID)
	if err != nil || len(ids) == 0 {
		return domain.ImageRenderStatusNone
	}
	status := domain.ImageRenderStatusReady
	for _, id := range ids {
		q, err := s.questions.FindByID(ctx, id)
		if err != nil {
			return domain.ImageRenderStatusPending
		}
		switch q.ImageRenderStatus {
		case domain.ImageRenderStatusFailed:
			return domain.ImageRenderStatusFailed
		case domain.ImageRenderStatusReady:
			// keep scanning
		default: // none or pending
			status = domain.ImageRenderStatusPending
		}
	}
	return status
}

func canManageQuiz(caller domain.Caller, quiz *domain.Quiz) bool {
	return caller.CanManage(quiz.UserID, domain.PermQuizzesUpdateAny)
}

func canDeleteQuiz(caller domain.Caller, quiz *domain.Quiz) bool {
	return caller.CanManage(quiz.UserID, domain.PermQuizzesDeleteAny)
}

func (s *service) canViewQuiz(ctx context.Context, caller domain.Caller, quiz *domain.Quiz) (bool, error) {
	if canManageQuiz(caller, quiz) {
		return true, nil
	}
	if caller.HasPermission(domain.PermQuizzesViewAny) {
		return true, nil
	}
	ok, err := s.members.Exists(ctx, quiz.ClassID, caller.UserID)
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (s *service) Create(ctx context.Context, dto domain.CreateQuizDTO) (*domain.Quiz, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	class, err := s.classes.FindByID(ctx, dto.ClassID)
	if err != nil {
		return nil, err
	}
	if !canManageClass(caller, class) {
		return nil, domain.ErrForbidden
	}
	if (dto.TrackTabSwitches || dto.RequireGPS) && !caller.HasFeature(domain.FeatureAdvancedAntiCheat) {
		return nil, domain.NewFeatureError(caller.Ent.Plan, domain.FeatureAdvancedAntiCheat)
	}
	mode, val, wpp := domain.NormalizeNegativeMark(dto.NegativeMarkMode, dto.NegativeValue, dto.WrongsPerPoint)
	if err := domain.ValidateNegativeMark(mode, val, wpp); err != nil {
		return nil, err
	}
	quiz := &domain.Quiz{
		OrganizationID:             class.OrganizationID,
		UserID:                     caller.UserID,
		ClassID:                    dto.ClassID,
		Title:                      dto.Title,
		Description:                dto.Description,
		DurationMinutes:            dto.DurationMinutes,
		NoBackNavigation:           dto.NoBackNavigation,
		ShuffleQuestions:           dto.ShuffleQuestions,
		ShuffleOptions:             dto.ShuffleOptions,
		TrackTabSwitches:           dto.TrackTabSwitches,
		RequireGPS:                 dto.RequireGPS,
		DisableCopyPaste:           dto.DisableCopyPaste,
		DisableRightClickShortcuts: dto.DisableRightClickShortcuts,
		ShowResults:                dto.ShowResults,
		RenderAsImage:              dto.RenderAsImage,
		NegativeMarkMode:           mode,
		NegativeValue:              val,
		WrongsPerPoint:             wpp,
	}
	if err := s.repo.Create(ctx, quiz); err != nil {
		return nil, err
	}
	s.logger.Info("quiz created",
		"quiz_id", quiz.ID.String(),
		"class_id", quiz.ClassID.String(),
		"created_by", caller.UserID.String(),
	)
	// A quiz has no rules at creation, so there is usually nothing to render yet;
	// harmless no-op unless rules were somehow already present.
	s.enqueueQuizImageRenders(ctx, quiz)
	return quiz, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Quiz, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	quiz, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	visible, err := s.canViewQuiz(ctx, caller, quiz)
	if err != nil {
		return nil, err
	}
	if !visible {
		return nil, domain.ErrForbidden
	}
	quiz.ImageRenderStatus = s.quizImageRenderStatus(ctx, quiz)
	return quiz, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, dto domain.UpdateQuizDTO) (*domain.Quiz, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	quiz, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canManageQuiz(caller, quiz) {
		return nil, domain.ErrForbidden
	}
	wantsTab := dto.TrackTabSwitches != nil && *dto.TrackTabSwitches
	wantsGPS := dto.RequireGPS != nil && *dto.RequireGPS
	if (wantsTab || wantsGPS) && !caller.HasFeature(domain.FeatureAdvancedAntiCheat) {
		return nil, domain.NewFeatureError(caller.Ent.Plan, domain.FeatureAdvancedAntiCheat)
	}
	if dto.Title != nil {
		quiz.Title = *dto.Title
	}
	if dto.Description != nil {
		quiz.Description = *dto.Description
	}
	if dto.DurationMinutes != nil {
		quiz.DurationMinutes = *dto.DurationMinutes
	}
	if dto.NoBackNavigation != nil {
		quiz.NoBackNavigation = *dto.NoBackNavigation
	}
	if dto.ShuffleQuestions != nil {
		quiz.ShuffleQuestions = *dto.ShuffleQuestions
	}
	if dto.ShuffleOptions != nil {
		quiz.ShuffleOptions = *dto.ShuffleOptions
	}
	if dto.TrackTabSwitches != nil {
		quiz.TrackTabSwitches = *dto.TrackTabSwitches
	}
	if dto.RequireGPS != nil {
		quiz.RequireGPS = *dto.RequireGPS
	}
	if dto.DisableCopyPaste != nil {
		quiz.DisableCopyPaste = *dto.DisableCopyPaste
	}
	if dto.DisableRightClickShortcuts != nil {
		quiz.DisableRightClickShortcuts = *dto.DisableRightClickShortcuts
	}
	if dto.ShowResults != nil {
		quiz.ShowResults = *dto.ShowResults
	}
	if dto.RenderAsImage != nil {
		quiz.RenderAsImage = *dto.RenderAsImage
	}
	if dto.NegativeMarkMode != nil {
		quiz.NegativeMarkMode = *dto.NegativeMarkMode
	}
	if dto.NegativeValue != nil {
		quiz.NegativeValue = *dto.NegativeValue
	}
	if dto.WrongsPerPoint != nil {
		quiz.WrongsPerPoint = *dto.WrongsPerPoint
	}
	mode, val, wpp := domain.NormalizeNegativeMark(quiz.NegativeMarkMode, quiz.NegativeValue, quiz.WrongsPerPoint)
	if err := domain.ValidateNegativeMark(mode, val, wpp); err != nil {
		return nil, err
	}
	quiz.NegativeMarkMode, quiz.NegativeValue, quiz.WrongsPerPoint = mode, val, wpp
	if err := s.repo.Update(ctx, quiz); err != nil {
		return nil, err
	}
	// If the quiz renders as image (newly turned on, or already on), make sure
	// every candidate question is queued for rendering.
	s.enqueueQuizImageRenders(ctx, quiz)
	return quiz, nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	quiz, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if !canDeleteQuiz(caller, quiz) {
		return domain.ErrForbidden
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.logger.Info("quiz deleted",
		"quiz_id", id.String(),
		"deleted_by", caller.UserID.String(),
	)
	return nil
}

func (s *service) List(ctx context.Context, q domain.ListQuizzesQuery) ([]domain.Quiz, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	scope := s.resolveListScope(caller)
	scope.ClassID = q.ClassID
	scope.ClassSessionID = q.ClassSessionID
	if canManage(caller) {
		scope.IncludeDeleted = q.IncludeDeleted
	}
	return s.repo.List(ctx, scope, q.ListParams)
}

func (s *service) ListMine(ctx context.Context, p domain.ListParams) ([]domain.MyExam, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}

	quizzes, total, err := s.repo.ListByMemberWithRooms(ctx, caller.UserID, p)
	if err != nil {
		return nil, 0, fmt.Errorf("listing my exams: %w", err)
	}

	now := time.Now()
	exams := make([]domain.MyExam, 0, len(quizzes))
	for i := range quizzes {
		q := quizzes[i]

		ex := domain.MyExam{
			QuizID:          q.ID,
			Title:           q.Title,
			ClassID:         q.ClassID,
			DurationMinutes: q.DurationMinutes,
			TotalScore:      q.TotalScore,
		}
		if q.Class != nil {
			ex.ClassName = q.Class.Name
		}

		// Pick the room to surface: prefer an open room, else the next upcoming.
		rooms, _, err := s.rooms.ListByQuiz(ctx, q.ID, domain.ListParams{Page: 1, PageSize: 10000})
		if err != nil {
			return nil, 0, fmt.Errorf("listing rooms for exam %s: %w", q.ID, err)
		}
		var open *domain.QuizRoom
		var nextUpcoming *domain.QuizRoom
		for j := range rooms {
			rm := rooms[j]
			if rm.IsRoomOpenAt(now) {
				r := rm
				open = &r
				break
			}
			if rm.StartedAt != nil && rm.StartedAt.After(now) {
				if nextUpcoming == nil || rm.StartedAt.Before(*nextUpcoming.StartedAt) {
					r := rm
					nextUpcoming = &r
				}
			}
		}
		chosen := open
		if chosen == nil {
			chosen = nextUpcoming
		}
		if chosen != nil {
			ex.Room = &domain.MyExamRoom{
				ID:             chosen.ID,
				ClassSessionID: chosen.ClassSessionID,
				StartedAt:      chosen.StartedAt,
				EndedAt:        chosen.EndedAt,
				IsOpen:         chosen.IsRoomOpenAt(now),
			}
		}

		// Caller's own submission decides submitted/graded.
		sub, err := s.submissions.FindByQuizAndUser(ctx, q.ID, caller.UserID)
		switch {
		case err == nil && sub != nil:
			switch sub.Status {
			case domain.SubmissionStatusGraded:
				ex.State = domain.MyExamStateGraded
				if s.resultsRevealed(ctx, &q, sub, now) {
					score := sub.TotalScore
					ex.Score = &score
				}
			default:
				ex.State = domain.MyExamStateSubmitted
			}
			ex.SubmittedAt = sub.SubmittedAt
		case errors.Is(err, domain.ErrNotFound):
			if ex.Room != nil && ex.Room.IsOpen {
				ex.State = domain.MyExamStateOpen
			} else {
				ex.State = domain.MyExamStateUpcoming
			}
		default:
			return nil, 0, fmt.Errorf("loading submission for exam %s: %w", q.ID, err)
		}

		exams = append(exams, ex)
	}

	return exams, total, nil
}

func (s *service) resolveListScope(caller domain.Caller) domain.QuizListScope {
	if caller.IsAdmin {
		return domain.QuizListScope{All: true}
	}
	if caller.HasPermission(domain.PermQuizzesViewAny) || caller.HasPermission(domain.PermQuizzesUpdateAny) {
		return domain.QuizListScope{All: true, OrganizationID: caller.OrgID}
	}
	userID := caller.UserID
	return domain.QuizListScope{
		OwnerID:      &userID,
		MemberUserID: &userID,
	}
}

func canManage(caller domain.Caller) bool {
	return caller.HasAny(domain.PermQuizzesUpdateAny)
}

func canManageClass(caller domain.Caller, class *domain.Class) bool {
	return caller.CanManage(class.UserID, domain.PermClassesUpdateAny)
}

func (s *service) CreateRule(ctx context.Context, quizID uuid.UUID, dto domain.CreateQuizRuleDTO) (*domain.QuizRule, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	quiz, err := s.repo.FindByID(ctx, quizID)
	if err != nil {
		return nil, err
	}
	if !canManageQuiz(caller, quiz) {
		return nil, domain.ErrForbidden
	}
	questionIDs := dto.QuestionIDs
	if questionIDs == nil {
		questionIDs = []uuid.UUID{}
	}
	overrides, err := normalizeNegativeOverrides(dto.NegativeOverrides)
	if err != nil {
		return nil, err
	}
	if err := validateRuleNegativeDefault(dto.NegativeDefaultMode, dto.NegativeDefaultValue, dto.NegativeDefaultWrongsPerPoint); err != nil {
		return nil, err
	}
	ndValue, ndWrongs := normalizeRuleNegativeDefault(dto.NegativeDefaultMode, dto.NegativeDefaultValue, dto.NegativeDefaultWrongsPerPoint)
	rule := &domain.QuizRule{
		QuizID:                        quizID,
		Type:                          dto.Type,
		BankID:                        dto.BankID,
		QuestionIDs:                   questionIDs,
		Count:                         dto.Count,
		IsDynamic:                     dto.IsDynamic,
		NegativeOverrides:             overrides,
		NegativeDefaultMode:           dto.NegativeDefaultMode,
		NegativeDefaultValue:          ndValue,
		NegativeDefaultWrongsPerPoint: ndWrongs,
	}
	if err := s.rules.Create(ctx, rule); err != nil {
		return nil, err
	}
	if err := s.recomputeQuizTotal(ctx, quizID); err != nil {
		s.logger.Warn("failed to recompute quiz total", "quiz_id", quizID.String(), "err", err)
	}
	// New questions entered the candidate set — render them if the quiz is image.
	s.enqueueQuizImageRenders(ctx, quiz)
	return rule, nil
}

func (s *service) GetRule(ctx context.Context, id uuid.UUID) (*domain.QuizRule, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	rule, err := s.rules.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	quiz, err := s.repo.FindByID(ctx, rule.QuizID)
	if err != nil {
		return nil, err
	}
	visible, err := s.canViewQuiz(ctx, caller, quiz)
	if err != nil {
		return nil, err
	}
	if !visible {
		return nil, domain.ErrForbidden
	}
	return rule, nil
}

func (s *service) UpdateRule(ctx context.Context, id uuid.UUID, dto domain.UpdateQuizRuleDTO) (*domain.QuizRule, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	rule, err := s.rules.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	quiz, err := s.repo.FindByID(ctx, rule.QuizID)
	if err != nil {
		return nil, err
	}
	if !canManageQuiz(caller, quiz) {
		return nil, domain.ErrForbidden
	}
	if dto.Type != nil {
		rule.Type = *dto.Type
	}
	if dto.BankID != nil {
		rule.BankID = dto.BankID
	}
	if dto.QuestionIDs != nil {
		rule.QuestionIDs = dto.QuestionIDs
	}
	if dto.Count != nil {
		rule.Count = *dto.Count
	}
	if dto.IsDynamic != nil {
		rule.IsDynamic = *dto.IsDynamic
	}
	if dto.NegativeOverrides != nil {
		overrides, err := normalizeNegativeOverrides(dto.NegativeOverrides)
		if err != nil {
			return nil, err
		}
		rule.NegativeOverrides = overrides
	}
	if dto.NegativeDefaultMode != nil {
		if err := validateRuleNegativeDefault(dto.NegativeDefaultMode, dto.NegativeDefaultValue, dto.NegativeDefaultWrongsPerPoint); err != nil {
			return nil, err
		}
		rule.NegativeDefaultMode = dto.NegativeDefaultMode
		rule.NegativeDefaultValue, rule.NegativeDefaultWrongsPerPoint = normalizeRuleNegativeDefault(dto.NegativeDefaultMode, dto.NegativeDefaultValue, dto.NegativeDefaultWrongsPerPoint)
	}
	if err := s.rules.Update(ctx, rule); err != nil {
		return nil, err
	}
	if err := s.recomputeQuizTotal(ctx, rule.QuizID); err != nil {
		s.logger.Warn("failed to recompute quiz total", "quiz_id", rule.QuizID.String(), "err", err)
	}
	// Rule bank/questions may have changed the candidate set — render if image.
	s.enqueueQuizImageRenders(ctx, quiz)
	return rule, nil
}

// validateRuleNegativeDefault accepts nil (keep question default) or one of the
// valid modes (none/per_wrong/accumulative) for a rule-wide default. The numeric
// fields are optional (nil derives from option count), but when supplied they
// must be in range for their mode: per_wrong value > 0, accumulative wrongs 2-5.
func validateRuleNegativeDefault(mode *domain.NegativeMarkMode, value *float64, wrongsPerPoint *int) error {
	if mode == nil {
		return nil
	}
	if !mode.Valid() {
		return domain.NewValidationError(map[string]string{"negative_default_mode": "invalid mode"})
	}
	switch *mode {
	case domain.NegativeMarkPerWrong:
		if value != nil && *value <= 0 {
			return domain.NewValidationError(map[string]string{"negative_default_value": "must be greater than 0"})
		}
	case domain.NegativeMarkAccumulative:
		if wrongsPerPoint != nil && (*wrongsPerPoint < 2 || *wrongsPerPoint > 5) {
			return domain.NewValidationError(map[string]string{"negative_default_wrongs_per_point": "must be between 2 and 5"})
		}
	}
	return nil
}

// normalizeRuleNegativeDefault keeps only the numeric field relevant to the mode
// so stored rows stay clean: per_wrong keeps value, accumulative keeps wrongs,
// none/nil clears both.
func normalizeRuleNegativeDefault(mode *domain.NegativeMarkMode, value *float64, wrongsPerPoint *int) (*float64, *int) {
	if mode == nil {
		return nil, nil
	}
	switch *mode {
	case domain.NegativeMarkPerWrong:
		return value, nil
	case domain.NegativeMarkAccumulative:
		return nil, wrongsPerPoint
	default:
		return nil, nil
	}
}

// normalizeNegativeOverrides normalizes and validates each per-question
// negative-marking override on a quiz rule.
func normalizeNegativeOverrides(in []domain.QuizQuestionNegativeOverride) ([]domain.QuizQuestionNegativeOverride, error) {
	out := make([]domain.QuizQuestionNegativeOverride, 0, len(in))
	for _, o := range in {
		m, v, w := domain.NormalizeNegativeMark(o.Mode, o.NegativeValue, o.WrongsPerPoint)
		if err := domain.ValidateNegativeMark(m, v, w); err != nil {
			return nil, err
		}
		out = append(out, domain.QuizQuestionNegativeOverride{
			QuestionID:     o.QuestionID,
			Mode:           m,
			NegativeValue:  v,
			WrongsPerPoint: w,
		})
	}
	return out, nil
}

func (s *service) DeleteRule(ctx context.Context, id uuid.UUID) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	rule, err := s.rules.FindByID(ctx, id)
	if err != nil {
		return err
	}
	quiz, err := s.repo.FindByID(ctx, rule.QuizID)
	if err != nil {
		return err
	}
	if !canManageQuiz(caller, quiz) {
		return domain.ErrForbidden
	}
	if err := s.rules.Delete(ctx, id); err != nil {
		return err
	}
	if err := s.recomputeQuizTotal(ctx, rule.QuizID); err != nil {
		s.logger.Warn("failed to recompute quiz total", "quiz_id", rule.QuizID.String(), "err", err)
	}
	return nil
}

// recomputeQuizTotal aggregates the max score of every question wired into
// the quiz's rules and persists it on the quiz row.
func (s *service) recomputeQuizTotal(ctx context.Context, quizID uuid.UUID) error {
	rules, _, err := s.rules.ListByQuiz(ctx, quizID, domain.ListParams{Page: 1, PageSize: 10000})
	if err != nil {
		return err
	}
	var total float64
	for _, r := range rules {
		switch r.Type {
		case domain.QuizRuleTypeManual:
			if len(r.QuestionIDs) == 0 {
				continue
			}
			qs, err := s.questions.FindByIDs(ctx, r.QuestionIDs)
			if err != nil {
				return err
			}
			for i := range qs {
				total += qs[i].MaxScore()
			}
		case domain.QuizRuleTypeRandom:
			if r.BankID == nil || r.Count == 0 {
				continue
			}
			all, err := s.questions.ListAllByBank(ctx, *r.BankID)
			if err != nil {
				return err
			}
			if len(all) == 0 {
				continue
			}
			var sum float64
			for i := range all {
				sum += all[i].MaxScore()
			}
			total += (sum / float64(len(all))) * float64(r.Count)
		}
	}
	// Round to 2 decimals — random rules use weighted average that
	// otherwise produces noisy repeating decimals like 41.857142857142854.
	total = math.Round(total*100) / 100
	quiz, err := s.repo.FindByID(ctx, quizID)
	if err != nil {
		return err
	}
	if quiz.TotalScore == total {
		return nil
	}
	quiz.TotalScore = total
	return s.repo.Update(ctx, quiz)
}

func (s *service) ListRules(ctx context.Context, quizID uuid.UUID, q domain.ListQuizRulesQuery) ([]domain.QuizRule, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	quiz, err := s.repo.FindByID(ctx, quizID)
	if err != nil {
		return nil, 0, err
	}
	visible, err := s.canViewQuiz(ctx, caller, quiz)
	if err != nil {
		return nil, 0, err
	}
	if !visible {
		return nil, 0, domain.ErrForbidden
	}
	return s.rules.ListByQuiz(ctx, quizID, q.ListParams)
}

func (s *service) CreateRoom(ctx context.Context, quizID uuid.UUID, dto domain.CreateQuizRoomDTO) (*domain.QuizRoom, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	if err := dto.Validate(); err != nil {
		return nil, err
	}
	quiz, err := s.repo.FindByID(ctx, quizID)
	if err != nil {
		return nil, err
	}
	if !canManageQuiz(caller, quiz) {
		return nil, domain.ErrForbidden
	}
	room := &domain.QuizRoom{
		QuizID:         quizID,
		ClassSessionID: dto.ClassSessionID,
		StartedAt:      dto.StartedAt,
		EndedAt:        dto.EndedAt,
	}
	if err := s.rooms.Create(ctx, room); err != nil {
		return nil, err
	}
	s.logger.Info("quiz room created",
		"room_id", room.ID.String(),
		"quiz_id", quizID.String(),
		"session_id", dto.ClassSessionID.String(),
		"created_by", caller.UserID.String(),
	)
	return room, nil
}

func (s *service) GetRoom(ctx context.Context, id uuid.UUID) (*domain.QuizRoom, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	room, err := s.rooms.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	quiz, err := s.repo.FindByID(ctx, room.QuizID)
	if err != nil {
		return nil, err
	}
	visible, err := s.canViewQuiz(ctx, caller, quiz)
	if err != nil {
		return nil, err
	}
	if !visible {
		return nil, domain.ErrForbidden
	}
	return room, nil
}

func (s *service) StartRoom(ctx context.Context, id uuid.UUID) (*domain.QuizRoom, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	room, err := s.rooms.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	quiz, err := s.repo.FindByID(ctx, room.QuizID)
	if err != nil {
		return nil, err
	}
	if !canManageQuiz(caller, quiz) {
		return nil, domain.ErrForbidden
	}
	now := time.Now()
	// Allow early open: only override StartedAt when it lies in the future.
	if room.StartedAt == nil || now.Before(*room.StartedAt) {
		room.StartedAt = &now
	}
	if room.EndedAt != nil && !room.EndedAt.After(now) {
		return nil, domain.ErrConflict
	}
	if err := s.rooms.Update(ctx, room); err != nil {
		return nil, err
	}
	s.logger.Info("quiz room started",
		"room_id", id.String(),
		"started_by", caller.UserID.String(),
	)
	return room, nil
}

func (s *service) EndRoom(ctx context.Context, id uuid.UUID) (*domain.QuizRoom, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	room, err := s.rooms.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	quiz, err := s.repo.FindByID(ctx, room.QuizID)
	if err != nil {
		return nil, err
	}
	if !canManageQuiz(caller, quiz) {
		return nil, domain.ErrForbidden
	}
	now := time.Now()
	if room.EndedAt != nil && !room.EndedAt.After(now) {
		return nil, domain.ErrConflict
	}
	room.EndedAt = &now
	if err := s.rooms.Update(ctx, room); err != nil {
		return nil, err
	}
	s.logger.Info("quiz room ended",
		"room_id", id.String(),
		"ended_by", caller.UserID.String(),
	)
	return room, nil
}

func (s *service) ListRooms(ctx context.Context, quizID uuid.UUID, q domain.ListQuizRoomsQuery) ([]domain.QuizRoom, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	quiz, err := s.repo.FindByID(ctx, quizID)
	if err != nil {
		return nil, 0, err
	}
	visible, err := s.canViewQuiz(ctx, caller, quiz)
	if err != nil {
		return nil, 0, err
	}
	if !visible {
		return nil, 0, domain.ErrForbidden
	}
	return s.rooms.ListByQuiz(ctx, quizID, q.ListParams)
}

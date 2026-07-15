package questionbanks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/queue"
)

type service struct {
	repo      domain.QuestionBankRepository
	questions domain.QuestionRepository
	media     domain.MediaRepository
	queue     *queue.Client
	logger    *slog.Logger
}

func NewService(
	repo domain.QuestionBankRepository,
	questions domain.QuestionRepository,
	media domain.MediaRepository,
	queueClient *queue.Client,
	logger *slog.Logger,
) domain.QuestionBankService {
	return &service{repo: repo, questions: questions, media: media, queue: queueClient, logger: logger}
}

// enqueueRenderImages schedules anti-cheat image (re)generation for a question.
// Runs on the media queue (S3-bound, like other slow media work) with a
// question-scoped TaskID so a rapid re-save coalesces onto one pending task
// instead of piling up. Best-effort: a failure to enqueue is logged, not
// surfaced — the images simply stay not-ready, which the take gate catches.
func (s *service) enqueueRenderImages(ctx context.Context, questionID uuid.UUID) {
	if s.queue == nil {
		return
	}
	payload, err := json.Marshal(domain.QuestionRenderImagesPayload{QuestionID: questionID})
	if err != nil {
		s.logger.Error("marshal render-images payload", "question_id", questionID.String(), "error", err)
		return
	}
	task := asynq.NewTask(domain.TypeQuestionRenderImages, payload)
	// A question-scoped TaskID coalesces rapid re-saves: while a render is still
	// pending, a repeat enqueue conflicts and is ignored — the pending task reads
	// the latest question row at processing time, so it renders current state.
	_, err = s.queue.Enqueue(task,
		asynq.Queue(domain.QueueMedia),
		asynq.TaskID("question-render-"+questionID.String()),
	)
	if err != nil && !errors.Is(err, asynq.ErrTaskIDConflict) {
		s.logger.Error("enqueue render-images", "question_id", questionID.String(), "error", err)
	}
}

// enqueueMediaCleanup schedules a purge of ALL media (rows + S3 objects) owned
// by a deleted question — teacher-uploaded body/option photos and worker-rendered
// anti-cheat images alike. Media is keyed by question, so a delete alone leaves
// every object orphaned. An empty collection name matches every collection.
// Best-effort: a failure to enqueue is logged, not surfaced, so the delete still
// succeeds and the orphans can be swept later.
func (s *service) enqueueMediaCleanup(ctx context.Context, questionID uuid.UUID) {
	if s.queue == nil {
		return
	}
	payload, err := json.Marshal(domain.MediaCleanupPayload{
		ModelType: domain.QuestionMediaModelType,
		ModelID:   questionID,
	})
	if err != nil {
		s.logger.Error("marshal media-cleanup payload", "question_id", questionID.String(), "error", err)
		return
	}
	if _, err := s.queue.Enqueue(asynq.NewTask(domain.TypeMediaCleanup, payload), asynq.Queue(domain.QueueMedia)); err != nil {
		s.logger.Error("enqueue media-cleanup", "question_id", questionID.String(), "error", err)
	}
}

// stripIncomingSystemImages nils the server-owned SystemImageMediaID on any
// option the client sent — only the worker may set it.
func stripIncomingSystemImages(in []domain.QuestionOption) []domain.QuestionOption {
	if in == nil {
		return in
	}
	out := make([]domain.QuestionOption, len(in))
	for i, o := range in {
		o.SystemImageMediaID = nil
		out[i] = o
	}
	return out
}

func (s *service) validateMetadataMedia(ctx context.Context, items []domain.QuestionMetadata) error {
	if err := domain.ValidateQuestionMetadata(items); err != nil {
		return err
	}
	for i, item := range items {
		m, err := s.media.FindByID(ctx, item.MediaID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return domain.NewValidationError(map[string]string{
					fmt.Sprintf("metadata[%d].media_id", i): "media not found",
				})
			}
			return err
		}
		if m.ModelType != domain.QuestionMediaModelType {
			return domain.NewValidationError(map[string]string{
				fmt.Sprintf("metadata[%d].media_id", i): "media must belong to a question",
			})
		}
		if item.Type == domain.QuestionMetadataPhoto && !strings.HasPrefix(m.MimeType, "image/") {
			return domain.NewValidationError(map[string]string{
				fmt.Sprintf("metadata[%d].media_id", i): "media is not an image",
			})
		}
	}
	return nil
}

func canManageBank(caller domain.Caller, bank *domain.QuestionBank) bool {
	if caller.IsAdmin {
		return true
	}
	if caller.HasPermission(domain.PermQuestionBanksUpdateAny) {
		if caller.OrgID != nil && bank.OrganizationID != *caller.OrgID {
			return false
		}
		return true
	}
	return false
}

func canDeleteBank(caller domain.Caller, bank *domain.QuestionBank) bool {
	if caller.IsAdmin {
		return true
	}
	if caller.HasPermission(domain.PermQuestionBanksDeleteAny) {
		if caller.OrgID != nil && bank.OrganizationID != *caller.OrgID {
			return false
		}
		return true
	}
	return false
}

func canViewBank(caller domain.Caller, bank *domain.QuestionBank) bool {
	if canManageBank(caller, bank) {
		return true
	}
	if caller.HasPermission(domain.PermQuestionBanksViewAny) {
		if caller.OrgID != nil && bank.OrganizationID != *caller.OrgID {
			return false
		}
		return true
	}
	return caller.OrgID != nil && bank.OrganizationID == *caller.OrgID
}

func (s *service) Create(ctx context.Context, dto domain.CreateQuestionBankDTO) (*domain.QuestionBank, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	if !caller.IsAdmin &&
		!caller.HasPermission(domain.PermQuestionBanksCreate) &&
		!caller.HasPermission(domain.PermQuestionBanksCreateAny) &&
		!caller.HasPermission(domain.PermQuestionBanksUpdateAny) {
		return nil, domain.ErrForbidden
	}
	if caller.OrgID == nil && !caller.IsAdmin {
		return nil, domain.ErrForbidden
	}
	bank := &domain.QuestionBank{
		Name:        dto.Name,
		Description: dto.Description,
	}
	if caller.OrgID != nil {
		bank.OrganizationID = *caller.OrgID
	}
	if err := s.repo.Create(ctx, bank); err != nil {
		return nil, err
	}
	s.logger.Info("question bank created",
		"bank_id", bank.ID.String(),
		"created_by", caller.UserID.String(),
	)
	return bank, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*domain.QuestionBank, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	bank, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canViewBank(caller, bank) {
		return nil, domain.ErrForbidden
	}
	return bank, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, dto domain.UpdateQuestionBankDTO) (*domain.QuestionBank, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	bank, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canManageBank(caller, bank) {
		return nil, domain.ErrForbidden
	}
	if dto.Name != nil {
		bank.Name = *dto.Name
	}
	if dto.Description != nil {
		bank.Description = *dto.Description
	}
	if err := s.repo.Update(ctx, bank); err != nil {
		return nil, err
	}
	return bank, nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	bank, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if !canDeleteBank(caller, bank) {
		return domain.ErrForbidden
	}
	// Snapshot the bank's questions before deleting it so their media can be
	// purged. Media is keyed by question (not bank), so the bank delete leaves
	// every question's uploaded + rendered images orphaned otherwise.
	questions, err := s.questions.ListAllByBank(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	for i := range questions {
		s.enqueueMediaCleanup(ctx, questions[i].ID)
	}
	s.logger.Info("question bank deleted",
		"bank_id", id.String(),
		"deleted_by", caller.UserID.String(),
	)
	return nil
}

func (s *service) List(ctx context.Context, p domain.ListParams) ([]domain.QuestionBank, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	if caller.IsAdmin {
		return s.repo.AdminList(ctx, domain.AdminListQuestionBanksQuery{ListParams: p})
	}
	if caller.OrgID == nil {
		return nil, 0, domain.ErrForbidden
	}
	if caller.HasPermission(domain.PermQuestionBanksViewAny) || caller.HasPermission(domain.PermQuestionBanksUpdateAny) {
		return s.repo.AdminList(ctx, domain.AdminListQuestionBanksQuery{OrganizationID: caller.OrgID, ListParams: p})
	}
	return s.repo.List(ctx, *caller.OrgID, p)
}

func (s *service) CreateQuestion(ctx context.Context, bankID uuid.UUID, dto domain.CreateQuestionDTO) (*domain.Question, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	bank, err := s.repo.FindByID(ctx, bankID)
	if err != nil {
		return nil, err
	}
	if !canManageBank(caller, bank) {
		return nil, domain.ErrForbidden
	}
	options := stripIncomingSystemImages(clearOptionImagesForNonChoice(dto.Type, dto.Options))
	if err := domain.ValidateQuestionOptions(dto.Type, options); err != nil {
		return nil, err
	}
	if err := s.validateMetadataMedia(ctx, dto.Metadata); err != nil {
		return nil, err
	}
	mode, val, wpp := domain.NormalizeNegativeMark(dto.NegativeMarkMode, dto.NegativeValue, dto.WrongsPerPoint)
	if dto.Type != domain.QuestionTypeChoice {
		mode, val, wpp = domain.NegativeMarkNone, 0, 0
	}
	if err := domain.ValidateNegativeMark(mode, val, wpp); err != nil {
		return nil, err
	}
	metadata := dto.Metadata
	if metadata == nil {
		metadata = []domain.QuestionMetadata{}
	}
	question := &domain.Question{
		BankID:            bankID,
		OrganizationID:    bank.OrganizationID,
		Text:              dto.Text,
		Type:              dto.Type,
		Options:           options,
		ModelAnswer:       dto.ModelAnswer,
		Metadata:          metadata,
		NegativeMarkMode:  mode,
		NegativeValue:     val,
		WrongsPerPoint:    wpp,
		MinSeconds:        dto.MinSeconds,
		ImageRenderStatus: domain.ImageRenderStatusNone,
	}
	// Rendering is not decided here — the quiz owns the render_as_image switch.
	// A new question starts as 'none'; when a quiz that renders as image uses it
	// (or is saved), the quiz service enqueues the render. The take gate is the
	// safety net for questions added to a bank after the quiz was saved.
	if err := s.questions.Create(ctx, question); err != nil {
		return nil, err
	}
	return question, nil
}

func (s *service) GetQuestion(ctx context.Context, id uuid.UUID) (*domain.Question, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	question, err := s.questions.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	bank, err := s.repo.FindByID(ctx, question.BankID)
	if err != nil {
		return nil, err
	}
	if !canViewBank(caller, bank) {
		return nil, domain.ErrForbidden
	}
	return question, nil
}

func (s *service) UpdateQuestion(ctx context.Context, id uuid.UUID, dto domain.UpdateQuestionDTO) (*domain.Question, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	question, err := s.questions.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	bank, err := s.repo.FindByID(ctx, question.BankID)
	if err != nil {
		return nil, err
	}
	if !canManageBank(caller, bank) {
		return nil, domain.ErrForbidden
	}
	// A question "participates" in image rendering once it has been rendered for
	// some quiz (status != none). Editing its content then re-renders the cache.
	participates := question.ImageRenderStatus != domain.ImageRenderStatusNone
	if dto.Text != nil {
		question.Text = *dto.Text
	}
	if dto.Type != nil {
		question.Type = *dto.Type
	}
	if dto.Options != nil {
		question.Options = stripIncomingSystemImages(dto.Options)
	}
	if dto.ModelAnswer != nil {
		question.ModelAnswer = *dto.ModelAnswer
	}
	question.Options = clearOptionImagesForNonChoice(question.Type, question.Options)
	if dto.Options != nil || dto.Type != nil {
		if err := domain.ValidateQuestionOptions(question.Type, question.Options); err != nil {
			return nil, err
		}
	}
	if dto.NegativeMarkMode != nil {
		question.NegativeMarkMode = *dto.NegativeMarkMode
	}
	if dto.NegativeValue != nil {
		question.NegativeValue = *dto.NegativeValue
	}
	if dto.WrongsPerPoint != nil {
		question.WrongsPerPoint = *dto.WrongsPerPoint
	}
	if dto.MinSeconds != nil {
		question.MinSeconds = *dto.MinSeconds
	}
	mode, val, wpp := domain.NormalizeNegativeMark(question.NegativeMarkMode, question.NegativeValue, question.WrongsPerPoint)
	if question.Type != domain.QuestionTypeChoice {
		mode, val, wpp = domain.NegativeMarkNone, 0, 0
	}
	if err := domain.ValidateNegativeMark(mode, val, wpp); err != nil {
		return nil, err
	}
	question.NegativeMarkMode, question.NegativeValue, question.WrongsPerPoint = mode, val, wpp
	if dto.Metadata != nil {
		if err := s.validateMetadataMedia(ctx, dto.Metadata); err != nil {
			return nil, err
		}
		question.Metadata = dto.Metadata
	}

	// If this question already participates in rendering and its rendered content
	// (body text, option values, or type) changed, the cached images are stale.
	// Mark not-ready and clear the stale ids so a mid-render take can't be served
	// old images; the worker refills them.
	contentChanged := dto.Text != nil || dto.Options != nil || dto.Type != nil
	enqueueRender := participates && contentChanged
	if enqueueRender {
		question.ImageRenderStatus = domain.ImageRenderStatusPending
		clearSystemImages(question)
	}

	if err := s.questions.Update(ctx, question); err != nil {
		return nil, err
	}
	if enqueueRender {
		s.enqueueRenderImages(ctx, question.ID)
	}
	return question, nil
}

// clearSystemImages nils every server-generated image reference on a question
// (body + options) so a pending/disabled question never carries stale ids.
func clearSystemImages(q *domain.Question) {
	q.SystemImageMediaID = nil
	for i := range q.Options {
		q.Options[i].SystemImageMediaID = nil
	}
}

// clearOptionImagesForNonChoice strips ImageMediaID from options when the
// question type is not choice (images are choice-only).
func clearOptionImagesForNonChoice(t domain.QuestionType, in []domain.QuestionOption) []domain.QuestionOption {
	if t == domain.QuestionTypeChoice || in == nil {
		return in
	}
	out := make([]domain.QuestionOption, len(in))
	for i, o := range in {
		o.ImageMediaID = nil
		out[i] = o
	}
	return out
}

func (s *service) DeleteQuestion(ctx context.Context, id uuid.UUID) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	question, err := s.questions.FindByID(ctx, id)
	if err != nil {
		return err
	}
	bank, err := s.repo.FindByID(ctx, question.BankID)
	if err != nil {
		return err
	}
	if !canDeleteBank(caller, bank) {
		return domain.ErrForbidden
	}
	if err := s.questions.Delete(ctx, id); err != nil {
		return err
	}
	s.enqueueMediaCleanup(ctx, id)
	return nil
}

func (s *service) ListQuestions(ctx context.Context, bankID uuid.UUID, q domain.ListQuestionsQuery) ([]domain.Question, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	bank, err := s.repo.FindByID(ctx, bankID)
	if err != nil {
		return nil, 0, err
	}
	if !canViewBank(caller, bank) {
		return nil, 0, domain.ErrForbidden
	}
	if !canManageBank(caller, bank) {
		q.IncludeDeleted = false
	}
	return s.questions.ListByBank(ctx, bankID, q)
}

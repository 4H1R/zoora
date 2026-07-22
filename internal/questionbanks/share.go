package questionbanks

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/4H1R/zoora/internal/domain"
)

// shareCodeAlphabet excludes visually ambiguous characters (0/O, 1/I). Length
// 32 divides 256 evenly, so byte-mod indexing carries no bias.
const shareCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

const shareCodeLength = 10

func generateShareCode() (string, error) {
	raw := make([]byte, shareCodeLength)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generating share code: %w", err)
	}
	out := make([]byte, shareCodeLength)
	for i, b := range raw {
		out[i] = shareCodeAlphabet[int(b)%len(shareCodeAlphabet)]
	}
	return string(out), nil
}

// normalizeShareCode makes redeem input forgiving: case-insensitive, with
// spaces and dashes stripped.
func normalizeShareCode(code string) string {
	code = strings.ToUpper(strings.TrimSpace(code))
	code = strings.ReplaceAll(code, " ", "")
	return strings.ReplaceAll(code, "-", "")
}

func errInvalidShareCode() error {
	// One generic message for every failure mode (unknown, revoked, expired,
	// deleted bank) so codes can't be probed across tenants.
	return domain.NewValidationError(map[string]string{"code": "invalid or expired code"})
}

func (s *service) GenerateShareCode(ctx context.Context, bankID uuid.UUID, dto domain.GenerateShareCodeDTO) (*domain.QuestionBankShareCode, error) {
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
	if bank.Status != domain.QuestionBankStatusReady {
		return nil, domain.NewValidationError(map[string]string{"bank": "bank is not ready to be shared"})
	}

	now := time.Now()
	var expiresAt *time.Time
	if dto.ExpiresInDays != nil {
		t := now.AddDate(0, 0, *dto.ExpiresInDays)
		expiresAt = &t
	}

	if err := s.repo.RevokeActiveShareCodesByBank(ctx, bankID, now); err != nil {
		return nil, err
	}
	// Retry a few times on the astronomically unlikely code collision.
	for range 3 {
		code, err := generateShareCode()
		if err != nil {
			return nil, err
		}
		sc := &domain.QuestionBankShareCode{
			BankID:         bankID,
			OrganizationID: bank.OrganizationID,
			Code:           code,
			CreatedBy:      caller.UserID,
			ExpiresAt:      expiresAt,
		}
		if err := s.repo.CreateShareCode(ctx, sc); err != nil {
			if errors.Is(err, domain.ErrConflict) {
				continue
			}
			return nil, err
		}
		s.logger.Info("question bank share code generated",
			"bank_id", bankID.String(),
			"created_by", caller.UserID.String(),
		)
		return sc, nil
	}
	return nil, domain.ErrInternal
}

func (s *service) GetShareCode(ctx context.Context, bankID uuid.UUID) (*domain.QuestionBankShareCode, error) {
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
	return s.repo.FindActiveShareCodeByBank(ctx, bankID)
}

func (s *service) RevokeShareCode(ctx context.Context, bankID uuid.UUID) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	bank, err := s.repo.FindByID(ctx, bankID)
	if err != nil {
		return err
	}
	if !canManageBank(caller, bank) {
		return domain.ErrForbidden
	}
	if err := s.repo.RevokeActiveShareCodesByBank(ctx, bankID, time.Now()); err != nil {
		return err
	}
	s.logger.Info("question bank share code revoked",
		"bank_id", bankID.String(),
		"revoked_by", caller.UserID.String(),
	)
	return nil
}

// resolveShareCode loads a code and its source bank, mapping every failure mode
// to the generic invalid-code error.
func (s *service) resolveShareCode(ctx context.Context, raw string) (*domain.QuestionBankShareCode, *domain.QuestionBank, error) {
	sc, err := s.repo.FindShareCodeByCode(ctx, normalizeShareCode(raw))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, nil, errInvalidShareCode()
		}
		return nil, nil, err
	}
	if !sc.Active(time.Now()) {
		return nil, nil, errInvalidShareCode()
	}
	bank, err := s.repo.FindByID(ctx, sc.BankID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, nil, errInvalidShareCode()
		}
		return nil, nil, err
	}
	return sc, bank, nil
}

func (s *service) PreviewShareCode(ctx context.Context, code string) (*domain.ShareCodePreview, error) {
	if _, ok := domain.CallerFromCtx(ctx); !ok {
		return nil, domain.ErrForbidden
	}
	sc, bank, err := s.resolveShareCode(ctx, code)
	if err != nil {
		return nil, err
	}
	count, err := s.questions.CountByBank(ctx, bank.ID)
	if err != nil {
		return nil, err
	}
	return &domain.ShareCodePreview{
		BankName:      bank.Name,
		Description:   bank.Description,
		QuestionCount: count,
		ExpiresAt:     sc.ExpiresAt,
	}, nil
}

func (s *service) RedeemShareCode(ctx context.Context, dto domain.RedeemShareCodeDTO) (*domain.QuestionBank, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	if !caller.IsAdmin &&
		!caller.HasPermission(domain.PermQuestionBanksCreate) &&
		!caller.HasPermission(domain.PermQuestionBanksCreateAny) {
		return nil, domain.ErrForbidden
	}
	if caller.OrgID == nil {
		return nil, domain.ErrForbidden
	}
	_, source, err := s.resolveShareCode(ctx, dto.Code)
	if err != nil {
		return nil, err
	}
	if s.queue == nil {
		return nil, domain.ErrInternal
	}

	target := &domain.QuestionBank{
		OrganizationID: *caller.OrgID,
		Name:           source.Name,
		Description:    source.Description,
		Status:         domain.QuestionBankStatusCopying,
	}
	if err := s.repo.Create(ctx, target); err != nil {
		return nil, err
	}

	payload, err := json.Marshal(domain.QuestionBankCopyPayload{
		SourceBankID: source.ID,
		TargetBankID: target.ID,
	})
	if err == nil {
		_, err = s.queue.Enqueue(asynq.NewTask(domain.TypeQuestionBankCopy, payload), asynq.Queue(domain.QueueMedia))
	}
	if err != nil {
		// Without the task the shell would sit in 'copying' forever — undo it.
		if derr := s.repo.HardDelete(ctx, target.ID); derr != nil {
			s.logger.Error("rollback redeemed bank shell", "bank_id", target.ID.String(), "error", derr)
		}
		s.logger.Error("enqueue question bank copy", "source_bank_id", source.ID.String(), "error", err)
		return nil, domain.ErrInternal
	}

	s.logger.Info("question bank share code redeemed",
		"source_bank_id", source.ID.String(),
		"target_bank_id", target.ID.String(),
		"redeemed_by", caller.UserID.String(),
	)
	return target, nil
}

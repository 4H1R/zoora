package questionbanks_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/4H1R/zoora/internal/domain"
)

func TestShareCode_Active(t *testing.T) {
	now := time.Now()
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	assert.True(t, (&domain.QuestionBankShareCode{}).Active(now))
	assert.True(t, (&domain.QuestionBankShareCode{ExpiresAt: &future}).Active(now))
	assert.False(t, (&domain.QuestionBankShareCode{ExpiresAt: &past}).Active(now))
	assert.False(t, (&domain.QuestionBankShareCode{RevokedAt: &past}).Active(now))
}

func TestBankService_GenerateShareCode(t *testing.T) {
	orgID := uuid.New()
	ctx := staffCtx(orgID)
	bankID := uuid.New()
	bankRepo := &mockBankRepo{}
	qRepo := &mockQuestionRepo{}

	bankRepo.On("FindByID", ctx, bankID).
		Return(&domain.QuestionBank{ID: bankID, OrganizationID: orgID, Status: domain.QuestionBankStatusReady}, nil)
	bankRepo.On("RevokeActiveShareCodesByBank", ctx, bankID, mock.AnythingOfType("time.Time")).Return(nil)
	bankRepo.On("CreateShareCode", ctx, mock.AnythingOfType("*domain.QuestionBankShareCode")).Return(nil)

	svc := newTestBankService(bankRepo, qRepo)
	days := 7
	sc, err := svc.GenerateShareCode(ctx, bankID, domain.GenerateShareCodeDTO{ExpiresInDays: &days})

	assert.NoError(t, err)
	assert.Len(t, sc.Code, 10)
	assert.Equal(t, orgID, sc.OrganizationID)
	if assert.NotNil(t, sc.ExpiresAt) {
		assert.WithinDuration(t, time.Now().AddDate(0, 0, 7), *sc.ExpiresAt, time.Minute)
	}
	bankRepo.AssertExpectations(t)
}

func TestBankService_GenerateShareCode_OtherOrgForbidden(t *testing.T) {
	ctx := staffCtx(uuid.New())
	bankID := uuid.New()
	bankRepo := &mockBankRepo{}

	bankRepo.On("FindByID", ctx, bankID).
		Return(&domain.QuestionBank{ID: bankID, OrganizationID: uuid.New(), Status: domain.QuestionBankStatusReady}, nil)

	svc := newTestBankService(bankRepo, &mockQuestionRepo{})
	_, err := svc.GenerateShareCode(ctx, bankID, domain.GenerateShareCodeDTO{})

	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestBankService_GenerateShareCode_CopyingBankRejected(t *testing.T) {
	orgID := uuid.New()
	ctx := staffCtx(orgID)
	bankID := uuid.New()
	bankRepo := &mockBankRepo{}

	bankRepo.On("FindByID", ctx, bankID).
		Return(&domain.QuestionBank{ID: bankID, OrganizationID: orgID, Status: domain.QuestionBankStatusCopying}, nil)

	svc := newTestBankService(bankRepo, &mockQuestionRepo{})
	_, err := svc.GenerateShareCode(ctx, bankID, domain.GenerateShareCodeDTO{})

	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestBankService_PreviewShareCode_NormalizesInput(t *testing.T) {
	orgID := uuid.New()
	ctx := staffCtx(orgID)
	bankID := uuid.New()
	bankRepo := &mockBankRepo{}
	qRepo := &mockQuestionRepo{}

	bankRepo.On("FindShareCodeByCode", ctx, "ABCD23EFGH").
		Return(&domain.QuestionBankShareCode{BankID: bankID, Code: "ABCD23EFGH"}, nil)
	bankRepo.On("FindByID", ctx, bankID).
		Return(&domain.QuestionBank{ID: bankID, Name: "Physics", Description: "d", Status: domain.QuestionBankStatusReady}, nil)
	qRepo.On("CountByBank", ctx, bankID).Return(int64(12), nil)

	svc := newTestBankService(bankRepo, qRepo)
	preview, err := svc.PreviewShareCode(ctx, " abcd-23ef gh ")

	assert.NoError(t, err)
	assert.Equal(t, "Physics", preview.BankName)
	assert.Equal(t, int64(12), preview.QuestionCount)
	bankRepo.AssertExpectations(t)
}

func TestBankService_PreviewShareCode_ExpiredIsGenericError(t *testing.T) {
	ctx := staffCtx(uuid.New())
	past := time.Now().Add(-time.Hour)
	bankRepo := &mockBankRepo{}

	bankRepo.On("FindShareCodeByCode", ctx, "ABCD23EFGH").
		Return(&domain.QuestionBankShareCode{BankID: uuid.New(), Code: "ABCD23EFGH", ExpiresAt: &past}, nil)

	svc := newTestBankService(bankRepo, &mockQuestionRepo{})
	_, err := svc.PreviewShareCode(ctx, "ABCD23EFGH")

	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestBankService_RedeemShareCode_UnknownCodeIsGenericError(t *testing.T) {
	ctx := staffCtx(uuid.New())
	bankRepo := &mockBankRepo{}

	bankRepo.On("FindShareCodeByCode", ctx, mock.Anything).Return(nil, domain.ErrNotFound)

	svc := newTestBankService(bankRepo, &mockQuestionRepo{})
	_, err := svc.RedeemShareCode(ctx, domain.RedeemShareCodeDTO{Code: "NOPE123456"})

	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestBankService_RevokeShareCode(t *testing.T) {
	orgID := uuid.New()
	ctx := staffCtx(orgID)
	bankID := uuid.New()
	bankRepo := &mockBankRepo{}

	bankRepo.On("FindByID", ctx, bankID).
		Return(&domain.QuestionBank{ID: bankID, OrganizationID: orgID, Status: domain.QuestionBankStatusReady}, nil)
	bankRepo.On("RevokeActiveShareCodesByBank", ctx, bankID, mock.AnythingOfType("time.Time")).Return(nil)

	svc := newTestBankService(bankRepo, &mockQuestionRepo{})
	assert.NoError(t, svc.RevokeShareCode(ctx, bankID))
	bankRepo.AssertExpectations(t)
}

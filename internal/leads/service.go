package leads

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/database"
)

type service struct {
	repo         domain.LeadRepository
	orgRepo      domain.OrganizationRepository
	settingsRepo domain.OrganizationSettingsRepository
	userRepo     domain.UserRepository
	roleRepo     domain.RoleRepository
	tx           *database.Transactor
	logger       *slog.Logger
}

func NewService(
	repo domain.LeadRepository,
	orgRepo domain.OrganizationRepository,
	settingsRepo domain.OrganizationSettingsRepository,
	userRepo domain.UserRepository,
	roleRepo domain.RoleRepository,
	tx *database.Transactor,
	logger *slog.Logger,
) domain.LeadService {
	return &service{
		repo:         repo,
		orgRepo:      orgRepo,
		settingsRepo: settingsRepo,
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		tx:           tx,
		logger:       logger,
	}
}

func (s *service) requireAdmin(ctx context.Context) (domain.Caller, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok || !caller.IsAdmin {
		return domain.Caller{}, domain.ErrForbidden
	}
	return caller, nil
}

// Submit records a public lead. It is unauthenticated: no caller check. A
// non-empty honeypot silently drops the submission (returns nil, nil — the
// handler still responds 200 so bots learn nothing). An open lead with the same
// phone is updated in place rather than duplicated.
func (s *service) Submit(ctx context.Context, dto domain.CreateLeadDTO) (*domain.Lead, error) {
	if dto.Website != "" {
		s.logger.Info("lead honeypot tripped, dropping", "phone", dto.Phone)
		return nil, nil
	}

	if existing, err := s.repo.FindOpenByPhone(ctx, dto.Phone); err == nil {
		existing.Name = dto.Name
		existing.OrgName = dto.OrgName
		if dto.Plan != "" {
			existing.Plan = dto.Plan
		}
		if dto.Note != "" {
			existing.Note = dto.Note
		}
		if err := s.repo.Update(ctx, existing); err != nil {
			return nil, err
		}
		s.logger.Info("lead updated from resubmit", "lead_id", existing.ID.String())
		return existing, nil
	}

	lead := &domain.Lead{
		Name:    dto.Name,
		Phone:   dto.Phone,
		OrgName: dto.OrgName,
		Plan:    dto.Plan,
		Note:    dto.Note,
		Status:  domain.LeadStatusNew,
	}
	if err := s.repo.Create(ctx, lead); err != nil {
		return nil, err
	}
	s.logger.Info("lead submitted", "lead_id", lead.ID.String())
	return lead, nil
}

func (s *service) AdminList(ctx context.Context, q domain.AdminListLeadsQuery) ([]domain.Lead, int64, error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, 0, err
	}
	if q.ListParams.Page < 1 {
		q.ListParams.Page = 1
	}
	if q.ListParams.PageSize <= 0 {
		q.ListParams.PageSize = domain.DefaultPageSize
	}
	return s.repo.AdminList(ctx, q)
}

func (s *service) UpdateStatus(ctx context.Context, id uuid.UUID, dto domain.UpdateLeadStatusDTO) (*domain.Lead, error) {
	caller, err := s.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if !dto.Status.Valid() {
		return nil, domain.NewValidationError(map[string]string{"status": "unknown status"})
	}
	lead, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	lead.Status = dto.Status
	if err := s.repo.Update(ctx, lead); err != nil {
		return nil, err
	}
	s.logger.Info("admin updated lead status",
		"lead_id", id.String(),
		"status", string(dto.Status),
		"by", caller.UserID.String(),
	)
	return lead, nil
}

func (s *service) AdminHardDelete(ctx context.Context, id uuid.UUID) error {
	caller, err := s.requireAdmin(ctx)
	if err != nil {
		return err
	}
	if err := s.repo.HardDelete(ctx, id); err != nil {
		return err
	}
	s.logger.Warn("admin hard-deleted lead", "lead_id", id.String(), "by", caller.UserID.String())
	return nil
}

// Convert provisions the org (+ settings) and owner account from a lead, marks
// it converted, and links the new org — all in one transaction so a partial
// failure leaves no orphaned org.
func (s *service) Convert(ctx context.Context, id uuid.UUID, dto domain.ConvertLeadDTO) (*domain.Lead, error) {
	caller, err := s.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if err := domain.ValidateSlug(dto.Slug); err != nil {
		return nil, err
	}
	if !dto.Plan.Valid() {
		return nil, domain.NewValidationError(map[string]string{"plan": "unknown plan"})
	}

	lead, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if lead.Status == domain.LeadStatusConverted {
		return nil, domain.ErrConflict
	}

	managerRole, err := s.roleRepo.FindPresetByName(ctx, domain.PresetRoleManager)
	if err != nil {
		return nil, fmt.Errorf("resolving manager preset role: %w", err)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(dto.OwnerPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("leads.service.Convert hash: %w", err)
	}

	err = s.tx.RunInTx(ctx, func(ctx context.Context) error {
		org := &domain.Organization{
			Name:          dto.OrgName,
			Slug:          dto.Slug,
			Status:        domain.OrganizationStatusActive,
			Plan:          dto.Plan,
			PlanExpiresAt: dto.PlanExpiresAt,
		}
		if err := s.orgRepo.Create(ctx, org); err != nil {
			return err // ErrConflict on duplicate slug rolls the tx back
		}
		if err := s.settingsRepo.Create(ctx, domain.NewDefaultOrganizationSettings(org.ID)); err != nil {
			return fmt.Errorf("creating organization settings: %w", err)
		}
		owner := &domain.User{
			OrganizationID: &org.ID,
			Username:       dto.OwnerUsername,
			Name:           dto.OwnerName,
			Password:       string(hashed),
			RoleID:         &managerRole.ID,
		}
		if err := s.userRepo.Create(ctx, owner); err != nil {
			return err
		}
		lead.Status = domain.LeadStatusConverted
		lead.OrganizationID = &org.ID
		if err := s.repo.Update(ctx, lead); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.logger.Info("admin converted lead",
		"lead_id", id.String(),
		"org_id", lead.OrganizationID.String(),
		"slug", dto.Slug,
		"by", caller.UserID.String(),
	)
	return lead, nil
}

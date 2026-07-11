package imports

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"strings"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/entitlements"
)

const (
	maxImportFileSize = 10 << 20 // 10MB
	progressBatch     = 25
	passwordAlphabet  = "abcdefghjkmnpqrstuvwxyz23456789" // no ambiguous chars
)

type ObjectStore interface {
	GetObject(ctx context.Context, key string) ([]byte, error)
}

type Enqueuer interface {
	Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}

type ResultStore interface {
	Set(ctx context.Context, jobID uuid.UUID, data []byte) error
	Get(ctx context.Context, jobID uuid.UUID) ([]byte, error)
}

type service struct {
	repo    domain.ImportJobRepository
	users   domain.UserRepository
	roles   domain.RoleRepository
	classes domain.ClassRepository
	members domain.ClassMemberRepository
	media   domain.MediaRepository
	ent     entitlements.Service
	storage ObjectStore
	queue   Enqueuer
	results ResultStore
	logger  *slog.Logger
}

func NewService(
	repo domain.ImportJobRepository,
	users domain.UserRepository,
	roles domain.RoleRepository,
	classes domain.ClassRepository,
	members domain.ClassMemberRepository,
	media domain.MediaRepository,
	ent entitlements.Service,
	storage ObjectStore,
	queue Enqueuer,
	results ResultStore,
	logger *slog.Logger,
) domain.ImportService {
	return &service{
		repo: repo, users: users, roles: roles, classes: classes, members: members,
		media: media, ent: ent, storage: storage, queue: queue, results: results, logger: logger,
	}
}

// requireImportPermission gates both job creation and visibility: user
// imports need users:create, class imports need classes:create_any (they
// assign arbitrary owners — the exact capability create_any exists for).
func requireImportPermission(caller domain.Caller, t domain.ImportType) error {
	if caller.IsAdmin {
		return nil
	}
	switch t {
	case domain.ImportTypeUsers:
		if caller.HasPermission(domain.PermUsersCreate) {
			return nil
		}
	case domain.ImportTypeClasses:
		if caller.HasPermission(domain.PermClassesCreateAny) {
			return nil
		}
	}
	return domain.ErrForbidden
}

func (s *service) Create(ctx context.Context, dto domain.CreateImportJobDTO) (*domain.ImportJob, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok || caller.OrgID == nil {
		return nil, domain.ErrForbidden
	}
	if err := requireImportPermission(caller, dto.Type); err != nil {
		return nil, err
	}

	m, err := s.media.FindByID(ctx, dto.MediaID)
	if err != nil {
		return nil, err
	}
	if m.OrganizationID == nil || *m.OrganizationID != *caller.OrgID {
		return nil, domain.ErrNotFound
	}
	if m.Size > maxImportFileSize {
		return nil, domain.NewValidationError(map[string]string{"media_id": "file exceeds 10MB"})
	}
	if !strings.HasSuffix(strings.ToLower(m.FileName), ".xlsx") {
		return nil, domain.NewValidationError(map[string]string{"media_id": "file must be .xlsx"})
	}

	job := &domain.ImportJob{
		OrganizationID: *caller.OrgID,
		UserID:         caller.UserID,
		MediaID:        dto.MediaID,
		Type:           dto.Type,
		Status:         domain.ImportStatusPending,
	}
	if err := s.repo.Create(ctx, job); err != nil {
		return nil, err
	}

	payload := domain.ImportProcessPayload{
		JobID:       job.ID,
		UserID:      caller.UserID,
		OrgID:       *caller.OrgID,
		IsAdmin:     caller.IsAdmin,
		Permissions: caller.Permissions,
		Plan:        caller.Ent.Plan,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("imports.service.Create marshal: %w", err)
	}
	task := asynq.NewTask(domain.TypeImportProcess, b)
	// MaxRetry(0): a crashed half-done import must not silently re-run;
	// TaskID dedupes double-submits of the same job.
	if _, err := s.queue.Enqueue(task, asynq.Queue(domain.QueueDefault), asynq.MaxRetry(0), asynq.TaskID("import:"+job.ID.String())); err != nil {
		msg := "failed to enqueue import task"
		job.Status = domain.ImportStatusFailed
		job.Error = &msg
		_ = s.repo.Update(ctx, job)
		return nil, fmt.Errorf("imports.service.Create enqueue: %w", err)
	}
	s.logger.Info("import job created", "job_id", job.ID.String(), "type", string(job.Type), "created_by", caller.UserID.String())
	return job, nil
}

func (s *service) Get(ctx context.Context, id uuid.UUID) (*domain.ImportJob, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok || caller.OrgID == nil {
		return nil, domain.ErrForbidden
	}
	job, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if job.OrganizationID != *caller.OrgID {
		return nil, domain.ErrNotFound
	}
	if err := requireImportPermission(caller, job.Type); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *service) Latest(ctx context.Context, t domain.ImportType) (*domain.ImportJob, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok || caller.OrgID == nil {
		return nil, domain.ErrForbidden
	}
	if err := requireImportPermission(caller, t); err != nil {
		return nil, err
	}
	job, err := s.repo.Latest(ctx, *caller.OrgID, t)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, nil // "no job yet" is a normal answer, not an error
	}
	return job, err
}

func (s *service) Result(ctx context.Context, id uuid.UUID) ([]byte, error) {
	if _, err := s.Get(ctx, id); err != nil {
		return nil, err
	}
	return s.results.Get(ctx, id)
}

// ProcessJob is the Asynq entrypoint. It reconstructs the enqueue-time
// caller (worker ctx has no auth middleware) and dispatches by type.
func (s *service) ProcessJob(ctx context.Context, p domain.ImportProcessPayload) error {
	job, err := s.repo.FindByID(ctx, p.JobID)
	if err != nil {
		return err
	}
	if job.Status != domain.ImportStatusPending {
		return nil // already handled (dup delivery / manual replay)
	}
	job.Status = domain.ImportStatusProcessing
	if err := s.repo.Update(ctx, job); err != nil {
		return err
	}

	caller := domain.Caller{
		UserID:      p.UserID,
		OrgID:       &p.OrgID,
		IsAdmin:     p.IsAdmin,
		Permissions: p.Permissions,
		Ent:         domain.PlanCatalog[p.Plan],
	}
	ctx = domain.WithCaller(ctx, caller)

	m, err := s.media.FindByID(ctx, job.MediaID)
	if err != nil {
		return s.fail(ctx, job, "uploaded file not found")
	}
	data, err := s.storage.GetObject(ctx, m.S3Key())
	if err != nil {
		s.logger.Error("import download failed", "job_id", job.ID.String(), "error", err)
		return s.fail(ctx, job, "could not read uploaded file")
	}

	switch job.Type {
	case domain.ImportTypeUsers:
		return s.processUsers(ctx, job, data)
	case domain.ImportTypeClasses:
		return s.processClasses(ctx, job, data)
	default:
		return s.fail(ctx, job, "unknown import type")
	}
}

// fail marks the job failed with a user-facing reason and swallows the
// task error — MaxRetry is 0 and a retry could double-create rows.
func (s *service) fail(ctx context.Context, job *domain.ImportJob, msg string) error {
	job.Status = domain.ImportStatusFailed
	job.Error = &msg
	if err := s.repo.Update(ctx, job); err != nil {
		return err
	}
	s.logger.Warn("import job failed", "job_id", job.ID.String(), "reason", msg)
	return nil
}

func generatePassword() (string, error) {
	b := make([]byte, 10)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(passwordAlphabet))))
		if err != nil {
			return "", fmt.Errorf("imports.generatePassword: %w", err)
		}
		b[i] = passwordAlphabet[n.Int64()]
	}
	return string(b), nil
}

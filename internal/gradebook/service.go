package gradebook

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/authz"
)

// service implements domain.GradebookService. RBAC hierarchy mirrors classes:
//
//	super-admin (caller.IsAdmin): full access
//	gradebook:*_any permission: full access within org
//	teacher (class.UserID == caller.UserID): manage gradebook of own class
//	student (enrolled via class_members): view-only
type service struct {
	columns      domain.GradebookColumnRepository
	cells        domain.GradebookCellRepository
	classes      domain.ClassRepository
	members      domain.ClassMemberRepository
	attendance   domain.AttendanceRepository
	practiceSubs domain.PracticeSubmissionRepository
	quizSubs     domain.QuizSubmissionRepository
	resolver     *authz.Resolver
	logger       *slog.Logger
}

func NewService(
	columns domain.GradebookColumnRepository,
	cells domain.GradebookCellRepository,
	classes domain.ClassRepository,
	members domain.ClassMemberRepository,
	attendance domain.AttendanceRepository,
	practiceSubs domain.PracticeSubmissionRepository,
	quizSubs domain.QuizSubmissionRepository,
	resolver *authz.Resolver,
	logger *slog.Logger,
) domain.GradebookService {
	return &service{
		columns:      columns,
		cells:        cells,
		classes:      classes,
		members:      members,
		attendance:   attendance,
		practiceSubs: practiceSubs,
		quizSubs:     quizSubs,
		resolver:     resolver,
		logger:       logger,
	}
}

// canManageGradebook returns true if caller can mutate gradebook columns/cells
// of the given class. Students never qualify here.
func canManageGradebook(caller domain.Caller, class *domain.Class) bool {
	return caller.CanManage(class.UserID, domain.PermGradebookUpdateAny)
}

func canDeleteGradebook(caller domain.Caller, class *domain.Class) bool {
	return caller.CanManage(class.UserID, domain.PermGradebookDeleteAny)
}

// viewScope resolves how much of the gradebook the caller may read: admins and
// gradebook:view_any holders get the whole org (ScopeAll), the class owner gets
// the class matrix (ScopeClass), enrolled students get only their own row
// (ScopeOwn), and everyone else is denied (ScopeNone).
func (s *service) viewScope(ctx context.Context, caller domain.Caller, class *domain.Class) (authz.Scope, error) {
	return s.resolver.Scope(ctx, caller, class, domain.PermGradebookViewAny)
}

func (s *service) CreateColumn(ctx context.Context, classID uuid.UUID, dto domain.CreateGradebookColumnDTO) (*domain.GradebookColumn, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	class, err := s.classes.FindByID(ctx, classID)
	if err != nil {
		return nil, err
	}
	if !canManageGradebook(caller, class) {
		return nil, domain.ErrForbidden
	}
	if dto.Type.IsAuto() && dto.SourceID == nil {
		return nil, domain.NewValidationError(map[string]string{"source_id": "required for auto column types"})
	}
	col := &domain.GradebookColumn{
		ClassID:    classID,
		Title:      dto.Title,
		Type:       dto.Type,
		SourceID:   dto.SourceID,
		MaxScore:   dto.MaxScore,
		OrderIndex: dto.OrderIndex,
	}
	if err := s.columns.Create(ctx, col); err != nil {
		return nil, err
	}
	s.logger.Info("gradebook column created",
		"column_id", col.ID.String(),
		"class_id", classID.String(),
		"created_by", caller.UserID.String(),
	)
	return col, nil
}

func (s *service) UpdateColumn(ctx context.Context, columnID uuid.UUID, dto domain.UpdateGradebookColumnDTO) (*domain.GradebookColumn, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	col, err := s.columns.FindByID(ctx, columnID)
	if err != nil {
		return nil, err
	}
	class, err := s.classes.FindByID(ctx, col.ClassID)
	if err != nil {
		return nil, err
	}
	if !canManageGradebook(caller, class) {
		return nil, domain.ErrForbidden
	}
	if dto.Title != nil {
		col.Title = *dto.Title
	}
	if dto.MaxScore != nil {
		col.MaxScore = dto.MaxScore
	}
	if dto.OrderIndex != nil {
		col.OrderIndex = *dto.OrderIndex
	}
	if err := s.columns.Update(ctx, col); err != nil {
		return nil, err
	}
	return col, nil
}

func (s *service) DeleteColumn(ctx context.Context, columnID uuid.UUID) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	col, err := s.columns.FindByID(ctx, columnID)
	if err != nil {
		return err
	}
	class, err := s.classes.FindByID(ctx, col.ClassID)
	if err != nil {
		return err
	}
	if !canDeleteGradebook(caller, class) {
		return domain.ErrForbidden
	}
	if err := s.columns.Delete(ctx, columnID); err != nil {
		return err
	}
	s.logger.Info("gradebook column deleted",
		"column_id", columnID.String(),
		"class_id", class.ID.String(),
		"deleted_by", caller.UserID.String(),
	)
	return nil
}

func (s *service) UpsertCell(ctx context.Context, classID, columnID uuid.UUID, dto domain.UpsertGradebookCellDTO) (*domain.GradebookCell, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	col, err := s.columns.FindByID(ctx, columnID)
	if err != nil {
		return nil, err
	}
	if col.ClassID != classID {
		return nil, domain.ErrNotFound
	}
	if col.Type.IsAuto() {
		return nil, domain.NewValidationError(map[string]string{"type": "cannot manually set value for auto columns"})
	}
	class, err := s.classes.FindByID(ctx, col.ClassID)
	if err != nil {
		return nil, err
	}
	if !canManageGradebook(caller, class) {
		return nil, domain.ErrForbidden
	}
	cell := &domain.GradebookCell{
		ColumnID:  columnID,
		StudentID: dto.StudentID,
		Value:     dto.Value,
		UpdatedAt: time.Now(),
	}
	if err := s.cells.Upsert(ctx, cell); err != nil {
		return nil, err
	}
	return cell, nil
}

func (s *service) ListColumns(ctx context.Context, classID uuid.UUID, q domain.ListGradebookColumnsQuery) ([]domain.GradebookColumn, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	class, err := s.classes.FindByID(ctx, classID)
	if err != nil {
		return nil, 0, err
	}
	scope, err := s.viewScope(ctx, caller, class)
	if err != nil {
		return nil, 0, err
	}
	if scope == authz.ScopeNone {
		return nil, 0, domain.ErrForbidden
	}
	return s.columns.ListByClass(ctx, classID, q)
}

func (s *service) GetMatrix(ctx context.Context, classID uuid.UUID) (*domain.GradebookMatrix, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	class, err := s.classes.FindByID(ctx, classID)
	if err != nil {
		return nil, err
	}
	scope, err := s.viewScope(ctx, caller, class)
	if err != nil {
		return nil, err
	}
	if scope == authz.ScopeNone {
		return nil, domain.ErrForbidden
	}
	return s.buildMatrix(ctx, caller, class, scope)
}

// buildMatrix assembles the matrix for an already-resolved class+scope. Callers
// (GetMatrix, GetMine) are responsible for loading the class and resolving the
// caller's scope first, so the class is not re-fetched here.
func (s *service) buildMatrix(ctx context.Context, caller domain.Caller, class *domain.Class, scope authz.Scope) (*domain.GradebookMatrix, error) {
	classID := class.ID

	columns, err := s.columns.ListAllByClass(ctx, classID)
	if err != nil {
		return nil, err
	}

	members, err := s.members.ListAllByClass(ctx, classID)
	if err != nil {
		return nil, err
	}

	// An enrolled student may only see their own row, never classmates' grades.
	if scope == authz.ScopeOwn {
		own := make([]domain.ClassMember, 0, 1)
		for _, m := range members {
			if m.UserID == caller.UserID {
				own = append(own, m)
			}
		}
		members = own
	}

	studentIDs := make([]uuid.UUID, len(members))
	studentByID := make(map[uuid.UUID]*domain.User, len(members))
	for i, m := range members {
		studentIDs[i] = m.UserID
		studentByID[m.UserID] = m.User
	}

	var manualColIDs []uuid.UUID
	for _, col := range columns {
		if !col.Type.IsAuto() {
			manualColIDs = append(manualColIDs, col.ID)
		}
	}

	cells, err := s.cells.ListByColumns(ctx, manualColIDs)
	if err != nil {
		return nil, err
	}

	cellIndex := make(map[string]string)
	for _, c := range cells {
		key := c.ColumnID.String() + ":" + c.StudentID.String()
		cellIndex[key] = c.Value
	}

	autoData := make(map[string]map[uuid.UUID]string)
	for _, col := range columns {
		if !col.Type.IsAuto() || col.SourceID == nil {
			continue
		}
		data, err := s.fetchAutoData(ctx, col)
		if err != nil {
			s.logger.Warn("failed to fetch auto data for gradebook column",
				"column_id", col.ID.String(),
				"type", string(col.Type),
				"error", err,
			)
			continue
		}
		autoData[col.ID.String()] = data
	}

	rows := make([]domain.GradebookMatrixRow, 0, len(studentIDs))
	for _, sid := range studentIDs {
		row := domain.GradebookMatrixRow{
			StudentID: sid,
			Student:   studentByID[sid],
			Cells:     make(map[string]string, len(columns)),
		}
		for _, col := range columns {
			colIDStr := col.ID.String()
			if col.Type.IsAuto() {
				if data, ok := autoData[colIDStr]; ok {
					row.Cells[colIDStr] = data[sid]
				}
			} else {
				key := colIDStr + ":" + sid.String()
				row.Cells[colIDStr] = cellIndex[key]
			}
		}
		rows = append(rows, row)
	}

	return &domain.GradebookMatrix{
		Columns: columns,
		Rows:    rows,
	}, nil
}

// GetMine builds the caller's own report card across every class they are
// enrolled in. It reuses GetMatrix per class (which resolves auto-computed
// columns too) and extracts only the caller's row.
func (s *service) GetMine(ctx context.Context) (*domain.MyGradebook, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	uid := caller.UserID

	classes, _, err := s.classes.List(ctx, domain.ClassListScope{MemberUserID: &uid}, domain.ListParams{Page: 1, PageSize: 10000})
	if err != nil {
		return nil, fmt.Errorf("listing my classes: %w", err)
	}

	out := &domain.MyGradebook{Classes: make([]domain.MyGradebookClass, 0, len(classes))}
	for i := range classes {
		cl := classes[i]
		// The class is already loaded from List above, so resolve scope and build
		// the matrix directly instead of GetMatrix re-fetching it by ID.
		scope, err := s.viewScope(ctx, caller, &cl)
		if err != nil {
			return nil, fmt.Errorf("resolving scope for class %s: %w", cl.ID, err)
		}
		if scope == authz.ScopeNone {
			// Skip classes the student can't view rather than failing the whole card.
			continue
		}
		matrix, err := s.buildMatrix(ctx, caller, &cl, scope)
		if err != nil {
			if errors.Is(err, domain.ErrForbidden) {
				continue
			}
			return nil, fmt.Errorf("building report card for class %s: %w", cl.ID, err)
		}
		own := make(map[string]string, len(matrix.Columns))
		for r := range matrix.Rows {
			if matrix.Rows[r].StudentID == uid {
				maps.Copy(own, matrix.Rows[r].Cells)
				break
			}
		}
		out.Classes = append(out.Classes, domain.MyGradebookClass{
			ClassID:   cl.ID,
			ClassName: cl.Name,
			Columns:   matrix.Columns,
			Cells:     own,
		})
	}
	return out, nil
}

func (s *service) fetchAutoData(ctx context.Context, col domain.GradebookColumn) (map[uuid.UUID]string, error) {
	result := make(map[uuid.UUID]string)
	sourceID := *col.SourceID

	switch col.Type {
	case domain.GradebookColumnAutoAttendance:
		attendances, _, err := s.attendance.ListBySession(ctx, sourceID, domain.ListAttendanceQuery{
			ListParams: domain.ListParams{Page: 1, PageSize: 1000},
		})
		if err != nil {
			return nil, fmt.Errorf("fetching attendance for session %s: %w", sourceID, err)
		}
		for _, a := range attendances {
			result[a.UserID] = string(a.Status)
		}

	case domain.GradebookColumnAutoPractice:
		subs, _, err := s.practiceSubs.ListByRoom(ctx, sourceID, domain.ListParams{Page: 1, PageSize: 1000})
		if err != nil {
			return nil, fmt.Errorf("fetching practice submissions for room %s: %w", sourceID, err)
		}
		for _, sub := range subs {
			if sub.Score != nil {
				result[sub.UserID] = fmt.Sprintf("%.1f", *sub.Score)
			} else {
				result[sub.UserID] = "submitted"
			}
		}

	case domain.GradebookColumnAutoQuiz:
		subs, _, err := s.quizSubs.ListByQuiz(ctx, sourceID, domain.ListSubmissionsQuery{
			ListParams: domain.ListParams{Page: 1, PageSize: 1000},
		})
		if err != nil {
			return nil, fmt.Errorf("fetching quiz submissions for quiz %s: %w", sourceID, err)
		}
		for _, sub := range subs {
			if sub.Status == domain.SubmissionStatusGraded {
				result[sub.UserID] = fmt.Sprintf("%.1f", sub.TotalScore)
			} else {
				result[sub.UserID] = string(sub.Status)
			}
		}
	}

	return result, nil
}

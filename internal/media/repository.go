package media

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/database"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) domain.MediaRepository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, m *domain.Media) error {
	if err := database.DB(ctx, r.db).Create(m).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("media.repository.Create: %w", err)
	}
	return nil
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Media, error) {
	var m domain.Media
	if err := database.DB(ctx, r.db).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("media.repository.FindByID: %w", err)
	}
	return &m, nil
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Delete(&domain.Media{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("media.repository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *repository) ListFolders(ctx context.Context, orgID uuid.UUID) ([]domain.MediaFolder, error) {
	var folders []domain.MediaFolder
	err := database.DB(ctx, r.db).
		Model(&domain.Media{}).
		Select("model_type, COUNT(*) AS file_count, COALESCE(SUM(size), 0) AS total_size").
		Where("organization_id = ?", orgID).
		Group("model_type").
		Order("model_type ASC").
		Scan(&folders).Error
	if err != nil {
		return nil, fmt.Errorf("media.repository.ListFolders: %w", err)
	}
	return folders, nil
}

func (r *repository) ListFiles(ctx context.Context, orgID uuid.UUID, modelType string, p domain.ListParams) ([]domain.Media, int64, error) {
	base := database.DB(ctx, r.db).
		Model(&domain.Media{}).
		Where("organization_id = ? AND model_type = ?", orgID, modelType)
	var items []domain.Media
	total, err := listparams.Paginate(base, p, &items)
	if err != nil {
		return nil, 0, fmt.Errorf("media.repository.ListFiles: %w", err)
	}
	return items, total, nil
}

// ownerResolverSQL resolves every media row to its named parent owner. Each
// row keeps its media columns plus a computed owner_id/owner_name reached by
// LEFT JOINing the owning entity for its model_type:
//   - question       -> question_bank (questions.bank_id)
//   - conversation   -> conversation
//   - live_room      -> class (via class_sessions)
//   - practice       -> class (practice_rooms.class_id)
//   - offline_room   -> class (offline_rooms.class_id)
//   - ticket         -> class (model_id IS the class id)
//   - anything else  -> NULL (falls into the "other" bucket downstream)
//
// The named arg @org binds media.organization_id.
const ownerResolverSQL = `
SELECT m.id, m.model_type, m.name, m.file_name, m.mime_type, m.size, m.created_at,
  CASE
    WHEN m.model_type = 'question'     THEN qb.id
    WHEN m.model_type = 'conversation' THEN conv.id
    WHEN m.model_type = 'live_room'    THEN lrcls.id
    WHEN m.model_type = 'practice'     THEN pr.class_id
    WHEN m.model_type = 'offline_room' THEN off.class_id
    WHEN m.model_type = 'ticket'       THEN tkcls.id
    ELSE NULL
  END AS owner_id,
  CASE
    WHEN m.model_type = 'question'     THEN qb.name
    WHEN m.model_type = 'conversation' THEN conv.name
    WHEN m.model_type = 'live_room'    THEN lrcls.name
    WHEN m.model_type = 'practice'     THEN prcls.name
    WHEN m.model_type = 'offline_room' THEN offcls.name
    WHEN m.model_type = 'ticket'       THEN tkcls.name
    ELSE NULL
  END AS owner_name
FROM media m
LEFT JOIN questions q          ON m.model_type = 'question'     AND q.id = m.model_id
LEFT JOIN question_banks qb    ON qb.id = q.bank_id
LEFT JOIN conversations conv   ON m.model_type = 'conversation' AND conv.id = m.model_id
LEFT JOIN live_rooms lr        ON m.model_type = 'live_room'    AND lr.id = m.model_id
LEFT JOIN class_sessions lrcs  ON lrcs.id = lr.class_session_id
LEFT JOIN classes lrcls        ON lrcls.id = lrcs.class_id
LEFT JOIN practice_rooms pr    ON m.model_type = 'practice'     AND pr.id = m.model_id
LEFT JOIN classes prcls        ON prcls.id = pr.class_id
LEFT JOIN offline_rooms off    ON m.model_type = 'offline_room' AND off.id = m.model_id
LEFT JOIN classes offcls       ON offcls.id = off.class_id
LEFT JOIN classes tkcls        ON m.model_type = 'ticket'       AND tkcls.id = m.model_id
WHERE m.organization_id = @org`

// ownerKindExpr maps a resolved row to its owner kind. organization -> shared;
// an unresolved owner_id -> other; otherwise the entity kind.
const ownerKindExpr = `
  CASE
    WHEN r.model_type = 'organization' THEN 'shared'
    WHEN r.owner_id IS NULL            THEN 'other'
    WHEN r.model_type = 'question'     THEN 'question_bank'
    WHEN r.model_type = 'conversation' THEN 'conversation'
    ELSE 'class'
  END`

func (r *repository) ListOwnerMedia(ctx context.Context, orgID uuid.UUID) ([]domain.MediaOwner, error) {
	q := `
SELECT owner_kind, owner_id, MAX(owner_name) AS name,
       COUNT(*) AS file_count, COALESCE(SUM(size), 0) AS total_size
FROM (
  SELECT r.size, r.owner_id, r.owner_name, ` + ownerKindExpr + ` AS owner_kind
  FROM (` + ownerResolverSQL + `) r
) g
GROUP BY owner_kind, owner_id`
	var owners []domain.MediaOwner
	if err := database.DB(ctx, r.db).Raw(q, sql.Named("org", orgID)).Scan(&owners).Error; err != nil {
		return nil, fmt.Errorf("media.repository.ListOwnerMedia: %w", err)
	}
	return owners, nil
}

// ownerFilesColumns maps the white-listed order_by values to concrete columns,
// guarding the dynamically-built ORDER BY against injection.
var ownerFilesColumns = map[string]string{
	"created_at": "created_at",
	"size":       "size",
	"name":       "name",
}

func (r *repository) ListOwnerFiles(ctx context.Context, orgID uuid.UUID, ownerKind string, ownerID *uuid.UUID, p domain.ListParams) ([]domain.OwnerFile, int64, error) {
	// The media rows for the owner UNION ALL the class's recordings. The
	// recordings branch self-nullifies for non-class owners (@kind guard) and
	// when ownerID is nil (cls.id = NULL matches nothing), so one query serves
	// every kind. Pagination + COUNT happen in SQL — no loading all rows.
	inner := `
SELECT id, source, model_type, name, mime_type, size, deletable, created_at FROM (
  SELECT z.id::text AS id, 'media' AS source, z.model_type,
         COALESCE(NULLIF(z.name, ''), z.file_name) AS name,
         z.mime_type, z.size, TRUE AS deletable, z.created_at
  FROM (
    SELECT r.id, r.model_type, r.name, r.file_name, r.mime_type, r.size, r.created_at,
           r.owner_id, ` + ownerKindExpr + ` AS owner_kind
    FROM (` + ownerResolverSQL + `) r
  ) z
  WHERE z.owner_kind = @kind AND (z.owner_id = @owner_id OR (@owner_id IS NULL AND z.owner_id IS NULL))

  UNION ALL

  SELECT rec.id::text, 'recording', 'recording',
         'Recording ' || to_char(rec.started_at, 'YYYY-MM-DD HH24:MI'),
         'video/mp4', rec.size, FALSE, rec.started_at
  FROM live_recordings rec
  JOIN live_rooms lr     ON lr.id = rec.live_room_id
  JOIN class_sessions cs ON cs.id = lr.class_session_id
  JOIN classes cls       ON cls.id = cs.class_id
  WHERE @kind = 'class' AND cls.organization_id = @org AND cls.id = @owner_id
) f
WHERE (@search = '' OR name ILIKE @like)`

	args := map[string]any{
		"org":      orgID,
		"kind":     ownerKind,
		"owner_id": ownerID,
		"search":   p.Search,
		"like":     "%" + p.Search + "%",
	}

	var total int64
	if err := database.DB(ctx, r.db).Raw(`SELECT COUNT(*) FROM (`+inner+`) c`, args).Scan(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("media.repository.ListOwnerFiles count: %w", err)
	}

	orderCol := ownerFilesColumns[p.OrderBy]
	if orderCol == "" {
		orderCol = "created_at"
	}
	orderDir := "DESC"
	if strings.EqualFold(p.OrderDir, "asc") {
		orderDir = "ASC"
	}
	// name/size ties on recordings share a timestamp — created_at, id break ties
	// for a stable page order.
	paged := inner + fmt.Sprintf(" ORDER BY %s %s, created_at DESC, id ASC LIMIT @limit OFFSET @offset", orderCol, orderDir)
	args["limit"] = p.Limit()
	args["offset"] = p.Offset()

	var files []domain.OwnerFile
	if err := database.DB(ctx, r.db).Raw(paged, args).Scan(&files).Error; err != nil {
		return nil, 0, fmt.Errorf("media.repository.ListOwnerFiles: %w", err)
	}
	return files, total, nil
}

func (r *repository) ListOwnerRecordings(ctx context.Context, orgID uuid.UUID) ([]domain.MediaOwner, error) {
	q := `
SELECT cls.id AS owner_id, cls.name AS name,
       COUNT(*) AS file_count, COALESCE(SUM(rec.size), 0) AS total_size
FROM live_recordings rec
JOIN live_rooms lr    ON lr.id = rec.live_room_id
JOIN class_sessions cs ON cs.id = lr.class_session_id
JOIN classes cls      ON cls.id = cs.class_id
WHERE cls.organization_id = @org
GROUP BY cls.id, cls.name`
	var owners []domain.MediaOwner
	if err := database.DB(ctx, r.db).Raw(q, sql.Named("org", orgID)).Scan(&owners).Error; err != nil {
		return nil, fmt.Errorf("media.repository.ListOwnerRecordings: %w", err)
	}
	for i := range owners {
		owners[i].OwnerKind = domain.MediaOwnerClass
	}
	return owners, nil
}

func (r *repository) ListByModel(ctx context.Context, modelType string, modelID uuid.UUID, collection string) ([]domain.Media, error) {
	q := database.DB(ctx, r.db).
		Where("model_type = ? AND model_id = ?", modelType, modelID).
		Order("order_column ASC, created_at ASC")
	if collection != "" {
		q = q.Where("collection_name = ?", collection)
	}
	var items []domain.Media
	if err := q.Find(&items).Error; err != nil {
		return nil, fmt.Errorf("media.repository.ListByModel: %w", err)
	}
	return items, nil
}

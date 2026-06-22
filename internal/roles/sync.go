package roles

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/cache"
	"github.com/4H1R/zoora/internal/platform/database"
)

// SyncPermissions reconciles the permissions table and preset-role grants with
// the code-defined source of truth (domain.AllPermissions plus the preset
// permission sets). It runs at API startup so renaming, adding, or removing a
// permission constant takes effect on an existing database without a
// destructive reseed: missing permission names are inserted, obsolete ones are
// deleted (cascading their role_permissions grants), and each preset role's
// grants are re-synced to its code-defined set so renamed permissions re-attach.
//
// Custom (non-preset) roles keep whatever still-valid grants they hold; a grant
// pointing at a deleted permission is removed by the cascade and must be
// re-assigned by an admin — there is no generic old→new name mapping for those.
func SyncPermissions(
	ctx context.Context,
	db *gorm.DB,
	transactor *database.Transactor,
	roleRepo domain.RoleRepository,
	permRepo domain.PermissionRepository,
	rdb *redis.Client,
	log *slog.Logger,
) error {
	// presetRoleIDs collects the preset roles touched so their cached permission
	// sets can be invalidated after the transaction commits.
	var presetRoleIDs []uuid.UUID
	err := transactor.RunInTx(ctx, func(ctx context.Context) error {
		tx := database.DB(ctx, db)

		// 1. Insert any missing permission names (idempotent on the name unique).
		want := make([]domain.Permission, 0, len(domain.AllPermissions))
		names := make([]string, 0, len(domain.AllPermissions))
		for _, n := range domain.AllPermissions {
			want = append(want, domain.Permission{Name: n})
			names = append(names, string(n))
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoNothing: true,
		}).Create(&want).Error; err != nil {
			return fmt.Errorf("upserting permissions: %w", err)
		}

		// 2. Delete permissions no longer defined in code; role_permissions
		//    rows referencing them cascade away (FK ON DELETE CASCADE).
		pruned := tx.Where("name NOT IN ?", names).Delete(&domain.Permission{})
		if pruned.Error != nil {
			return fmt.Errorf("pruning permissions: %w", pruned.Error)
		}

		// 3. Resolve the current name -> id map.
		all, err := permRepo.List(ctx)
		if err != nil {
			return fmt.Errorf("listing permissions: %w", err)
		}
		idByName := make(map[domain.PermissionName]uuid.UUID, len(all))
		for _, p := range all {
			idByName[p.Name] = p.ID
		}

		// 4. Re-sync each preset role's grants to its code-defined set.
		presets := []struct {
			Name  string
			Perms []domain.PermissionName
		}{
			{domain.PresetRoleManager, domain.ManagerPermissions},
			{domain.PresetRoleTeacher, domain.TeacherPermissions},
			{domain.PresetRoleStudent, domain.StudentPermissions},
		}
		for _, preset := range presets {
			role, err := roleRepo.FindPresetByName(ctx, preset.Name)
			if errors.Is(err, domain.ErrNotFound) {
				continue // preset not seeded on this DB; nothing to reconcile
			}
			if err != nil {
				return fmt.Errorf("loading preset role %s: %w", preset.Name, err)
			}
			ids := make([]uuid.UUID, 0, len(preset.Perms))
			for _, pn := range preset.Perms {
				if id, ok := idByName[pn]; ok {
					ids = append(ids, id)
				}
			}
			if err := roleRepo.SetPermissions(ctx, role.ID, ids); err != nil {
				return fmt.Errorf("syncing preset role %s grants: %w", preset.Name, err)
			}
			presetRoleIDs = append(presetRoleIDs, role.ID)
		}

		log.Info("permission sync complete",
			"total", len(all),
			"removed", pruned.RowsAffected,
		)
		return nil
	})
	if err != nil {
		return err
	}

	// Drop stale cached permission sets for the re-synced preset roles so the
	// new grants take effect immediately instead of after the cache TTL.
	for _, id := range presetRoleIDs {
		if cerr := cache.InvalidateRolePermissions(ctx, rdb, id); cerr != nil {
			log.Warn("failed to invalidate role permission cache", "role_id", id, "error", cerr)
		}
	}
	return nil
}

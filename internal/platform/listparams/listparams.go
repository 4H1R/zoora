// Package listparams is a shared helper for the standardized List API
// pattern: pagination + multi-field search + white-listed ordering.
//
// The handler owns the ListConfig (what is searchable / orderable) and
// binds query params into a domain.ListParams. The repository feeds the
// resulting ListParams into Apply + Paginate to produce a SQL query that
// is safe against injection — both SearchFields and OrderBy are drawn
// only from the handler's white-list, never directly from the client.
package listparams

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
)

// Bind parses page/search/order_by/order_dir from the request's query string
// and white-lists them against cfg. Disallowed values silently fall back to
// the defaults — invalid client input never produces a 400 here.
func Bind(c *gin.Context, cfg domain.ListConfig) domain.ListParams {
	pageSize := httpx.QueryInt(c, "page_size", 0)
	if pageSize <= 0 {
		pageSize = cfg.PageSize
	}
	if pageSize <= 0 {
		pageSize = domain.DefaultPageSize
	}
	pageSize = min(pageSize, domain.MaxPageSize)

	page := httpx.QueryInt(c, "page", 1)
	if page < 1 {
		page = 1
	}

	p := domain.ListParams{
		Page:     page,
		PageSize: pageSize,
	}

	if search := strings.TrimSpace(c.Query("search")); search != "" && len(cfg.AllowedSearchFields) > 0 {
		p.Search = search
		p.SearchFields = cfg.AllowedSearchFields
	}

	orderBy := strings.TrimSpace(c.Query("order_by"))
	if orderBy != "" && contains(cfg.AllowedOrderFields, orderBy) {
		p.OrderBy = orderBy
	} else {
		p.OrderBy = cfg.DefaultOrderBy
	}

	if p.OrderBy != "" {
		switch strings.ToLower(c.Query("order_dir")) {
		case "asc":
			p.OrderDir = "asc"
		case "desc":
			p.OrderDir = "desc"
		default:
			p.OrderDir = strings.ToLower(cfg.DefaultOrderDir)
			if p.OrderDir != "asc" && p.OrderDir != "desc" {
				p.OrderDir = "desc"
			}
		}
	}

	return p
}

// Apply layers the search WHERE clause and ORDER BY onto a gorm query.
// Pagination (OFFSET/LIMIT) is applied separately by Paginate so callers
// can Count() on the filtered-but-unpaginated base.
func Apply(base *gorm.DB, p domain.ListParams) *gorm.DB {
	if p.Search != "" && len(p.SearchFields) > 0 {
		conds := make([]string, 0, len(p.SearchFields))
		args := make([]any, 0, len(p.SearchFields))
		like := "%" + p.Search + "%"
		for _, f := range p.SearchFields {
			conds = append(conds, fmt.Sprintf("%s ILIKE ?", f))
			args = append(args, like)
		}
		base = base.Where(strings.Join(conds, " OR "), args...)
	}
	if p.OrderBy != "" {
		dir := "DESC"
		if strings.EqualFold(p.OrderDir, "asc") {
			dir = "ASC"
		}
		base = base.Order(fmt.Sprintf("%s %s", p.OrderBy, dir))
	}
	return base
}

// Paginate runs COUNT on base and then the offset/limited SELECT into dst.
// base must already have Model and any feature-specific filters applied;
// search + ordering are added here via Apply so the COUNT reflects filters
// but not the offset/limit.
func Paginate[T any](base *gorm.DB, p domain.ListParams, dst *[]T) (int64, error) {
	filtered := Apply(base, p)

	var total int64
	if err := filtered.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return 0, err
	}
	if err := filtered.Session(&gorm.Session{}).Offset(p.Offset()).Limit(p.Limit()).Find(dst).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

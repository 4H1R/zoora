package domain

// DefaultPageSize is the fixed page size applied to all List APIs. Pinning
// this server-side (rather than reading from the client) prevents abuse of
// expensive queries via arbitrarily large page sizes.
const DefaultPageSize = 20

// ListParams carries pagination, search, and ordering for List APIs.
// Values are populated by the handler after white-listing against ListConfig;
// services and repositories treat ListParams as already-validated input.
type ListParams struct {
	Page         int
	PageSize     int
	Search       string
	SearchFields []string
	OrderBy      string
	OrderDir     string
}

// Offset returns the SQL OFFSET derived from Page/PageSize (1-based pages).
func (p ListParams) Offset() int {
	if p.Page <= 1 {
		return 0
	}
	return (p.Page - 1) * p.Limit()
}

// Limit returns the effective page size, falling back to DefaultPageSize.
func (p ListParams) Limit() int {
	if p.PageSize <= 0 {
		return DefaultPageSize
	}
	return p.PageSize
}

// ListConfig is the handler-owned white-list that gates what a client may
// search or order by on a given endpoint. Empty slices mean the corresponding
// feature is disabled and any client-supplied value is ignored.
type ListConfig struct {
	// AllowedSearchFields are the DB column names permitted in the OR search
	// clause. Empty means search is disabled.
	AllowedSearchFields []string
	// AllowedOrderFields are the DB column names permitted for ORDER BY.
	// Empty means ordering falls back to DefaultOrderBy unconditionally.
	AllowedOrderFields []string
	// DefaultOrderBy is applied when the client supplies no order_by, or
	// supplies one not in AllowedOrderFields. Empty means no ORDER BY.
	DefaultOrderBy string
	// DefaultOrderDir is "asc" or "desc"; defaults to "desc" when blank.
	DefaultOrderDir string
	// PageSize overrides DefaultPageSize for this endpoint. Zero → default.
	PageSize int
}

package listparams

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/domain"
)

func TestBindDefaultsAndSanitizesInvalidQueryValues(t *testing.T) {
	c := testContext("/items?page=0&page_size=-10&search=%20%20&order_by=DROP%20TABLE&order_dir=sideways")

	params := Bind(c, domain.ListConfig{
		AllowedSearchFields: []string{"name", "email"},
		AllowedOrderFields:  []string{"name", "created_at"},
		DefaultOrderBy:      "created_at",
		DefaultOrderDir:     "sideways",
		PageSize:            15,
	})

	if params.Page != 1 {
		t.Fatalf("Page = %d, want 1", params.Page)
	}
	if params.PageSize != 15 {
		t.Fatalf("PageSize = %d, want 15", params.PageSize)
	}
	if params.Search != "" || len(params.SearchFields) != 0 {
		t.Fatalf("blank search should be ignored, got %#v", params)
	}
	if params.OrderBy != "created_at" {
		t.Fatalf("OrderBy = %q, want default", params.OrderBy)
	}
	if params.OrderDir != "desc" {
		t.Fatalf("OrderDir = %q, want desc fallback", params.OrderDir)
	}
}

func TestBindAcceptsWhitelistedSearchAndOrdering(t *testing.T) {
	c := testContext("/items?page=3&page_size=50&search=%20alice%20&order_by=name&order_dir=ASC")

	params := Bind(c, domain.ListConfig{
		AllowedSearchFields: []string{"name", "email"},
		AllowedOrderFields:  []string{"name", "created_at"},
		DefaultOrderBy:      "created_at",
		DefaultOrderDir:     "desc",
	})

	if params.Page != 3 || params.PageSize != 50 {
		t.Fatalf("pagination = (%d,%d), want (3,50)", params.Page, params.PageSize)
	}
	if params.Search != "alice" {
		t.Fatalf("Search = %q, want trimmed alice", params.Search)
	}
	if got := len(params.SearchFields); got != 2 {
		t.Fatalf("SearchFields len = %d, want 2", got)
	}
	if params.OrderBy != "name" || params.OrderDir != "asc" {
		t.Fatalf("ordering = (%q,%q), want (name,asc)", params.OrderBy, params.OrderDir)
	}
}

func TestBindDisablesSearchWhenNoFieldsAllowed(t *testing.T) {
	c := testContext("/items?search=alice&order_dir=desc")

	params := Bind(c, domain.ListConfig{})

	if params.Search != "" {
		t.Fatalf("Search = %q, want empty when search is disabled", params.Search)
	}
	if params.PageSize != domain.DefaultPageSize {
		t.Fatalf("PageSize = %d, want default %d", params.PageSize, domain.DefaultPageSize)
	}
	if params.OrderBy != "" || params.OrderDir != "" {
		t.Fatalf("ordering should be empty when no default order exists, got (%q,%q)", params.OrderBy, params.OrderDir)
	}
}

func testContext(target string) *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, target, nil)
	return c
}

package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/cache"
)

// parseHostLabel strips the port and the base-domain suffix, returning the
// left-most subdomain label ("" for the apex itself).
func parseHostLabel(host, base string) string {
	if i := strings.IndexByte(host, ':'); i >= 0 {
		host = host[:i]
	}
	if host == base {
		return ""
	}
	suffix := "." + base
	if !strings.HasSuffix(host, suffix) {
		return ""
	}
	label := strings.TrimSuffix(host, suffix)
	// Only the first label matters (acme.foo.localhost -> acme).
	if i := strings.IndexByte(label, '.'); i >= 0 {
		label = label[:i]
	}
	return label
}

// Tenant resolves the request Host into a domain.HostContext and injects it.
// admin label -> HostClassAdmin; known active/suspended slug -> HostClassTenant;
// anything else -> HostClassUnknown (handlers decide how to respond).
func Tenant(rdb *redis.Client, orgRepo domain.OrganizationRepository, baseDomain, adminSub string) gin.HandlerFunc {
	return func(c *gin.Context) {
		label := parseHostLabel(c.Request.Host, baseDomain)
		hc := domain.HostContext{Slug: label}

		switch label {
		case adminSub:
			hc.Class = domain.HostClassAdmin
		case "":
			hc.Class = domain.HostClassUnknown
		default:
			if orgID, status, ok := resolveOrg(c, rdb, orgRepo, label); ok {
				hc.Class = domain.HostClassTenant
				hc.OrgID = &orgID
				hc.OrgStatus = status
			} else {
				hc.Class = domain.HostClassUnknown
			}
		}

		c.Request = c.Request.WithContext(domain.WithHostContext(c.Request.Context(), hc))
		c.Next()
	}
}

// OnDemandTLSCheck gates Caddy's on-demand TLS issuance. Caddy calls it with
// ?domain=<sni> before obtaining a cert; a 200 authorizes, anything else
// refuses. Only the admin label, the canonical www host, and labels backed by a
// real organization are allowed — otherwise any *.<base> hostname could exhaust
// the Let's Encrypt rate limit.
func OnDemandTLSCheck(rdb *redis.Client, orgRepo domain.OrganizationRepository, baseDomain, adminSub string) gin.HandlerFunc {
	return func(c *gin.Context) {
		host := c.Query("domain")
		if host == "" {
			c.Status(http.StatusBadRequest)
			return
		}
		switch label := parseHostLabel(host, baseDomain); label {
		case adminSub, "www":
			c.Status(http.StatusOK)
		case "":
			// Apex is an explicit Caddy site (managed cert), not on-demand; deny.
			c.Status(http.StatusForbidden)
		default:
			if _, _, ok := resolveOrg(c, rdb, orgRepo, label); ok {
				c.Status(http.StatusOK)
			} else {
				c.Status(http.StatusForbidden)
			}
		}
	}
}

func resolveOrg(c *gin.Context, rdb *redis.Client, orgRepo domain.OrganizationRepository, slug string) (orgID uuid.UUID, status domain.OrganizationStatus, ok bool) {
	ctx := c.Request.Context()
	if rdb != nil {
		if id, st, err := cache.GetTenant(ctx, rdb, slug); err == nil {
			return id, st, true
		}
	}
	org, err := orgRepo.FindBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return uuid.Nil, "", false
		}
		return uuid.Nil, "", false
	}
	if rdb != nil {
		_ = cache.SetTenant(ctx, rdb, slug, org.ID, org.Status)
	}
	return org.ID, org.Status, true
}

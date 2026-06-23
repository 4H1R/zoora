package domain

import (
	"context"
	"errors"
	"regexp"

	"github.com/google/uuid"
)

// HostClass is the tenancy scope a request's Host header resolves to.
type HostClass int

const (
	// HostClassUnknown means the Host did not resolve to admin or a known org.
	HostClassUnknown HostClass = iota
	// HostClassAdmin is the reserved admin subdomain (platform-wide scope).
	HostClassAdmin
	// HostClassTenant is a <slug>.<base> subdomain resolving to one org.
	HostClassTenant
)

// HostContext is the resolved tenancy of the current request, injected by the
// tenant middleware and read by login, the public /org resolver, and the auth
// middleware's org-match assertion.
type HostContext struct {
	Class     HostClass
	Slug      string
	OrgID     *uuid.UUID
	OrgStatus OrganizationStatus
}

type hostCtxKey struct{}

func WithHostContext(ctx context.Context, h HostContext) context.Context {
	return context.WithValue(ctx, hostCtxKey{}, h)
}

func HostContextFromCtx(ctx context.Context) (HostContext, bool) {
	h, ok := ctx.Value(hostCtxKey{}).(HostContext)
	return h, ok
}

// ReservedSlugs are subdomain labels the platform serves itself; orgs may not
// claim them or their tenant /api path would collide with infra hosts.
var ReservedSlugs = map[string]struct{}{
	"api": {}, "app": {}, "admin": {}, "www": {}, "s3": {},
	"livekit": {}, "mail": {}, "static": {}, "assets": {}, "cdn": {},
}

var slugPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

// ErrInvalidSlug is returned when a slug fails format or reserved-word checks.
var ErrInvalidSlug = errors.New("invalid organization slug")

// ValidateSlug enforces DNS-label format (lowercase alnum + internal dashes,
// 2–63 chars), and rejects reserved labels.
func ValidateSlug(slug string) error {
	if len(slug) < 2 || !slugPattern.MatchString(slug) {
		return ErrInvalidSlug
	}
	if _, reserved := ReservedSlugs[slug]; reserved {
		return ErrInvalidSlug
	}
	return nil
}

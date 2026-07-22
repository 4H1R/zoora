package calendar

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

// fakeRepo records the scope it was called with so we can assert resolution.
type fakeRepo struct {
	gotScope domain.ClassListScope
}

func (f *fakeRepo) ListEvents(_ context.Context, scope domain.ClassListScope, _ domain.CalendarRange) ([]domain.CalendarEvent, error) {
	f.gotScope = scope
	return nil, nil
}

func newCtx(c domain.Caller) context.Context {
	return domain.WithCaller(context.Background(), c)
}

func rng() domain.CalendarRange {
	return domain.CalendarRange{From: time.Now().Add(-time.Hour), To: time.Now().Add(time.Hour)}
}

func TestListEvents_NoCaller_Forbidden(t *testing.T) {
	svc := NewService(&fakeRepo{}, nil, nil)
	_, err := svc.ListEvents(context.Background(), rng())
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("want ErrForbidden, got %v", err)
	}
}

func TestResolveScope_Admin_All(t *testing.T) {
	fr := &fakeRepo{}
	svc := NewService(fr, nil, nil)
	_, err := svc.ListEvents(newCtx(domain.Caller{IsAdmin: true}), rng())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !fr.gotScope.All || fr.gotScope.OrganizationID != nil {
		t.Fatalf("admin should be All org-wide unbounded, got %+v", fr.gotScope)
	}
}

func TestResolveScope_ViewAny_OrgBounded(t *testing.T) {
	fr := &fakeRepo{}
	svc := NewService(fr, nil, nil)
	org := uuid.New()
	caller := domain.Caller{OrgID: &org, Permissions: []string{string(domain.PermClassesViewAny)}}
	_, _ = svc.ListEvents(newCtx(caller), rng())
	if !fr.gotScope.All || fr.gotScope.OrganizationID == nil || *fr.gotScope.OrganizationID != org {
		t.Fatalf("view_any should be All bounded to org, got %+v", fr.gotScope)
	}
}

func TestResolveScope_Plain_TeacherAndMember(t *testing.T) {
	fr := &fakeRepo{}
	svc := NewService(fr, nil, nil)
	uid := uuid.New()
	caller := domain.Caller{UserID: uid}
	_, _ = svc.ListEvents(newCtx(caller), rng())
	s := fr.gotScope
	if s.All || s.TeacherID == nil || s.MemberUserID == nil || *s.TeacherID != uid || *s.MemberUserID != uid {
		t.Fatalf("plain caller should scope to teacher+member of self, got %+v", s)
	}
}

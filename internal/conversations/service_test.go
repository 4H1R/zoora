package conversations_test

import (
	"context"
	"errors"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/conversations"
	"github.com/4H1R/zoora/internal/domain"
)

// ---- in-memory fake repo set ----
//
// A single fakeStore backs all four repo interfaces so cross-repo
// relationships (membership gating messages, reactions counting per message,
// unread counts derived from send-order) behave consistently, mirroring what
// the real Postgres-backed repos would enforce. Order is tracked explicitly
// (convOrder/msgOrder) rather than relying on uuid.New() being sortable,
// since only real uuidv7 ids (assigned by the DB) are time-ordered.

type fakeStore struct {
	mu sync.Mutex

	convs     map[uuid.UUID]domain.Conversation
	convOrder []uuid.UUID
	directIdx map[string]uuid.UUID // orgID|directKey -> conv id

	members map[uuid.UUID][]domain.ConversationMember // convID -> members

	msgs     map[uuid.UUID]domain.ConversationMessage
	msgOrder map[uuid.UUID][]uuid.UUID // convID -> message ids, oldest first

	reactions map[uuid.UUID][]domain.ConversationMessageReaction // msgID -> reactions
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		convs:     map[uuid.UUID]domain.Conversation{},
		directIdx: map[string]uuid.UUID{},
		members:   map[uuid.UUID][]domain.ConversationMember{},
		msgs:      map[uuid.UUID]domain.ConversationMessage{},
		msgOrder:  map[uuid.UUID][]uuid.UUID{},
		reactions: map[uuid.UUID][]domain.ConversationMessageReaction{},
	}
}

// ---- conversation repo ----

type fakeConvRepo struct{ s *fakeStore }

func (r fakeConvRepo) Create(_ context.Context, c *domain.Conversation) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	if c.DirectKey != nil {
		key := c.OrganizationID.String() + "|" + *c.DirectKey
		if _, exists := r.s.directIdx[key]; exists {
			return domain.ErrConflict
		}
		r.s.directIdx[key] = c.ID
	}
	now := time.Now()
	c.CreatedAt, c.UpdatedAt = now, now
	r.s.convs[c.ID] = *c
	r.s.convOrder = append(r.s.convOrder, c.ID)
	return nil
}

func (r fakeConvRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.Conversation, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	c, ok := r.s.convs[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &c, nil
}

func (r fakeConvRepo) Update(_ context.Context, c *domain.Conversation) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.convs[c.ID]; !ok {
		return domain.ErrNotFound
	}
	c.UpdatedAt = time.Now()
	r.s.convs[c.ID] = *c
	return nil
}

func (r fakeConvRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.convs[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.s.convs, id)
	delete(r.s.members, id)
	return nil
}

func (r fakeConvRepo) FindDirect(_ context.Context, orgID uuid.UUID, dk string) (*domain.Conversation, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	id, ok := r.s.directIdx[orgID.String()+"|"+dk]
	if !ok {
		return nil, domain.ErrNotFound
	}
	c := r.s.convs[id]
	return &c, nil
}

func (r fakeConvRepo) ListForUser(_ context.Context, orgID, userID uuid.UUID, q domain.ListConversationsQuery) ([]domain.Conversation, int64, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	var out []domain.Conversation
	for _, id := range r.s.convOrder {
		c := r.s.convs[id]
		if c.OrganizationID != orgID {
			continue
		}
		if q.Type != nil && c.Type != *q.Type {
			continue
		}
		isMember := false
		for _, m := range r.s.members[id] {
			if m.UserID == userID {
				isMember = true
				break
			}
		}
		if !isMember {
			continue
		}
		out = append(out, c)
	}
	return out, int64(len(out)), nil
}

func (r fakeConvRepo) Touch(_ context.Context, id uuid.UUID) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	c, ok := r.s.convs[id]
	if !ok {
		return domain.ErrNotFound
	}
	c.UpdatedAt = time.Now()
	r.s.convs[id] = c
	return nil
}

// ---- member repo ----

type fakeMemberRepo struct{ s *fakeStore }

func (r fakeMemberRepo) memberIdx(convID, userID uuid.UUID) int {
	for i, m := range r.s.members[convID] {
		if m.UserID == userID {
			return i
		}
	}
	return -1
}

func (r fakeMemberRepo) Create(_ context.Context, m *domain.ConversationMember) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if r.memberIdx(m.ConversationID, m.UserID) >= 0 {
		return domain.ErrConflict
	}
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	if m.JoinedAt.IsZero() {
		m.JoinedAt = time.Now()
	}
	r.s.members[m.ConversationID] = append(r.s.members[m.ConversationID], *m)
	return nil
}

func (r fakeMemberRepo) CreateMany(_ context.Context, members []domain.ConversationMember) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for _, m := range members {
		if r.memberIdx(m.ConversationID, m.UserID) >= 0 {
			return domain.ErrConflict
		}
	}
	for i := range members {
		m := members[i]
		if m.ID == uuid.Nil {
			m.ID = uuid.New()
		}
		if m.JoinedAt.IsZero() {
			m.JoinedAt = time.Now()
		}
		r.s.members[m.ConversationID] = append(r.s.members[m.ConversationID], m)
	}
	return nil
}

func (r fakeMemberRepo) FindByConversationAndUser(_ context.Context, convID, userID uuid.UUID) (*domain.ConversationMember, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	i := r.memberIdx(convID, userID)
	if i < 0 {
		return nil, domain.ErrNotFound
	}
	m := r.s.members[convID][i]
	return &m, nil
}

func (r fakeMemberRepo) Delete(_ context.Context, convID, userID uuid.UUID) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	i := r.memberIdx(convID, userID)
	if i < 0 {
		return domain.ErrNotFound
	}
	ms := r.s.members[convID]
	r.s.members[convID] = append(ms[:i], ms[i+1:]...)
	return nil
}

func (r fakeMemberRepo) ListByConversation(_ context.Context, convID uuid.UUID) ([]domain.ConversationMember, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	out := make([]domain.ConversationMember, len(r.s.members[convID]))
	copy(out, r.s.members[convID])
	return out, nil
}

func (r fakeMemberRepo) ListUserIDs(_ context.Context, convID uuid.UUID) ([]uuid.UUID, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	out := make([]uuid.UUID, 0, len(r.s.members[convID]))
	for _, m := range r.s.members[convID] {
		out = append(out, m.UserID)
	}
	return out, nil
}

func (r fakeMemberRepo) SetLastRead(_ context.Context, convID, userID, messageID uuid.UUID, at time.Time) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	i := r.memberIdx(convID, userID)
	if i < 0 {
		return domain.ErrNotFound
	}
	r.s.members[convID][i].LastReadMessageID = &messageID
	r.s.members[convID][i].LastReadAt = &at
	return nil
}

func (r fakeMemberRepo) SetMuted(_ context.Context, convID, userID uuid.UUID, until *time.Time) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	i := r.memberIdx(convID, userID)
	if i < 0 {
		return domain.ErrNotFound
	}
	r.s.members[convID][i].MutedUntil = until
	return nil
}

// UnreadCount counts messages with sender != userID that were created after
// the member's last-read position, using insertion order (msgOrder) as the
// "newer than" relation — the fake equivalent of the real repo's
// `id > last_read_message_id` keyset comparison on time-sortable uuidv7 ids.
func (r fakeMemberRepo) UnreadCount(_ context.Context, convID, userID uuid.UUID) (int64, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	order := r.s.msgOrder[convID]
	lastReadPos := -1
	if i := r.memberIdx(convID, userID); i >= 0 {
		if lr := r.s.members[convID][i].LastReadMessageID; lr != nil {
			for pos, id := range order {
				if id == *lr {
					lastReadPos = pos
					break
				}
			}
		}
	}
	var count int64
	for pos, id := range order {
		if pos <= lastReadPos {
			continue
		}
		msg := r.s.msgs[id]
		if msg.SenderID == nil || *msg.SenderID != userID {
			count++
		}
	}
	return count, nil
}

// ---- message repo ----

type fakeMessageRepo struct{ s *fakeStore }

func (r fakeMessageRepo) Create(_ context.Context, m *domain.ConversationMessage) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	if _, exists := r.s.msgs[m.ID]; exists {
		return domain.ErrConflict
	}
	now := time.Now()
	m.CreatedAt, m.UpdatedAt = now, now
	r.s.msgs[m.ID] = *m
	r.s.msgOrder[m.ConversationID] = append(r.s.msgOrder[m.ConversationID], m.ID)
	return nil
}

func (r fakeMessageRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.ConversationMessage, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	m, ok := r.s.msgs[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &m, nil
}

func (r fakeMessageRepo) Update(_ context.Context, m *domain.ConversationMessage) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.msgs[m.ID]; !ok {
		return domain.ErrNotFound
	}
	m.UpdatedAt = time.Now()
	r.s.msgs[m.ID] = *m
	return nil
}

func (r fakeMessageRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	m, ok := r.s.msgs[id]
	if !ok {
		return domain.ErrNotFound
	}
	delete(r.s.msgs, id)
	order := r.s.msgOrder[m.ConversationID]
	for i, mid := range order {
		if mid == id {
			r.s.msgOrder[m.ConversationID] = append(order[:i], order[i+1:]...)
			break
		}
	}
	return nil
}

func (r fakeMessageRepo) ListWindow(_ context.Context, convID uuid.UUID, cur domain.MessageCursor) ([]domain.ConversationMessage, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	order := r.s.msgOrder[convID]
	limit := cur.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	// newest-first
	out := make([]domain.ConversationMessage, 0, len(order))
	for i := len(order) - 1; i >= 0; i-- {
		out = append(out, r.s.msgs[order[i]])
	}
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (r fakeMessageRepo) Latest(_ context.Context, convID uuid.UUID) (*domain.ConversationMessage, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	order := r.s.msgOrder[convID]
	if len(order) == 0 {
		return nil, domain.ErrNotFound
	}
	m := r.s.msgs[order[len(order)-1]]
	return &m, nil
}

func (r fakeMessageRepo) ListPinned(_ context.Context, convID uuid.UUID) ([]domain.ConversationMessage, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	var out []domain.ConversationMessage
	for _, id := range r.s.msgOrder[convID] {
		m := r.s.msgs[id]
		if m.IsPinned {
			out = append(out, m)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		var ti, tj time.Time
		if out[i].PinnedAt != nil {
			ti = *out[i].PinnedAt
		}
		if out[j].PinnedAt != nil {
			tj = *out[j].PinnedAt
		}
		return ti.After(tj)
	})
	return out, nil
}

func (r fakeMessageRepo) SetPinned(_ context.Context, id uuid.UUID, pinned bool, by *uuid.UUID, at *time.Time) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	m, ok := r.s.msgs[id]
	if !ok {
		return domain.ErrNotFound
	}
	m.IsPinned = pinned
	m.PinnedBy = by
	m.PinnedAt = at
	r.s.msgs[id] = m
	return nil
}

func (r fakeMessageRepo) SearchInConversation(_ context.Context, convID uuid.UUID, q string, limit int) ([]domain.ConversationMessage, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	var out []domain.ConversationMessage
	for _, id := range r.s.msgOrder[convID] {
		m := r.s.msgs[id]
		if len(out) >= limit {
			break
		}
		out = append(out, m)
	}
	return out, nil
}

// SearchGlobal is a minimal fake: it ignores ranking (no FTS engine in the
// in-memory store) and just filters org-scoped, member-visible messages
// whose content contains q, newest-first, capped at limit. Good enough for
// the existing unit test suite, which does not exercise Search directly.
func (r fakeMessageRepo) SearchGlobal(_ context.Context, orgID, userID uuid.UUID, q string, limit int) ([]domain.ConversationMessage, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	var out []domain.ConversationMessage
	for _, convID := range r.s.convOrder {
		conv := r.s.convs[convID]
		if conv.OrganizationID != orgID {
			continue
		}
		isMember := false
		for _, m := range r.s.members[convID] {
			if m.UserID == userID {
				isMember = true
				break
			}
		}
		if !isMember {
			continue
		}
		order := r.s.msgOrder[convID]
		for i := len(order) - 1; i >= 0; i-- {
			m := r.s.msgs[order[i]]
			if q != "" && !strings.Contains(m.Content, q) {
				continue
			}
			out = append(out, m)
			if len(out) >= limit {
				return out, nil
			}
		}
	}
	return out, nil
}

// ---- reaction repo ----

type fakeReactionRepo struct{ s *fakeStore }

func (r fakeReactionRepo) Create(_ context.Context, x *domain.ConversationMessageReaction) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for _, existing := range r.s.reactions[x.MessageID] {
		if existing.UserID == x.UserID && existing.Emoji == x.Emoji {
			return domain.ErrConflict
		}
	}
	if x.ID == uuid.Nil {
		x.ID = uuid.New()
	}
	r.s.reactions[x.MessageID] = append(r.s.reactions[x.MessageID], *x)
	return nil
}

func (r fakeReactionRepo) Delete(_ context.Context, messageID, userID uuid.UUID, emoji string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	list := r.s.reactions[messageID]
	for i, x := range list {
		if x.UserID == userID && x.Emoji == emoji {
			r.s.reactions[messageID] = append(list[:i], list[i+1:]...)
			return nil
		}
	}
	return nil
}

func (r fakeReactionRepo) FindByMessageAndUser(_ context.Context, messageID, userID uuid.UUID, emoji string) (*domain.ConversationMessageReaction, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for _, x := range r.s.reactions[messageID] {
		if x.UserID == userID && x.Emoji == emoji {
			cp := x
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r fakeReactionRepo) CountByMessage(_ context.Context, messageID uuid.UUID) (map[string]int, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	out := map[string]int{}
	for _, x := range r.s.reactions[messageID] {
		out[x.Emoji]++
	}
	return out, nil
}

// ---- transactor ----

type noopTx struct{}

func (noopTx) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// ---- test harness ----

func newTestService(_ *testing.T) (domain.ConversationService, *fakeStore) {
	s := newFakeStore()
	svc := conversations.NewService(
		fakeConvRepo{s: s},
		fakeMemberRepo{s: s},
		fakeMessageRepo{s: s},
		fakeReactionRepo{s: s},
		nil, // mentionRepo: unused until Phase 3
		noopTx{},
		nil, // logger left nil to catch accidental use in Phase 1 methods; wire slog.Default() if a method needs logging
		nil, // broadcaster: wired in Phase 2
		nil, // notifier: wired in Phase 3
		nil, // userLookup: wired in Phase 3 — nil skips cross-org checks in unit tests
		nil, // mediaLookup: wired in Phase 3 — nil skips attachment validation in unit tests
	)
	return svc, s
}

// ---- Step 12: fakes for the mention/notify/userLookup/media ports ----
//
// These back the Phase 3 coverage below (mention persistence, notification
// fan-out incl. mute-gating, media validation). Kept separate from the four
// repo fakes above since only a handful of tests need them.

// mentionCall records one ConversationMentionRepository.CreateMany invocation.
type mentionCall struct {
	messageID uuid.UUID
	userIDs   []uuid.UUID
}

type fakeMentionRepo struct {
	mu    sync.Mutex
	calls []mentionCall
}

func (f *fakeMentionRepo) CreateMany(_ context.Context, messageID uuid.UUID, userIDs []uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	ids := append([]uuid.UUID(nil), userIDs...)
	f.calls = append(f.calls, mentionCall{messageID: messageID, userIDs: ids})
	return nil
}

// notifyCall records one notifier.NotifyMessage invocation (the service's
// port, not the concrete conversations.Notifier).
type notifyCall struct {
	convID     uuid.UUID
	recipients []uuid.UUID
}

// fakeNotifier stands in for the service's notifier port directly — it
// records recipients as resolved by the SERVICE (DM/channel = all-but-sender,
// group = mentioned-only) without exercising the real Notifier's mute
// filtering. Use conversations.NewNotifier + fakeNotificationService (below)
// for tests that need mute-gating.
type fakeNotifier struct {
	mu    sync.Mutex
	calls []notifyCall
}

func (f *fakeNotifier) NotifyMessage(_ context.Context, conv *domain.Conversation, _ *domain.ConversationMessage, recipientIDs []uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	ids := append([]uuid.UUID(nil), recipientIDs...)
	f.calls = append(f.calls, notifyCall{convID: conv.ID, recipients: ids})
	return nil
}

// fakeUserLookup backs the service's userLookup port for tests that need
// cross-org resolution to succeed for specific users (nil map entries look
// up as domain.ErrNotFound, mirroring "user does not exist").
type fakeUserLookup struct {
	orgs map[uuid.UUID]uuid.UUID // userID -> orgID
}

func (f *fakeUserLookup) OrgID(_ context.Context, userID uuid.UUID) (*uuid.UUID, error) {
	if org, ok := f.orgs[userID]; ok {
		o := org
		return &o, nil
	}
	return nil, domain.ErrNotFound
}

// fakeMediaLookup backs the service's mediaLookup port for the attachment
// validation tests.
type fakeMediaLookup struct {
	items map[uuid.UUID]domain.Media
}

func (f *fakeMediaLookup) FindByID(_ context.Context, id uuid.UUID) (*domain.Media, error) {
	if m, ok := f.items[id]; ok {
		cp := m
		return &cp, nil
	}
	return nil, domain.ErrNotFound
}

// fakeNotificationService implements domain.NotificationService by embedding
// a nil interface and overriding only SendSystem — the single method the
// real conversations.Notifier calls. Any other method would panic on the nil
// embed, but none is exercised by these tests. Used with the REAL
// conversations.Notifier (not fakeNotifier) so its mute-gating logic runs.
type fakeNotificationService struct {
	domain.NotificationService
	mu    sync.Mutex
	calls []domain.SystemNotificationInput
}

func (f *fakeNotificationService) SendSystem(_ context.Context, in domain.SystemNotificationInput) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, in)
	return nil
}

// newServiceWithMentionAndNotify wires a fake mentionRepo + fake notifier
// (the service port, mute-gating NOT exercised) for the mention-persistence
// and recipient-resolution tests.
func newServiceWithMentionAndNotify(_ *testing.T) (domain.ConversationService, *fakeStore, *fakeMentionRepo, *fakeNotifier) {
	s := newFakeStore()
	mentionRepo := &fakeMentionRepo{}
	notif := &fakeNotifier{}
	svc := conversations.NewService(
		fakeConvRepo{s: s},
		fakeMemberRepo{s: s},
		fakeMessageRepo{s: s},
		fakeReactionRepo{s: s},
		mentionRepo,
		noopTx{},
		slog.Default(),
		nil, // broadcaster
		notif,
		nil, // userLookup: nil skips cross-org checks
		nil, // mediaLookup
	)
	return svc, s, mentionRepo, notif
}

// newServiceWithRealNotifier wires the REAL conversations.Notifier (backed
// by a fake domain.NotificationService) so its muted-member filtering runs
// end-to-end through SendMessage's recipient resolution.
func newServiceWithRealNotifier(_ *testing.T) (domain.ConversationService, *fakeStore, *fakeNotificationService) {
	s := newFakeStore()
	notifSvc := &fakeNotificationService{}
	realNotifier := conversations.NewNotifier(notifSvc, fakeMemberRepo{s: s})
	svc := conversations.NewService(
		fakeConvRepo{s: s},
		fakeMemberRepo{s: s},
		fakeMessageRepo{s: s},
		fakeReactionRepo{s: s},
		nil, // mentionRepo: not needed for mute-gating coverage
		noopTx{},
		slog.Default(),
		nil, // broadcaster
		realNotifier,
		nil, // userLookup
		nil, // mediaLookup
	)
	return svc, s, notifSvc
}

// newServiceWithMedia wires a fake mediaLookup for the attachment-validation
// tests. The returned *fakeMediaLookup starts empty — tests populate it
// (typically after creating a conversation, so ModelID can reference the
// real conv.ID) before calling SendMessage.
func newServiceWithMedia(_ *testing.T) (domain.ConversationService, *fakeStore, *fakeMediaLookup) {
	s := newFakeStore()
	media := &fakeMediaLookup{items: map[uuid.UUID]domain.Media{}}
	svc := conversations.NewService(
		fakeConvRepo{s: s},
		fakeMemberRepo{s: s},
		fakeMessageRepo{s: s},
		fakeReactionRepo{s: s},
		nil, // mentionRepo
		noopTx{},
		slog.Default(),
		nil, // broadcaster
		nil, // notifier
		nil, // userLookup
		media,
	)
	return svc, s, media
}

func callerCtx(orgID, userID uuid.UUID, isAdmin bool, perms ...string) context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{
		UserID:      userID,
		OrgID:       &orgID,
		IsAdmin:     isAdmin,
		Permissions: perms,
		Ent:         domain.PlanCatalog[domain.PlanKey(domain.TierPro, 50)],
	})
}

// ---- Step 8: TDD anchor test ----

func TestCreateOrGetDirect_ReturnsSameConversation(t *testing.T) {
	orgA, userA, userB := uuid.New(), uuid.New(), uuid.New()
	svc, _ := newTestService(t)
	ctx := callerCtx(orgA, userA, false)

	first, err := svc.CreateOrGetDirect(ctx, domain.CreateDirectDTO{UserID: userB.String()})
	require.NoError(t, err)

	second, err := svc.CreateOrGetDirect(ctx, domain.CreateDirectDTO{UserID: userB.String()})
	require.NoError(t, err)

	assert.Equal(t, first.ID, second.ID)
}

// ---- Step 10: remaining coverage ----

func TestCreateOrGetDirect_SelfRejected(t *testing.T) {
	orgA, userA := uuid.New(), uuid.New()
	svc, _ := newTestService(t)
	ctx := callerCtx(orgA, userA, false)

	_, err := svc.CreateOrGetDirect(ctx, domain.CreateDirectDTO{UserID: userA.String()})
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestCreateGroupOrChannel_RequiresManagePermission(t *testing.T) {
	orgA, userA := uuid.New(), uuid.New()
	svc, _ := newTestService(t)
	ctx := callerCtx(orgA, userA, false) // no PermConversationsManage, not admin

	_, err := svc.CreateGroupOrChannel(ctx, domain.CreateConversationDTO{
		Type: domain.ConversationTypeGroup,
		Name: "Engineering",
	})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestCreateGroupOrChannel_AdminSuccess(t *testing.T) {
	orgA, userA := uuid.New(), uuid.New()
	svc, _ := newTestService(t)
	ctx := callerCtx(orgA, userA, true)

	conv, err := svc.CreateGroupOrChannel(ctx, domain.CreateConversationDTO{
		Type: domain.ConversationTypeChannel,
		Name: "Announcements",
	})
	require.NoError(t, err)
	assert.Equal(t, domain.ConversationTypeChannel, conv.Type)
}

func TestSendMessage_IdempotentClientID(t *testing.T) {
	orgA, userA, userB := uuid.New(), uuid.New(), uuid.New()
	svc, store := newTestService(t)
	ctx := callerCtx(orgA, userA, false)

	conv, err := svc.CreateOrGetDirect(ctx, domain.CreateDirectDTO{UserID: userB.String()})
	require.NoError(t, err)

	clientID := uuid.New().String()
	first, err := svc.SendMessage(ctx, conv.ID, domain.SendConversationMessageDTO{
		ID: &clientID, Content: "hello",
	})
	require.NoError(t, err)

	second, err := svc.SendMessage(ctx, conv.ID, domain.SendConversationMessageDTO{
		ID: &clientID, Content: "hello",
	})
	require.NoError(t, err)

	assert.Equal(t, first.ID, second.ID)
	assert.Len(t, store.msgOrder[conv.ID], 1, "idempotent resend must not create a second row")
}

func TestSendMessage_ChannelWriteGuard_NonAdminRejected(t *testing.T) {
	orgA, userA, userB := uuid.New(), uuid.New(), uuid.New()
	svc, _ := newTestService(t)
	adminCtx := callerCtx(orgA, userA, true)

	conv, err := svc.CreateGroupOrChannel(adminCtx, domain.CreateConversationDTO{
		Type:      domain.ConversationTypeChannel,
		Name:      "Announcements",
		MemberIDs: []string{userB.String()},
	})
	require.NoError(t, err)

	memberCtx := callerCtx(orgA, userB, false)
	_, err = svc.SendMessage(memberCtx, conv.ID, domain.SendConversationMessageDTO{Content: "hi all"})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestSendMessage_ChannelWriteGuard_AdminAllowed(t *testing.T) {
	orgA, userA := uuid.New(), uuid.New()
	svc, _ := newTestService(t)
	adminCtx := callerCtx(orgA, userA, true)

	conv, err := svc.CreateGroupOrChannel(adminCtx, domain.CreateConversationDTO{
		Type: domain.ConversationTypeChannel,
		Name: "Announcements",
	})
	require.NoError(t, err)

	_, err = svc.SendMessage(adminCtx, conv.ID, domain.SendConversationMessageDTO{Content: "hi all"})
	assert.NoError(t, err)
}

func TestSendMessage_NonMember_Forbidden(t *testing.T) {
	orgA, userA, userB, outsider := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	svc, _ := newTestService(t)
	ctx := callerCtx(orgA, userA, false)

	conv, err := svc.CreateOrGetDirect(ctx, domain.CreateDirectDTO{UserID: userB.String()})
	require.NoError(t, err)

	outsiderCtx := callerCtx(orgA, outsider, false)
	_, err = svc.SendMessage(outsiderCtx, conv.ID, domain.SendConversationMessageDTO{Content: "sneaky"})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestListForCaller_UnreadCountAfterSend(t *testing.T) {
	orgA, userA, userB := uuid.New(), uuid.New(), uuid.New()
	svc, _ := newTestService(t)
	ctxA := callerCtx(orgA, userA, false)
	ctxB := callerCtx(orgA, userB, false)

	conv, err := svc.CreateOrGetDirect(ctxA, domain.CreateDirectDTO{UserID: userB.String()})
	require.NoError(t, err)

	_, err = svc.SendMessage(ctxA, conv.ID, domain.SendConversationMessageDTO{Content: "hi B"})
	require.NoError(t, err)

	convsForB, _, err := svc.ListForCaller(ctxB, domain.ListConversationsQuery{})
	require.NoError(t, err)
	require.Len(t, convsForB, 1)
	assert.EqualValues(t, 1, convsForB[0].UnreadCount)
	require.NotNil(t, convsForB[0].LastMessage)
	assert.Equal(t, "hi B", convsForB[0].LastMessage.Content)

	// The sender's own message never counts as unread for the sender.
	convsForA, _, err := svc.ListForCaller(ctxA, domain.ListConversationsQuery{})
	require.NoError(t, err)
	require.Len(t, convsForA, 1)
	assert.EqualValues(t, 0, convsForA[0].UnreadCount)
}

func TestMarkRead_ClearsUnreadCount(t *testing.T) {
	orgA, userA, userB := uuid.New(), uuid.New(), uuid.New()
	svc, _ := newTestService(t)
	ctxA := callerCtx(orgA, userA, false)
	ctxB := callerCtx(orgA, userB, false)

	conv, err := svc.CreateOrGetDirect(ctxA, domain.CreateDirectDTO{UserID: userB.String()})
	require.NoError(t, err)

	msg, err := svc.SendMessage(ctxA, conv.ID, domain.SendConversationMessageDTO{Content: "hi B"})
	require.NoError(t, err)

	err = svc.MarkRead(ctxB, conv.ID, domain.MarkReadDTO{MessageID: msg.ID.String()})
	require.NoError(t, err)

	convsForB, _, err := svc.ListForCaller(ctxB, domain.ListConversationsQuery{})
	require.NoError(t, err)
	require.Len(t, convsForB, 1)
	assert.EqualValues(t, 0, convsForB[0].UnreadCount)
}

func TestToggleReaction_AddThenRemove(t *testing.T) {
	orgA, userA, userB := uuid.New(), uuid.New(), uuid.New()
	svc, _ := newTestService(t)
	ctxA := callerCtx(orgA, userA, false)
	ctxB := callerCtx(orgA, userB, false)

	conv, err := svc.CreateOrGetDirect(ctxA, domain.CreateDirectDTO{UserID: userB.String()})
	require.NoError(t, err)
	msg, err := svc.SendMessage(ctxA, conv.ID, domain.SendConversationMessageDTO{Content: "hi B"})
	require.NoError(t, err)

	got, err := svc.ToggleReaction(ctxB, msg.ID, domain.ToggleConversationReactionDTO{Emoji: "👍"})
	require.NoError(t, err)
	assert.Equal(t, 1, got.Reactions["👍"])

	got, err = svc.ToggleReaction(ctxB, msg.ID, domain.ToggleConversationReactionDTO{Emoji: "👍"})
	require.NoError(t, err)
	assert.Equal(t, 0, got.Reactions["👍"])
}

func TestGet_NonMember_Forbidden(t *testing.T) {
	orgA, userA, userB, outsider := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	svc, _ := newTestService(t)
	ctx := callerCtx(orgA, userA, false)

	conv, err := svc.CreateOrGetDirect(ctx, domain.CreateDirectDTO{UserID: userB.String()})
	require.NoError(t, err)

	outsiderCtx := callerCtx(orgA, outsider, false)
	_, err = svc.Get(outsiderCtx, conv.ID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestCaller_NoFeature_Forbidden(t *testing.T) {
	orgA, userA := uuid.New(), uuid.New()
	svc, _ := newTestService(t)
	ctx := domain.WithCaller(context.Background(), domain.Caller{
		UserID: userA, OrgID: &orgA,
		Ent: domain.PlanCatalog[domain.PlanKey(domain.TierFree, 50)],
	})

	_, err := svc.CreateOrGetDirect(ctx, domain.CreateDirectDTO{UserID: uuid.New().String()})
	assert.True(t, errors.Is(err, domain.ErrFeatureNotInPlan))
}

func TestEditMessage_OtherUser_Forbidden(t *testing.T) {
	orgA, userA, userB := uuid.New(), uuid.New(), uuid.New()
	svc, _ := newTestService(t)
	ctxA := callerCtx(orgA, userA, false)
	ctxB := callerCtx(orgA, userB, false)

	conv, err := svc.CreateOrGetDirect(ctxA, domain.CreateDirectDTO{UserID: userB.String()})
	require.NoError(t, err)
	msg, err := svc.SendMessage(ctxA, conv.ID, domain.SendConversationMessageDTO{Content: "hi B"})
	require.NoError(t, err)

	_, err = svc.EditMessage(ctxB, msg.ID, domain.UpdateConversationMessageDTO{Content: "hacked"})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

// managePerm is the org-wide conversations:manage permission string, granted to
// managers in the tests below.
var managePerm = string(domain.PermConversationsManage)

// newGroup creates a group in org owned by a same-org manager (non-platform-
// admin) who holds PermConversationsManage, returning the manager ctx + conv.
// The manager becomes a conversation-admin member (the creator).
func newGroup(t *testing.T, svc domain.ConversationService, org, manager uuid.UUID) *domain.Conversation {
	t.Helper()
	mgrCtx := callerCtx(org, manager, false, managePerm)
	conv, err := svc.CreateGroupOrChannel(mgrCtx, domain.CreateConversationDTO{
		Type: domain.ConversationTypeGroup,
		Name: "Engineering",
	})
	require.NoError(t, err)
	return conv
}

// FIX 2 regression guard: a PermConversationsManage holder from a DIFFERENT org
// must NOT be able to Update another org's conversation (FindByID is not
// org-scoped, so the service's org check is the only guard).
func TestUpdate_DifferentOrgManager_Forbidden(t *testing.T) {
	orgA, orgB, userA, managerB := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	svc, _ := newTestService(t)
	conv := newGroup(t, svc, orgA, userA)

	name := "hijacked"
	otherOrgCtx := callerCtx(orgB, managerB, false, managePerm)
	_, err := svc.Update(otherOrgCtx, conv.ID, domain.UpdateConversationDTO{Name: &name})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestUpdate_SameOrgManager_Allowed(t *testing.T) {
	orgA, userA, manager := uuid.New(), uuid.New(), uuid.New()
	svc, _ := newTestService(t)
	conv := newGroup(t, svc, orgA, userA)

	// A different same-org manager who is NOT a member may still Update.
	name := "renamed"
	mgrCtx := callerCtx(orgA, manager, false, managePerm)
	got, err := svc.Update(mgrCtx, conv.ID, domain.UpdateConversationDTO{Name: &name})
	require.NoError(t, err)
	assert.Equal(t, "renamed", got.Name)
}

func TestUpdate_ConvAdminMember_Allowed(t *testing.T) {
	orgA, userA, convAdmin := uuid.New(), uuid.New(), uuid.New()
	svc, _ := newTestService(t)
	conv := newGroup(t, svc, orgA, userA)

	// Add convAdmin as an admin-role member; they hold NO org permission.
	mgrCtx := callerCtx(orgA, userA, false, managePerm)
	_, err := svc.AddMember(mgrCtx, conv.ID, domain.AddConversationMemberDTO{
		UserID: convAdmin.String(), Role: domain.ConversationMemberRoleAdmin,
	})
	require.NoError(t, err)

	name := "by-conv-admin"
	adminCtx := callerCtx(orgA, convAdmin, false)
	got, err := svc.Update(adminCtx, conv.ID, domain.UpdateConversationDTO{Name: &name})
	require.NoError(t, err)
	assert.Equal(t, "by-conv-admin", got.Name)
}

func TestUpdate_PlainMember_Forbidden(t *testing.T) {
	orgA, userA, plain := uuid.New(), uuid.New(), uuid.New()
	svc, _ := newTestService(t)
	conv := newGroup(t, svc, orgA, userA)

	mgrCtx := callerCtx(orgA, userA, false, managePerm)
	_, err := svc.AddMember(mgrCtx, conv.ID, domain.AddConversationMemberDTO{
		UserID: plain.String(), Role: domain.ConversationMemberRoleMember,
	})
	require.NoError(t, err)

	name := "nope"
	plainCtx := callerCtx(orgA, plain, false)
	_, err = svc.Update(plainCtx, conv.ID, domain.UpdateConversationDTO{Name: &name})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestUpdate_PlatformAdminCrossOrg_Allowed(t *testing.T) {
	orgA, userA, orgB, platformAdmin := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	svc, _ := newTestService(t)
	conv := newGroup(t, svc, orgA, userA)

	// Platform admin (IsAdmin) belongs to a different org yet still crosses.
	name := "by-platform-admin"
	adminCtx := callerCtx(orgB, platformAdmin, true)
	got, err := svc.Update(adminCtx, conv.ID, domain.UpdateConversationDTO{Name: &name})
	require.NoError(t, err)
	assert.Equal(t, "by-platform-admin", got.Name)
}

// FIX 2 regression guard for the message path: a different-org manager must not
// delete another org's message.
func TestDeleteMessage_DifferentOrgManager_Forbidden(t *testing.T) {
	orgA, orgB, userA, managerB := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	svc, _ := newTestService(t)
	conv := newGroup(t, svc, orgA, userA)

	// The creator (userA, a member) sends a message in org A.
	authorCtx := callerCtx(orgA, userA, false, managePerm)
	msg, err := svc.SendMessage(authorCtx, conv.ID, domain.SendConversationMessageDTO{Content: "hi"})
	require.NoError(t, err)

	otherOrgCtx := callerCtx(orgB, managerB, false, managePerm)
	err = svc.DeleteMessage(otherOrgCtx, msg.ID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestDeleteMessage_SameOrgManager_Allowed(t *testing.T) {
	orgA, userA, manager := uuid.New(), uuid.New(), uuid.New()
	svc, store := newTestService(t)
	conv := newGroup(t, svc, orgA, userA)

	authorCtx := callerCtx(orgA, userA, false, managePerm)
	msg, err := svc.SendMessage(authorCtx, conv.ID, domain.SendConversationMessageDTO{Content: "hi"})
	require.NoError(t, err)

	mgrCtx := callerCtx(orgA, manager, false, managePerm)
	err = svc.DeleteMessage(mgrCtx, msg.ID)
	require.NoError(t, err)
	assert.Len(t, store.msgOrder[conv.ID], 0)
}

func TestAddMember_ConvAdmin_Works_PlainMember_Forbidden(t *testing.T) {
	orgA, userA, convAdmin, plain, newbie := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	svc, _ := newTestService(t)
	conv := newGroup(t, svc, orgA, userA)

	mgrCtx := callerCtx(orgA, userA, false, managePerm)
	_, err := svc.AddMember(mgrCtx, conv.ID, domain.AddConversationMemberDTO{
		UserID: convAdmin.String(), Role: domain.ConversationMemberRoleAdmin,
	})
	require.NoError(t, err)
	_, err = svc.AddMember(mgrCtx, conv.ID, domain.AddConversationMemberDTO{
		UserID: plain.String(), Role: domain.ConversationMemberRoleMember,
	})
	require.NoError(t, err)

	// conv-admin member (no org perm) can add.
	adminCtx := callerCtx(orgA, convAdmin, false)
	added, err := svc.AddMember(adminCtx, conv.ID, domain.AddConversationMemberDTO{UserID: newbie.String()})
	require.NoError(t, err)
	assert.Equal(t, newbie, added.UserID)

	// plain member cannot add.
	plainCtx := callerCtx(orgA, plain, false)
	_, err = svc.AddMember(plainCtx, conv.ID, domain.AddConversationMemberDTO{UserID: uuid.New().String()})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestRemoveMember_ConvAdmin_Works(t *testing.T) {
	orgA, userA, convAdmin, victim := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	svc, _ := newTestService(t)
	conv := newGroup(t, svc, orgA, userA)

	mgrCtx := callerCtx(orgA, userA, false, managePerm)
	_, err := svc.AddMember(mgrCtx, conv.ID, domain.AddConversationMemberDTO{
		UserID: convAdmin.String(), Role: domain.ConversationMemberRoleAdmin,
	})
	require.NoError(t, err)
	_, err = svc.AddMember(mgrCtx, conv.ID, domain.AddConversationMemberDTO{
		UserID: victim.String(), Role: domain.ConversationMemberRoleMember,
	})
	require.NoError(t, err)

	adminCtx := callerCtx(orgA, convAdmin, false)
	err = svc.RemoveMember(adminCtx, conv.ID, victim)
	require.NoError(t, err)

	members, err := svc.ListMembers(adminCtx, conv.ID)
	require.NoError(t, err)
	for _, m := range members {
		assert.NotEqual(t, victim, m.UserID)
	}
}

// ---- Step 12: mention persistence, notify fan-out, mute-gating, search, media ----

func TestSendMessage_MentionPersistence_FiltersNonMembers(t *testing.T) {
	orgA, userA, memberB, outsiderC := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	svc, _, mentionRepo, _ := newServiceWithMentionAndNotify(t)
	adminCtx := callerCtx(orgA, userA, true)

	conv, err := svc.CreateGroupOrChannel(adminCtx, domain.CreateConversationDTO{
		Type:      domain.ConversationTypeGroup,
		Name:      "Engineering",
		MemberIDs: []string{memberB.String()},
	})
	require.NoError(t, err)

	msg, err := svc.SendMessage(adminCtx, conv.ID, domain.SendConversationMessageDTO{
		Content:        "hey @b @outsider @me",
		MentionUserIDs: []string{memberB.String(), outsiderC.String(), userA.String()},
	})
	require.NoError(t, err)

	mentionRepo.mu.Lock()
	defer mentionRepo.mu.Unlock()
	require.Len(t, mentionRepo.calls, 1, "only one CreateMany call, for the member-only mention set")
	assert.Equal(t, msg.ID, mentionRepo.calls[0].messageID)
	assert.Equal(t, []uuid.UUID{memberB}, mentionRepo.calls[0].userIDs,
		"outsiderC (non-member) and userA (self) must be filtered out")
}

func TestSendMessage_GroupNotify_OnlyMentioned(t *testing.T) {
	orgA, userA, memberB, memberC := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	svc, _, _, notif := newServiceWithMentionAndNotify(t)
	adminCtx := callerCtx(orgA, userA, true)

	conv, err := svc.CreateGroupOrChannel(adminCtx, domain.CreateConversationDTO{
		Type:      domain.ConversationTypeGroup,
		Name:      "Engineering",
		MemberIDs: []string{memberB.String(), memberC.String()},
	})
	require.NoError(t, err)

	_, err = svc.SendMessage(adminCtx, conv.ID, domain.SendConversationMessageDTO{
		Content:        "hey @b",
		MentionUserIDs: []string{memberB.String()},
	})
	require.NoError(t, err)

	notif.mu.Lock()
	defer notif.mu.Unlock()
	require.Len(t, notif.calls, 1)
	assert.Equal(t, []uuid.UUID{memberB}, notif.calls[0].recipients,
		"group conversations notify only @mentioned members, not the whole roster")
}

func TestSendMessage_DirectNotify_TargetsOtherMember(t *testing.T) {
	orgA, userA, userB := uuid.New(), uuid.New(), uuid.New()
	svc, _, _, notif := newServiceWithMentionAndNotify(t)
	ctx := callerCtx(orgA, userA, false)

	conv, err := svc.CreateOrGetDirect(ctx, domain.CreateDirectDTO{UserID: userB.String()})
	require.NoError(t, err)

	_, err = svc.SendMessage(ctx, conv.ID, domain.SendConversationMessageDTO{Content: "hi B"})
	require.NoError(t, err)

	notif.mu.Lock()
	defer notif.mu.Unlock()
	require.Len(t, notif.calls, 1)
	assert.Equal(t, []uuid.UUID{userB}, notif.calls[0].recipients,
		"DMs notify the other member unconditionally, no mention required")
}

// TestSendMessage_MutedMember_SkippedByNotifier wires the REAL
// conversations.Notifier (not the fake service port) so its muted_until
// filtering actually runs, backed by a fake domain.NotificationService that
// records the resolved audience.
func TestSendMessage_MutedMember_SkippedByNotifier(t *testing.T) {
	orgA, userA, memberB, memberC := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	svc, _, notifSvc := newServiceWithRealNotifier(t)
	adminCtx := callerCtx(orgA, userA, true)

	conv, err := svc.CreateGroupOrChannel(adminCtx, domain.CreateConversationDTO{
		Type:      domain.ConversationTypeChannel,
		Name:      "Announcements",
		MemberIDs: []string{memberB.String(), memberC.String()},
	})
	require.NoError(t, err)

	// memberB mutes the channel for the next 24h.
	mutedCtx := callerCtx(orgA, memberB, false)
	future := time.Now().Add(24 * time.Hour)
	require.NoError(t, svc.SetMuted(mutedCtx, conv.ID, &future))

	_, err = svc.SendMessage(adminCtx, conv.ID, domain.SendConversationMessageDTO{Content: "hi all"})
	require.NoError(t, err)

	notifSvc.mu.Lock()
	defer notifSvc.mu.Unlock()
	require.Len(t, notifSvc.calls, 1)
	assert.ElementsMatch(t, []uuid.UUID{memberC}, notifSvc.calls[0].Audience.UserIDs,
		"muted memberB is excluded by the Notifier; sender userA is excluded by recipient resolution")
}

// TestCreateGroupOrChannel_UserLookup_SkipsCrossOrgMember exercises the
// fakeUserLookup port directly (wired non-nil, unlike the other harnesses
// above which skip the cross-org check via nil): a requested member that
// resolves to a different org must be silently skipped, while a same-org
// member is added normally.
func TestCreateGroupOrChannel_UserLookup_SkipsCrossOrgMember(t *testing.T) {
	s := newFakeStore()
	orgA, orgB := uuid.New(), uuid.New()
	userA, crossOrgUser, sameOrgUser := uuid.New(), uuid.New(), uuid.New()
	users := &fakeUserLookup{orgs: map[uuid.UUID]uuid.UUID{
		crossOrgUser: orgB,
		sameOrgUser:  orgA,
	}}
	svc := conversations.NewService(
		fakeConvRepo{s: s},
		fakeMemberRepo{s: s},
		fakeMessageRepo{s: s},
		fakeReactionRepo{s: s},
		nil, // mentionRepo
		noopTx{},
		slog.Default(),
		nil, // broadcaster
		nil, // notifier
		users,
		nil, // mediaLookup
	)
	ctx := callerCtx(orgA, userA, true)

	conv, err := svc.CreateGroupOrChannel(ctx, domain.CreateConversationDTO{
		Type:      domain.ConversationTypeGroup,
		Name:      "Engineering",
		MemberIDs: []string{crossOrgUser.String(), sameOrgUser.String()},
	})
	require.NoError(t, err)

	members, err := svc.ListMembers(ctx, conv.ID)
	require.NoError(t, err)
	var ids []uuid.UUID
	for _, m := range members {
		ids = append(ids, m.UserID)
	}
	assert.Contains(t, ids, sameOrgUser)
	assert.NotContains(t, ids, crossOrgUser, "a member id resolving to a different org must be silently skipped")
}

func TestSearch_RequiresMinLength(t *testing.T) {
	orgA, userA := uuid.New(), uuid.New()
	svc, _ := newTestService(t)
	ctx := callerCtx(orgA, userA, false)

	_, err := svc.Search(ctx, "ab", 10)
	assert.ErrorIs(t, err, domain.ErrValidation, "queries under 3 chars must be rejected")

	_, err = svc.Search(ctx, "abc", 10)
	assert.NoError(t, err, "a 3-char query is the minimum accepted length")
}

// TestSendMessage_MediaValidation_RejectsForeignAttachment covers Step 10:
// a media row that exists and belongs to the caller's org but is bound
// (model_id) to a DIFFERENT conversation must be rejected.
func TestSendMessage_MediaValidation_RejectsForeignAttachment(t *testing.T) {
	orgA, userA, userB := uuid.New(), uuid.New(), uuid.New()
	svc, _, media := newServiceWithMedia(t)
	ctx := callerCtx(orgA, userA, false)

	conv, err := svc.CreateOrGetDirect(ctx, domain.CreateDirectDTO{UserID: userB.String()})
	require.NoError(t, err)

	mediaID := uuid.New()
	media.items[mediaID] = domain.Media{
		ID:             mediaID,
		OrganizationID: &orgA,
		ModelType:      domain.MediaModelConversation,
		ModelID:        uuid.New(), // bound to a DIFFERENT conversation
	}

	_, err = svc.SendMessage(ctx, conv.ID, domain.SendConversationMessageDTO{
		Content:  "photo",
		MediaIDs: []string{mediaID.String()},
	})
	assert.ErrorIs(t, err, domain.ErrValidation)
}

// TestSendMessage_MediaValidation_AllowsBoundAttachment covers the happy
// path: a media row in the caller's org, bound to this conversation, is
// accepted and its id round-trips into the created message.
func TestSendMessage_MediaValidation_AllowsBoundAttachment(t *testing.T) {
	orgA, userA, userB := uuid.New(), uuid.New(), uuid.New()
	svc, _, media := newServiceWithMedia(t)
	ctx := callerCtx(orgA, userA, false)

	conv, err := svc.CreateOrGetDirect(ctx, domain.CreateDirectDTO{UserID: userB.String()})
	require.NoError(t, err)

	mediaID := uuid.New()
	media.items[mediaID] = domain.Media{
		ID:             mediaID,
		OrganizationID: &orgA,
		ModelType:      domain.MediaModelConversation,
		ModelID:        conv.ID,
	}

	msg, err := svc.SendMessage(ctx, conv.ID, domain.SendConversationMessageDTO{
		Content:  "photo",
		MediaIDs: []string{mediaID.String()},
	})
	require.NoError(t, err)
	assert.JSONEq(t, `["`+mediaID.String()+`"]`, string(msg.MediaIDs))
}

// TestListMessages_SerializesReactions covers Step 11: a page of messages
// returned by ListMessages must have Reactions populated per message, not
// just ToggleReaction's single-message return path (already covered above).
func TestListMessages_SerializesReactions(t *testing.T) {
	orgA, userA, userB := uuid.New(), uuid.New(), uuid.New()
	svc, _ := newTestService(t)
	ctxA := callerCtx(orgA, userA, false)
	ctxB := callerCtx(orgA, userB, false)

	conv, err := svc.CreateOrGetDirect(ctxA, domain.CreateDirectDTO{UserID: userB.String()})
	require.NoError(t, err)
	msg, err := svc.SendMessage(ctxA, conv.ID, domain.SendConversationMessageDTO{Content: "hi B"})
	require.NoError(t, err)

	_, err = svc.ToggleReaction(ctxB, msg.ID, domain.ToggleConversationReactionDTO{Emoji: "🔥"})
	require.NoError(t, err)

	msgs, err := svc.ListMessages(ctxA, conv.ID, domain.MessageCursor{})
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	assert.Equal(t, 1, msgs[0].Reactions["🔥"], "ListMessages must serialize reaction counts, not leave Reactions nil")
}

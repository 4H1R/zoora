package conversations

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/4H1R/zoora/internal/domain"
)

// broadcaster is the realtime port (Phase 2 supplies the WS+Redis impl; nil = no-op).
type broadcaster interface {
	ToConversation(ctx context.Context, convID uuid.UUID, eventType string, data any)
	ToUser(ctx context.Context, userID uuid.UUID, eventType string, data any)
	// ToUsers fans one event out to many per-user channels (impl batches, e.g.
	// a single Redis pipeline, instead of one round-trip per user).
	ToUsers(ctx context.Context, userIDs []uuid.UUID, eventType string, data any)
}

// notifier is the notifications port (Phase 3 supplies impl; nil = no-op).
type notifier interface {
	NotifyMessage(ctx context.Context, conv *domain.Conversation, msg *domain.ConversationMessage, recipientIDs []uuid.UUID) error
}

// userDirectory is the cross-org guard + directory port. REQUIRED (never
// nil): it backs multi-tenant security guards, so a missing impl must fail at
// construction, not silently skip the checks. Tests use a permissive stub.
type userDirectory interface {
	// FilterSameOrg returns the subset of ids belonging to users in orgID, in
	// one query. Unknown and cross-org ids are dropped.
	FilterSameOrg(ctx context.Context, orgID uuid.UUID, ids []uuid.UUID) ([]uuid.UUID, error)
	DirectorySearch(ctx context.Context, orgID uuid.UUID, query string, limit int) ([]domain.DirectoryUser, error)
	DirectoryByUsername(ctx context.Context, orgID uuid.UUID, username string) (*domain.DirectoryUser, error)
}

// attachmentValidator is the attachment-authz port. REQUIRED (never nil) for
// the same reason as userDirectory: it is the org/conversation binding check
// on message attachments. Tests use a permissive stub.
type attachmentValidator interface {
	// ValidateAttachments fails unless every media id exists, belongs to orgID,
	// and is already bound to convID (model_type=conversation, model_id=convID).
	ValidateAttachments(ctx context.Context, orgID, convID uuid.UUID, mediaIDs []string) error
}

// presenceReader is the online/last-seen read port (Phase 2 supplies an
// adapter over chathub.Presence; nil = presence disabled, returns empty).
type presenceReader interface {
	Get(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]domain.PresenceStatus, error)
}

// enqueuer is the async task port used to schedule attachment cleanup when a
// conversation is deleted (nil = skip enqueue, which is what unit tests do).
type enqueuer interface {
	Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}

type service struct {
	convRepo     domain.ConversationRepository
	memberRepo   domain.ConversationMemberRepository
	messageRepo  domain.ConversationMessageRepository
	reactionRepo domain.ConversationReactionRepository
	mentionRepo  domain.ConversationMentionRepository
	transactor   domain.Transactor
	logger       *slog.Logger
	rt           broadcaster         // may be nil (no realtime)
	notif        notifier            // may be nil (no notifications)
	users        userDirectory       // required — multi-tenant guard
	media        attachmentValidator // required — attachment authz
	presence     presenceReader      // may be nil (presence disabled)
	queue        enqueuer            // may be nil (no cleanup enqueue)
}

func NewService(
	convRepo domain.ConversationRepository,
	memberRepo domain.ConversationMemberRepository,
	messageRepo domain.ConversationMessageRepository,
	reactionRepo domain.ConversationReactionRepository,
	mentionRepo domain.ConversationMentionRepository,
	transactor domain.Transactor,
	logger *slog.Logger,
	rt broadcaster,
	notif notifier,
	users userDirectory,
	media attachmentValidator,
	presence presenceReader,
	queue enqueuer,
) domain.ConversationService {
	// Fail loudly at wiring time: a nil users/media port would silently disable
	// the cross-org and attachment authz guards, not just a feature.
	if users == nil || media == nil {
		panic("conversations.NewService: userDirectory and attachmentValidator are required (security guards)")
	}
	return &service{convRepo, memberRepo, messageRepo, reactionRepo, mentionRepo, transactor, logger, rt, notif, users, media, presence, queue}
}

// caller resolves the authenticated caller and enforces the org + feature
// gate every conversations operation requires.
func (s *service) caller(ctx context.Context) (domain.Caller, error) {
	c, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.Caller{}, domain.ErrForbidden
	}
	if c.OrgID == nil {
		return domain.Caller{}, domain.ErrForbidden
	}
	if !c.IsAdmin && !c.HasFeature(domain.FeatureChat) {
		return domain.Caller{}, domain.NewFeatureError(c.Ent.Plan, domain.FeatureChat)
	}
	return c, nil
}

// requireMember enforces hard conversation membership: even platform admins /
// PermConversationsManage holders cannot read a conversation's contents
// (DMs especially) unless they are actually a member.
func (s *service) requireMember(ctx context.Context, convID, userID uuid.UUID) (*domain.ConversationMember, error) {
	m, err := s.memberRepo.FindByConversationAndUser(ctx, convID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrForbidden
		}
		return nil, err
	}
	return m, nil
}

// memberOrNil looks up the caller's membership without turning "not a
// member" into an error — used by the manage-authz tier (Update/Delete/
// AddMember/RemoveMember/Pin/Unpin) where a platform admin or org-wide
// PermConversationsManage holder may act without being a member.
func (s *service) memberOrNil(ctx context.Context, convID, userID uuid.UUID) (*domain.ConversationMember, error) {
	m, err := s.memberRepo.FindByConversationAndUser(ctx, convID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return m, nil
}

// canManageConversation implements the manage-authz tier shared by
// Update/Delete/AddMember/RemoveMember/PinMessage/UnpinMessage. Only platform
// admins cross org boundaries; an org-wide PermConversationsManage holder must
// be in the SAME org as the conversation (conversationRepository.FindByID is
// not org-scoped, so this is the multi-tenant guard). A conversation-admin
// member is inherently same-org.
func (s *service) canManageConversation(caller domain.Caller, conv *domain.Conversation, member *domain.ConversationMember) bool {
	if caller.IsAdmin { // platform admin crosses orgs
		return true
	}
	if caller.HasPermission(domain.PermConversationsManage) &&
		caller.OrgID != nil && conv != nil && conv.OrganizationID == *caller.OrgID {
		return true
	}
	return member != nil && member.Role == domain.ConversationMemberRoleAdmin
}

// canManageMessage implements the sender-or-admin authz tier shared by
// EditMessage/DeleteMessage. Same org-scoping as canManageConversation: the
// PermConversationsManage path requires the message's conversation to be in
// the caller's org.
func (s *service) canManageMessage(caller domain.Caller, conv *domain.Conversation, msg *domain.ConversationMessage) bool {
	if caller.IsAdmin {
		return true
	}
	if msg.SenderID != nil && *msg.SenderID == caller.UserID { // sender
		return true
	}
	return caller.HasPermission(domain.PermConversationsManage) &&
		caller.OrgID != nil && conv != nil && conv.OrganizationID == *caller.OrgID
}

func (s *service) CreateOrGetDirect(ctx context.Context, dto domain.CreateDirectDTO) (*domain.Conversation, error) {
	caller, err := s.caller(ctx)
	if err != nil {
		return nil, err
	}
	other, err := uuid.Parse(dto.UserID)
	if err != nil {
		return nil, domain.NewValidationError(map[string]string{"user_id": "invalid uuid"})
	}
	if other == caller.UserID {
		return nil, domain.NewValidationError(map[string]string{"user_id": "cannot DM yourself"})
	}
	sameOrg, err := s.users.FilterSameOrg(ctx, *caller.OrgID, []uuid.UUID{other})
	if err != nil {
		return nil, err
	}
	if len(sameOrg) == 0 {
		return nil, domain.ErrForbidden
	}

	dk := directKey(caller.UserID, other)
	if existing, err := s.convRepo.FindDirect(ctx, *caller.OrgID, dk); err == nil {
		return s.withMembers(ctx, existing), nil
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	var conv *domain.Conversation
	err = s.transactor.RunInTx(ctx, func(txCtx context.Context) error {
		conv = &domain.Conversation{
			OrganizationID: *caller.OrgID,
			Type:           domain.ConversationTypeDirect,
			CreatedBy:      &caller.UserID,
			DirectKey:      &dk,
		}
		if cerr := s.convRepo.Create(txCtx, conv); cerr != nil {
			// Propagate the conflict out — do NOT re-query on the aborted tx.
			// Postgres marks a tx failed after the first erroring statement, so
			// a same-tx FindDirect would fail with "current transaction is
			// aborted". The race-loser re-fetches on the OUTER ctx below.
			return cerr
		}
		now := time.Now()
		return s.memberRepo.CreateMany(txCtx, []domain.ConversationMember{
			{ConversationID: conv.ID, UserID: caller.UserID, Role: domain.ConversationMemberRoleMember, JoinedAt: now},
			{ConversationID: conv.ID, UserID: other, Role: domain.ConversationMemberRoleMember, JoinedAt: now},
		})
	})
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			// race: another request created the DM first — re-fetch on a fresh
			// query using the outer ctx (the tx above is rolled back).
			winner, ferr := s.convRepo.FindDirect(ctx, *caller.OrgID, dk)
			if ferr != nil {
				return nil, ferr
			}
			return s.withMembers(ctx, winner), nil
		}
		return nil, err
	}
	conv = s.withMembers(ctx, conv)
	if s.rt != nil {
		// The other member hasn't joined this brand-new DM's WS room, so their
		// sidebar would miss it until a manual refresh. Nudge their per-user
		// channel directly (mirrors AddMember) so the list surfaces the new
		// conversation in real time — the first message's conversation_bump
		// alone can't, since the row isn't loaded yet.
		s.rt.ToUser(ctx, other, "conversation_updated", conv)
	}
	return conv, nil
}

// withMembers decorates a conversation with its (User-preloaded) member rows —
// the DM pair is what lets the client title a direct conversation and resolve
// presence without an extra roster round-trip. Best-effort: a failed roster
// read returns the bare conversation.
func (s *service) withMembers(ctx context.Context, conv *domain.Conversation) *domain.Conversation {
	if conv == nil {
		return conv
	}
	if members, err := s.memberRepo.ListByConversation(ctx, conv.ID); err == nil {
		conv.Members = members
	}
	return conv
}

const directorySearchLimit = 20

func (s *service) SearchDirectory(ctx context.Context, query string, limit int) ([]domain.DirectoryUser, error) {
	caller, err := s.caller(ctx)
	if err != nil {
		return nil, err
	}
	if caller.OrgID == nil {
		return nil, domain.ErrForbidden
	}
	if limit <= 0 || limit > directorySearchLimit {
		limit = directorySearchLimit
	}
	// Over-fetch by one so dropping the caller still yields up to `limit` rows.
	rows, err := s.users.DirectorySearch(ctx, *caller.OrgID, query, limit+1)
	if err != nil {
		return nil, err
	}
	out := make([]domain.DirectoryUser, 0, len(rows))
	for _, u := range rows {
		if u.ID == caller.UserID {
			continue
		}
		out = append(out, u)
		if len(out) == limit {
			break
		}
	}
	return out, nil
}

func (s *service) GetDirectoryUser(ctx context.Context, username string) (*domain.DirectoryUser, error) {
	caller, err := s.caller(ctx)
	if err != nil {
		return nil, err
	}
	if caller.OrgID == nil {
		return nil, domain.ErrForbidden
	}
	return s.users.DirectoryByUsername(ctx, *caller.OrgID, username)
}

func (s *service) CreateGroupOrChannel(ctx context.Context, dto domain.CreateConversationDTO) (*domain.Conversation, error) {
	caller, err := s.caller(ctx)
	if err != nil {
		return nil, err
	}
	if !caller.IsAdmin && !caller.HasPermission(domain.PermConversationsManage) {
		return nil, domain.ErrForbidden
	}
	if dto.Type != domain.ConversationTypeGroup && dto.Type != domain.ConversationTypeChannel {
		return nil, domain.NewValidationError(map[string]string{"type": "must be group or channel"})
	}
	if dto.Name == "" {
		return nil, domain.NewValidationError(map[string]string{"name": "required for group/channel"})
	}

	// Parse + dedup the requested member ids, then org-filter them in ONE
	// query before the tx (cross-org / unresolvable / malformed ids are
	// silently skipped, as before).
	requested := make([]uuid.UUID, 0, len(dto.MemberIDs))
	seen := map[uuid.UUID]bool{caller.UserID: true}
	for _, idStr := range dto.MemberIDs {
		uid, perr := uuid.Parse(idStr)
		if perr != nil || seen[uid] {
			continue
		}
		seen[uid] = true
		requested = append(requested, uid)
	}
	allowed, err := s.users.FilterSameOrg(ctx, *caller.OrgID, requested)
	if err != nil {
		return nil, err
	}

	var conv *domain.Conversation
	err = s.transactor.RunInTx(ctx, func(txCtx context.Context) error {
		conv = &domain.Conversation{
			OrganizationID: *caller.OrgID,
			Type:           dto.Type,
			Name:           dto.Name,
			ColorIndex:     dto.ColorIndex,
			CreatedBy:      &caller.UserID,
		}
		if cerr := s.convRepo.Create(txCtx, conv); cerr != nil {
			return cerr
		}
		now := time.Now()
		members := make([]domain.ConversationMember, 0, len(allowed)+1)
		members = append(members, domain.ConversationMember{
			ConversationID: conv.ID, UserID: caller.UserID, Role: domain.ConversationMemberRoleAdmin, JoinedAt: now,
		})
		for _, uid := range allowed {
			members = append(members, domain.ConversationMember{
				ConversationID: conv.ID, UserID: uid, Role: domain.ConversationMemberRoleMember, JoinedAt: now,
			})
		}
		return s.memberRepo.CreateMany(txCtx, members)
	})
	if err != nil {
		return nil, err
	}
	return conv, nil
}

// CreateForClass creates a group/channel seeded with the given members WITHOUT a
// conversations:manage check — the caller was already authorized at the class
// level (class ownership). CreatorID becomes admin; MemberIDs become members.
// Member ids are still org-filtered and deduped against the creator, so a stray
// roster id can never leak a cross-org user into the conversation.
func (s *service) CreateForClass(ctx context.Context, in domain.ProvisionClassChatDTO) (*domain.Conversation, error) {
	if in.Type != domain.ConversationTypeGroup && in.Type != domain.ConversationTypeChannel {
		return nil, domain.NewValidationError(map[string]string{"type": "must be group or channel"})
	}
	if in.Name == "" {
		return nil, domain.NewValidationError(map[string]string{"name": "required for group/channel"})
	}

	requested := make([]uuid.UUID, 0, len(in.MemberIDs))
	seen := map[uuid.UUID]bool{in.CreatorID: true}
	for _, uid := range in.MemberIDs {
		if seen[uid] {
			continue
		}
		seen[uid] = true
		requested = append(requested, uid)
	}
	allowed, err := s.users.FilterSameOrg(ctx, in.OrganizationID, requested)
	if err != nil {
		return nil, err
	}

	var conv *domain.Conversation
	err = s.transactor.RunInTx(ctx, func(txCtx context.Context) error {
		conv = &domain.Conversation{
			OrganizationID: in.OrganizationID,
			Type:           in.Type,
			Name:           in.Name,
			ColorIndex:     in.ColorIndex,
			CreatedBy:      &in.CreatorID,
		}
		if cerr := s.convRepo.Create(txCtx, conv); cerr != nil {
			return cerr
		}
		now := time.Now()
		members := make([]domain.ConversationMember, 0, len(allowed)+1)
		members = append(members, domain.ConversationMember{
			ConversationID: conv.ID, UserID: in.CreatorID, Role: domain.ConversationMemberRoleAdmin, JoinedAt: now,
		})
		for _, uid := range allowed {
			members = append(members, domain.ConversationMember{
				ConversationID: conv.ID, UserID: uid, Role: domain.ConversationMemberRoleMember, JoinedAt: now,
			})
		}
		return s.memberRepo.CreateMany(txCtx, members)
	})
	if err != nil {
		return nil, err
	}
	return conv, nil
}

// SyncClassMembers adds every memberID that is not already in the conversation
// (additive, idempotent) and returns the refreshed conversation. New members are
// org-filtered against the conversation's org before insert.
func (s *service) SyncClassMembers(ctx context.Context, convID uuid.UUID, memberIDs []uuid.UUID) (*domain.Conversation, error) {
	conv, err := s.convRepo.FindByID(ctx, convID)
	if err != nil {
		return nil, err
	}
	existing, err := s.memberRepo.ListUserIDs(ctx, convID)
	if err != nil {
		return nil, err
	}
	have := make(map[uuid.UUID]bool, len(existing))
	for _, id := range existing {
		have[id] = true
	}
	missing := make([]uuid.UUID, 0, len(memberIDs))
	seen := map[uuid.UUID]bool{}
	for _, id := range memberIDs {
		if have[id] || seen[id] {
			continue
		}
		seen[id] = true
		missing = append(missing, id)
	}
	if len(missing) == 0 {
		return conv, nil
	}
	allowed, err := s.users.FilterSameOrg(ctx, conv.OrganizationID, missing)
	if err != nil {
		return nil, err
	}
	if len(allowed) == 0 {
		return conv, nil
	}
	now := time.Now()
	rows := make([]domain.ConversationMember, 0, len(allowed))
	for _, uid := range allowed {
		rows = append(rows, domain.ConversationMember{
			ConversationID: convID, UserID: uid, Role: domain.ConversationMemberRoleMember, JoinedAt: now,
		})
	}
	if err := s.memberRepo.CreateMany(ctx, rows); err != nil {
		return nil, err
	}
	return conv, nil
}

// convForManage is the shared preamble of the manage-authz tier
// (Update/Delete/AddMember/RemoveMember): resolve caller, load the
// conversation, and enforce canManageConversation.
func (s *service) convForManage(ctx context.Context, convID uuid.UUID) (domain.Caller, *domain.Conversation, error) {
	caller, err := s.caller(ctx)
	if err != nil {
		return domain.Caller{}, nil, err
	}
	conv, err := s.convRepo.FindByID(ctx, convID)
	if err != nil {
		return domain.Caller{}, nil, err
	}
	member, err := s.memberOrNil(ctx, convID, caller.UserID)
	if err != nil {
		return domain.Caller{}, nil, err
	}
	if !s.canManageConversation(caller, conv, member) {
		return domain.Caller{}, nil, domain.ErrForbidden
	}
	return caller, conv, nil
}

// msgConvForManage is convForManage keyed by message id (Pin/Unpin).
func (s *service) msgConvForManage(ctx context.Context, msgID uuid.UUID) (domain.Caller, *domain.ConversationMessage, error) {
	caller, err := s.caller(ctx)
	if err != nil {
		return domain.Caller{}, nil, err
	}
	msg, err := s.messageRepo.FindByID(ctx, msgID)
	if err != nil {
		return domain.Caller{}, nil, err
	}
	conv, err := s.convRepo.FindByID(ctx, msg.ConversationID)
	if err != nil {
		return domain.Caller{}, nil, err
	}
	member, err := s.memberOrNil(ctx, msg.ConversationID, caller.UserID)
	if err != nil {
		return domain.Caller{}, nil, err
	}
	if !s.canManageConversation(caller, conv, member) {
		return domain.Caller{}, nil, domain.ErrForbidden
	}
	return caller, msg, nil
}

// msgForSenderOrManage is the sender-or-admin preamble (EditMessage/
// DeleteMessage): resolve caller, load message + conversation, and enforce
// canManageMessage.
func (s *service) msgForSenderOrManage(ctx context.Context, msgID uuid.UUID) (domain.Caller, *domain.Conversation, *domain.ConversationMessage, error) {
	caller, err := s.caller(ctx)
	if err != nil {
		return domain.Caller{}, nil, nil, err
	}
	msg, err := s.messageRepo.FindByID(ctx, msgID)
	if err != nil {
		return domain.Caller{}, nil, nil, err
	}
	conv, err := s.convRepo.FindByID(ctx, msg.ConversationID)
	if err != nil {
		return domain.Caller{}, nil, nil, err
	}
	if !s.canManageMessage(caller, conv, msg) {
		return domain.Caller{}, nil, nil, domain.ErrForbidden
	}
	return caller, conv, msg, nil
}

func (s *service) Get(ctx context.Context, id uuid.UUID) (*domain.Conversation, error) {
	caller, err := s.caller(ctx)
	if err != nil {
		return nil, err
	}
	conv, err := s.convRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireMember(ctx, id, caller.UserID); err != nil {
		return nil, err
	}
	if conv.Type == domain.ConversationTypeDirect {
		conv = s.withMembers(ctx, conv)
	}
	return conv, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, dto domain.UpdateConversationDTO) (*domain.Conversation, error) {
	_, conv, err := s.convForManage(ctx, id)
	if err != nil {
		return nil, err
	}
	if dto.Name != nil {
		conv.Name = *dto.Name
	}
	if dto.AvatarURL != nil {
		conv.AvatarURL = *dto.AvatarURL
	}
	if dto.ColorIndex != nil {
		conv.ColorIndex = *dto.ColorIndex
	}
	if err := s.convRepo.Update(ctx, conv); err != nil {
		return nil, err
	}
	if s.rt != nil {
		s.rt.ToConversation(ctx, conv.ID, "conversation_updated", conv)
	}
	return conv, nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	caller, _, err := s.convForManage(ctx, id)
	if err != nil {
		return err
	}
	// Capture the roster BEFORE the delete cascades the member rows away — we
	// need it to notify each member's per-user channel that the conversation is
	// gone. Best-effort: a failed read just skips the realtime nudge.
	members, merr := s.memberRepo.ListByConversation(ctx, id)
	if err := s.convRepo.Delete(ctx, id); err != nil {
		return err
	}
	if s.rt != nil && merr == nil {
		// Members haven't joined the (now-deleted) WS room, so nudge each
		// per-user channel directly: clients drop the thread + list row, and a
		// member viewing the conversation is redirected back to the list.
		// deleted_by lets a client suppress the "was deleted" toast for the
		// actor themselves (who already got a delete-success toast).
		for _, m := range members {
			s.rt.ToUser(ctx, m.UserID, "conversation_deleted", map[string]any{
				"conversation_id": id.String(),
				"deleted_by":      caller.UserID.String(),
			})
		}
	}
	s.enqueueAttachmentCleanup(ctx, id)
	return nil
}

// enqueueAttachmentCleanup schedules deletion of a deleted conversation's chat
// attachments (media rows + S3 objects). The DB cascade drops messages, but the
// polymorphic media table has no FK to conversations, so its rows and S3
// objects would otherwise orphan. Best-effort: a nil queue or enqueue failure
// is logged and never blocks the delete.
func (s *service) enqueueAttachmentCleanup(ctx context.Context, convID uuid.UUID) {
	if s.queue == nil {
		return
	}
	payload, err := json.Marshal(domain.MediaCleanupPayload{
		ModelType:      domain.MediaModelConversation,
		ModelID:        convID,
		CollectionName: domain.MediaCollectionAttach,
	})
	if err != nil {
		s.logger.Error("attachment cleanup enqueue: marshal payload", "conversation_id", convID.String(), "error", err)
		return
	}
	if _, err := s.queue.Enqueue(asynq.NewTask(domain.TypeMediaCleanup, payload), asynq.Queue(domain.QueueMedia)); err != nil {
		s.logger.Error("attachment cleanup enqueue", "conversation_id", convID.String(), "error", err)
	}
}

// ListForCaller populates the computed UnreadCount + LastMessage fields for
// the whole page with two grouped queries (unread counts, latest messages).
func (s *service) ListForCaller(ctx context.Context, q domain.ListConversationsQuery) ([]domain.Conversation, int64, error) {
	caller, err := s.caller(ctx)
	if err != nil {
		return nil, 0, err
	}
	convs, total, err := s.convRepo.ListForUser(ctx, *caller.OrgID, caller.UserID, q)
	if err != nil {
		return nil, 0, err
	}
	if len(convs) == 0 {
		return convs, total, nil
	}
	ids := make([]uuid.UUID, len(convs))
	for i := range convs {
		ids[i] = convs[i].ID
	}
	unread, err := s.memberRepo.UnreadCounts(ctx, caller.UserID, ids)
	if err != nil {
		return nil, 0, err
	}
	latest, err := s.messageRepo.LatestByConversation(ctx, ids)
	if err != nil {
		return nil, 0, err
	}
	directIDs := make([]uuid.UUID, 0, len(convs))
	for i := range convs {
		if convs[i].Type == domain.ConversationTypeDirect {
			directIDs = append(directIDs, convs[i].ID)
		}
	}
	// Decorate with the member rows the client needs: the full DM pair (title/
	// presence) and the viewer's own row everywhere (muted state).
	members, err := s.memberRepo.ListPageMembers(ctx, ids, directIDs, caller.UserID)
	if err != nil {
		return nil, 0, err
	}
	byConv := make(map[uuid.UUID][]domain.ConversationMember, len(convs))
	for _, m := range members {
		byConv[m.ConversationID] = append(byConv[m.ConversationID], m)
	}
	for i := range convs {
		convs[i].UnreadCount = unread[convs[i].ID]
		convs[i].Members = byConv[convs[i].ID]
		if last, ok := latest[convs[i].ID]; ok {
			m := last
			convs[i].LastMessage = &m
		}
	}
	return convs, total, nil
}

func (s *service) AddMember(ctx context.Context, convID uuid.UUID, dto domain.AddConversationMemberDTO) (*domain.ConversationMember, error) {
	caller, conv, err := s.convForManage(ctx, convID)
	if err != nil {
		return nil, err
	}
	// A DM's roster is fixed by its direct_key: growing it would silently turn
	// it into a group while colliding with the one-DM-per-pair invariant.
	if conv.Type == domain.ConversationTypeDirect {
		return nil, domain.NewValidationError(map[string]string{"conversation": "cannot add members to a direct conversation"})
	}
	uid, perr := uuid.Parse(dto.UserID)
	if perr != nil {
		return nil, domain.NewValidationError(map[string]string{"user_id": "invalid uuid"})
	}
	sameOrg, err := s.users.FilterSameOrg(ctx, *caller.OrgID, []uuid.UUID{uid})
	if err != nil {
		return nil, err
	}
	if len(sameOrg) == 0 {
		return nil, domain.ErrForbidden
	}
	role := dto.Role
	if role == "" {
		role = domain.ConversationMemberRoleMember
	}
	if !role.Valid() {
		return nil, domain.NewValidationError(map[string]string{"role": "invalid role"})
	}
	m := &domain.ConversationMember{ConversationID: convID, UserID: uid, Role: role, JoinedAt: time.Now()}
	if err := s.memberRepo.Create(ctx, m); err != nil {
		return nil, err
	}
	if s.rt != nil {
		s.rt.ToConversation(ctx, convID, "member_added", map[string]any{
			"conversation_id": convID.String(),
			"user_id":         uid.String(),
		})
		// The added user hasn't joined the WS room yet, so nudge their sidebar
		// directly via their user channel.
		s.rt.ToUser(ctx, uid, "conversation_updated", conv)
	}
	return m, nil
}

func (s *service) RemoveMember(ctx context.Context, convID, userID uuid.UUID) error {
	_, conv, err := s.convForManage(ctx, convID)
	if err != nil {
		return err
	}
	// Removing either side of a DM strands it: the direct_key still exists, so
	// the pair could never DM again (CreateOrGetDirect returns the stranded
	// conversation). Delete the conversation instead.
	if conv.Type == domain.ConversationTypeDirect {
		return domain.NewValidationError(map[string]string{"conversation": "cannot remove members from a direct conversation"})
	}
	if err := s.memberRepo.Delete(ctx, convID, userID); err != nil {
		return err
	}
	if s.rt != nil {
		s.rt.ToConversation(ctx, convID, "member_removed", map[string]any{
			"conversation_id": convID.String(),
			"user_id":         userID.String(),
		})
	}
	return nil
}

func (s *service) ListMembers(ctx context.Context, convID uuid.UUID) ([]domain.ConversationMember, error) {
	caller, err := s.caller(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireMember(ctx, convID, caller.UserID); err != nil {
		return nil, err
	}
	return s.memberRepo.ListByConversation(ctx, convID)
}

func (s *service) Leave(ctx context.Context, convID uuid.UUID) error {
	caller, err := s.caller(ctx)
	if err != nil {
		return err
	}
	if _, err := s.requireMember(ctx, convID, caller.UserID); err != nil {
		return err
	}
	conv, err := s.convRepo.FindByID(ctx, convID)
	if err != nil {
		return err
	}
	// Leaving a DM would strand it forever (see RemoveMember): the surviving
	// direct_key blocks re-creation while the leaver is no longer a member.
	if conv.Type == domain.ConversationTypeDirect {
		return domain.NewValidationError(map[string]string{"conversation": "cannot leave a direct conversation"})
	}
	return s.memberRepo.Delete(ctx, convID, caller.UserID)
}

func (s *service) SendMessage(ctx context.Context, convID uuid.UUID, dto domain.SendConversationMessageDTO) (*domain.ConversationMessage, error) {
	caller, err := s.caller(ctx)
	if err != nil {
		return nil, err
	}
	conv, err := s.convRepo.FindByID(ctx, convID)
	if err != nil {
		return nil, err
	}
	member, err := s.requireMember(ctx, convID, caller.UserID)
	if err != nil {
		return nil, err
	}
	// Channel: only admins (member role admin) or platform admins may post.
	if conv.Type == domain.ConversationTypeChannel &&
		member.Role != domain.ConversationMemberRoleAdmin &&
		!caller.IsAdmin && !caller.HasPermission(domain.PermConversationsManage) {
		return nil, domain.ErrForbidden
	}

	// Idempotency: client-supplied id → pre-check, scoped to this conversation
	// + sender (see ownMessageInConv — an unscoped lookup would let any member
	// "send" with a foreign message id and read messages from other
	// conversations).
	if dto.ID != nil {
		if id, perr := uuid.Parse(*dto.ID); perr == nil {
			existing, ferr := s.ownMessageInConv(ctx, id, convID, caller.UserID)
			switch {
			case ferr == nil:
				return existing, nil
			case errors.Is(ferr, domain.ErrNotFound):
				// fresh id — proceed with the insert
			default:
				return nil, ferr
			}
		}
	}

	// Attachment authz: each media row must exist, belong to the caller's org,
	// and already be bound to this conversation. The client presigns via the
	// media endpoint with model_type=conversation, model_id=<convID> BEFORE
	// sending the message — this is the actual authz gate, since PresignUpload
	// does not verify write access to arbitrary model_ids.
	if len(dto.MediaIDs) > 0 {
		if err := s.media.ValidateAttachments(ctx, *caller.OrgID, convID, dto.MediaIDs); err != nil {
			return nil, err
		}
	}

	msg := &domain.ConversationMessage{
		ConversationID: convID,
		SenderID:       &caller.UserID,
		Content:        dto.Content,
		MediaIDs:       json.RawMessage(`[]`),
		AsDocument:     dto.AsDocument && len(dto.MediaIDs) > 0,
	}
	if dto.ID != nil {
		id, _ := uuid.Parse(*dto.ID)
		msg.ID = id
	}
	if len(dto.MediaIDs) > 0 {
		if b, merr := json.Marshal(dto.MediaIDs); merr == nil {
			msg.MediaIDs = b
		}
	}
	if dto.ReplyToMessageID != nil {
		rid, perr := uuid.Parse(*dto.ReplyToMessageID)
		if perr != nil {
			return nil, domain.NewValidationError(map[string]string{"reply_to_message_id": "invalid uuid"})
		}
		parent, ferr := s.messageRepo.FindByID(ctx, rid)
		if ferr != nil || parent.ConversationID != convID {
			return nil, domain.NewValidationError(map[string]string{"reply_to_message_id": "not found in conversation"})
		}
		msg.ReplyToMessageID = &rid
	}

	if err := s.messageRepo.Create(ctx, msg); err != nil {
		if errors.Is(err, domain.ErrConflict) && dto.ID != nil {
			// idempotent race: someone inserted our id first — return it only
			// if it's OUR message in THIS conversation (same scoping as above).
			id, _ := uuid.Parse(*dto.ID)
			existing, ferr := s.ownMessageInConv(ctx, id, convID, caller.UserID)
			if ferr == nil {
				return existing, nil
			}
			if errors.Is(ferr, domain.ErrConflict) {
				return nil, domain.ErrConflict
			}
			return nil, err
		}
		return nil, err
	}
	_ = s.convRepo.Touch(ctx, convID)

	// Parse + persist mentions: only conversation members may be mentioned,
	// and the sender is never mentioned (self-mentions are dropped).
	var mentioned []uuid.UUID
	if len(dto.MentionUserIDs) > 0 {
		memberIDs, _ := s.memberRepo.ListUserIDs(ctx, convID)
		memberSet := map[uuid.UUID]bool{}
		for _, id := range memberIDs {
			memberSet[id] = true
		}
		for _, idStr := range dto.MentionUserIDs {
			if uid, perr := uuid.Parse(idStr); perr == nil && memberSet[uid] && uid != caller.UserID {
				mentioned = append(mentioned, uid)
			}
		}
		if len(mentioned) > 0 {
			if err := s.mentionRepo.CreateMany(ctx, msg.ID, mentioned); err != nil {
				s.logger.Error("conversations.SendMessage mentions", "message_id", msg.ID, "error", err)
			}
		}
	}

	// Attribute the sender on the returned row: the repo Create doesn't preload
	// Sender, and with `json:"sender,omitempty"` a nil Sender is dropped from the
	// POST response — which would clobber the WS echo's good sender on the client
	// merge. SenderID is the caller here, so populate it from caller identity.
	if msg.Sender == nil && msg.SenderID != nil && *msg.SenderID == caller.UserID {
		msg.Sender = &domain.User{ID: caller.UserID, Name: caller.Name}
	}

	// Phase 3 Step 5 wires notifications via `mentioned`; Phase 2 wires rt broadcast.
	s.afterSend(ctx, conv, msg, caller, mentioned)
	return msg, nil
}

// ownMessageInConv resolves a client-supplied idempotency id: it returns the
// existing message iff it lives in convID and was sent by senderID. A foreign
// (other conversation / other sender) id yields ErrConflict so it can neither
// be read nor hijacked; an unknown id passes ErrNotFound through.
func (s *service) ownMessageInConv(ctx context.Context, id, convID, senderID uuid.UUID) (*domain.ConversationMessage, error) {
	existing, err := s.messageRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing.ConversationID != convID || existing.SenderID == nil || *existing.SenderID != senderID {
		return nil, domain.ErrConflict
	}
	return existing, nil
}

// afterSend fans a freshly-sent message out to realtime + notifications. The
// member roster is fetched ONCE and the mute filter applied once, shared by
// both tiers. mentioned holds the validated (member, non-self) mention user
// ids persisted by SendMessage, consumed to resolve group notification
// recipients. Everything here is best-effort: failures are logged, never
// surfaced to the sender.
func (s *service) afterSend(ctx context.Context, conv *domain.Conversation, msg *domain.ConversationMessage, caller domain.Caller, mentioned []uuid.UUID) {
	if s.rt == nil && s.notif == nil {
		return
	}
	members, merr := s.memberRepo.ListByConversation(ctx, conv.ID)
	if merr != nil {
		s.logger.Error("conversations.afterSend member list", "conversation_id", conv.ID, "error", merr)
		return
	}
	unmuted := unmutedRecipients(members, caller.UserID, time.Now())

	if s.rt != nil {
		s.rt.ToConversation(ctx, conv.ID, "new_message", messagePayload(msg, caller))

		// Two-tier fanout: also nudge each unmuted non-sender member's
		// per-user channel with a compact payload, so a client's sidebar can
		// update without having joined this conversation's WS room.
		//
		// Distinct event type from the room's "new_message": a client viewing
		// this conversation receives BOTH the room event (full payload) and
		// this per-user firehose (compact payload) for the same message id.
		// Sharing the type would let an id-dedup drop the full event and keep
		// the compact one, rendering a message with a missing sender/reply/
		// media. "conversation_bump" keeps the firehose a sidebar-only signal.
		s.rt.ToUsers(ctx, unmuted, "conversation_bump", map[string]any{
			"conversation_id": conv.ID,
			"id":              msg.ID,
			"sender_id":       msg.SenderID,
			// Sender name so a group row's "Name: preview" renders without the
			// full-payload room event (recipients haven't joined this room).
			"sender": map[string]any{"id": caller.UserID.String(), "name": caller.Name},
			"content":    msg.Content,
			"created_at": msg.CreatedAt,
		})
	}

	if s.notif == nil {
		return
	}
	var recipients []uuid.UUID
	switch conv.Type {
	case domain.ConversationTypeDirect, domain.ConversationTypeChannel:
		recipients = unmuted
	case domain.ConversationTypeGroup:
		// groups notify only @mentioned (still mute-gated)
		unmutedSet := make(map[uuid.UUID]bool, len(unmuted))
		for _, id := range unmuted {
			unmutedSet[id] = true
		}
		for _, id := range mentioned {
			if unmutedSet[id] {
				recipients = append(recipients, id)
			}
		}
	}
	if len(recipients) > 0 {
		if err := s.notif.NotifyMessage(ctx, conv, msg, recipients); err != nil {
			s.logger.Error("conversations.afterSend notify", "conversation_id", conv.ID, "error", err)
		}
	}
}

// messagePayload mirrors the legacy chatMessagePayload shape for the realtime
// wire format. Sender identity comes from the message row itself — falling
// back to the caller only for a freshly-created row (whose SenderID IS the
// caller and has no preloaded Sender) — so an admin editing someone else's
// message never re-attributes it to the editor.
func messagePayload(msg *domain.ConversationMessage, caller domain.Caller) map[string]any {
	var senderID, sender any
	if msg.SenderID != nil {
		id := msg.SenderID.String()
		name := ""
		switch {
		case msg.Sender != nil:
			name = msg.Sender.Name
		case *msg.SenderID == caller.UserID:
			name = caller.Name
		}
		senderID = id
		sender = map[string]any{"id": id, "name": name}
	}
	return map[string]any{
		"id":                  msg.ID.String(),
		"conversation_id":     msg.ConversationID.String(),
		"sender_id":           senderID,
		"sender":              sender,
		"content":             msg.Content,
		"reply_to_message_id": msg.ReplyToMessageID,
		"media_ids":           msg.MediaIDs,
		"as_document":         msg.AsDocument,
		"is_edited":           msg.IsEdited,
		"created_at":          msg.CreatedAt,
	}
}

// serializeMessages populates each message's computed Reactions field with a
// single GROUP BY (message_id, emoji) query for the page. Best-effort: a
// failed count query leaves Reactions empty rather than failing the read.
// media_ids are returned as-is (raw, unsigned) — the client resolves URLs on
// demand via GET /media/:id/download-url, so no media-signing port is needed.
func (s *service) serializeMessages(ctx context.Context, msgs []domain.ConversationMessage) []domain.ConversationMessage {
	if len(msgs) == 0 {
		return msgs
	}
	ids := make([]uuid.UUID, len(msgs))
	for i := range msgs {
		ids[i] = msgs[i].ID
	}
	counts, err := s.reactionRepo.CountByMessages(ctx, ids)
	if err != nil {
		return msgs
	}
	for i := range msgs {
		if c, ok := counts[msgs[i].ID]; ok {
			msgs[i].Reactions = c
		}
	}
	return msgs
}

func (s *service) ListMessages(ctx context.Context, convID uuid.UUID, cur domain.MessageCursor) ([]domain.ConversationMessage, error) {
	caller, err := s.caller(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireMember(ctx, convID, caller.UserID); err != nil {
		return nil, err
	}
	msgs, err := s.messageRepo.ListWindow(ctx, convID, cur)
	if err != nil {
		return nil, err
	}
	return s.serializeMessages(ctx, msgs), nil
}

func (s *service) EditMessage(ctx context.Context, msgID uuid.UUID, dto domain.UpdateConversationMessageDTO) (*domain.ConversationMessage, error) {
	caller, conv, msg, err := s.msgForSenderOrManage(ctx, msgID)
	if err != nil {
		return nil, err
	}
	msg.Content = dto.Content
	msg.IsEdited = true
	if err := s.messageRepo.Update(ctx, msg); err != nil {
		return nil, err
	}
	if s.rt != nil {
		s.rt.ToConversation(ctx, conv.ID, "message_updated", messagePayload(msg, caller))
	}
	return msg, nil
}

func (s *service) DeleteMessage(ctx context.Context, msgID uuid.UUID) error {
	_, conv, _, err := s.msgForSenderOrManage(ctx, msgID)
	if err != nil {
		return err
	}
	if err := s.messageRepo.Delete(ctx, msgID); err != nil {
		return err
	}
	if s.rt != nil {
		s.rt.ToConversation(ctx, conv.ID, "message_deleted", map[string]any{
			"id":              msgID.String(),
			"conversation_id": conv.ID.String(),
		})
	}
	return nil
}

// ToggleReaction mirrors the legacy chat ToggleReaction tx-based toggle,
// adjusted for the new schema: ConversationMessage.Reactions is a
// gorm:"-" computed field (not a persisted JSON column), so unlike legacy
// chat there is no message row write here — only the reaction row itself
// changes, and the returned message's Reactions map is filled from a fresh
// CountByMessage after the toggle commits.
func (s *service) ToggleReaction(ctx context.Context, msgID uuid.UUID, dto domain.ToggleConversationReactionDTO) (*domain.ConversationMessage, error) {
	caller, err := s.caller(ctx)
	if err != nil {
		return nil, err
	}
	msg, err := s.messageRepo.FindByID(ctx, msgID)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireMember(ctx, msg.ConversationID, caller.UserID); err != nil {
		return nil, err
	}

	added := false
	err = s.transactor.RunInTx(ctx, func(txCtx context.Context) error {
		existing, ferr := s.reactionRepo.FindByMessageAndUser(txCtx, msgID, caller.UserID, dto.Emoji)
		if ferr != nil && !errors.Is(ferr, domain.ErrNotFound) {
			return ferr
		}
		if existing != nil {
			return s.reactionRepo.Delete(txCtx, msgID, caller.UserID, dto.Emoji)
		}
		added = true
		return s.reactionRepo.Create(txCtx, &domain.ConversationMessageReaction{
			MessageID: msgID,
			UserID:    caller.UserID,
			Emoji:     dto.Emoji,
			CreatedAt: time.Now(),
		})
	})
	if err != nil {
		return nil, err
	}

	counts, err := s.reactionRepo.CountByMessage(ctx, msgID)
	if err != nil {
		return nil, err
	}
	msg.Reactions = counts
	if s.rt != nil {
		event := "reaction_removed"
		if added {
			event = "reaction_added"
		}
		s.rt.ToConversation(ctx, msg.ConversationID, event, map[string]any{
			"message_id": msgID.String(),
			"emoji":      dto.Emoji,
			"user_id":    caller.UserID.String(),
			"counts":     counts,
		})
	}
	return msg, nil
}

func (s *service) MarkRead(ctx context.Context, convID uuid.UUID, dto domain.MarkReadDTO) error {
	caller, err := s.caller(ctx)
	if err != nil {
		return err
	}
	if _, err := s.requireMember(ctx, convID, caller.UserID); err != nil {
		return err
	}
	msgID, perr := uuid.Parse(dto.MessageID)
	if perr != nil {
		return domain.NewValidationError(map[string]string{"message_id": "invalid uuid"})
	}
	// The read pointer must reference a message in THIS conversation — unread
	// counts are keyset comparisons on it, so an arbitrary (foreign) id would
	// corrupt them.
	msg, err := s.messageRepo.FindByID(ctx, msgID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.NewValidationError(map[string]string{"message_id": "not found in conversation"})
		}
		return err
	}
	if msg.ConversationID != convID {
		return domain.NewValidationError(map[string]string{"message_id": "not found in conversation"})
	}
	if err := s.memberRepo.SetLastRead(ctx, convID, caller.UserID, msgID, time.Now()); err != nil {
		return err
	}
	if s.rt != nil {
		s.rt.ToConversation(ctx, convID, "message_read", map[string]any{
			"conversation_id": convID.String(),
			"user_id":         caller.UserID.String(),
			"message_id":      msgID.String(),
		})
	}
	return nil
}

func (s *service) SetMuted(ctx context.Context, convID uuid.UUID, until *time.Time) error {
	caller, err := s.caller(ctx)
	if err != nil {
		return err
	}
	if _, err := s.requireMember(ctx, convID, caller.UserID); err != nil {
		return err
	}
	return s.memberRepo.SetMuted(ctx, convID, caller.UserID, until)
}

func (s *service) PinMessage(ctx context.Context, msgID uuid.UUID) error {
	caller, _, err := s.msgConvForManage(ctx, msgID)
	if err != nil {
		return err
	}
	now := time.Now()
	return s.messageRepo.SetPinned(ctx, msgID, true, &caller.UserID, &now)
}

func (s *service) UnpinMessage(ctx context.Context, msgID uuid.UUID) error {
	if _, _, err := s.msgConvForManage(ctx, msgID); err != nil {
		return err
	}
	return s.messageRepo.SetPinned(ctx, msgID, false, nil, nil)
}

func (s *service) ListPinned(ctx context.Context, convID uuid.UUID) ([]domain.ConversationMessage, error) {
	caller, err := s.caller(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireMember(ctx, convID, caller.UserID); err != nil {
		return nil, err
	}
	msgs, err := s.messageRepo.ListPinned(ctx, convID)
	if err != nil {
		return nil, err
	}
	return s.serializeMessages(ctx, msgs), nil
}

// Search performs a global ranked full-text search across every conversation
// the caller is a member of, org-scoped.
func (s *service) Search(ctx context.Context, q string, limit int) ([]domain.ConversationMessage, error) {
	caller, err := s.caller(ctx)
	if err != nil {
		return nil, err
	}
	if len(q) < 3 {
		return nil, domain.NewValidationError(map[string]string{"q": "must be at least 3 characters"})
	}
	msgs, err := s.messageRepo.SearchGlobal(ctx, *caller.OrgID, caller.UserID, q, limit)
	if err != nil {
		return nil, err
	}
	return s.serializeMessages(ctx, msgs), nil
}

// SearchInConversation performs an in-conversation ILIKE nav search, gated on
// hard membership (same tier as ListMessages).
func (s *service) SearchInConversation(ctx context.Context, convID uuid.UUID, q string, limit int) ([]domain.ConversationMessage, error) {
	caller, err := s.caller(ctx)
	if err != nil {
		return nil, err
	}
	if len(q) < 2 {
		return nil, domain.NewValidationError(map[string]string{"q": "must be at least 2 characters"})
	}
	if _, err := s.requireMember(ctx, convID, caller.UserID); err != nil {
		return nil, err
	}
	msgs, err := s.messageRepo.SearchInConversation(ctx, convID, q, limit)
	if err != nil {
		return nil, err
	}
	return s.serializeMessages(ctx, msgs), nil
}

// Presence returns the online/last-seen status for the requested user ids,
// restricted to users in the caller's organization (one batch query). Ids in
// a different org (or that don't resolve to a user) are silently dropped, so
// callers can only observe presence of people they could plausibly share a
// conversation with. Requested ids are capped by the handler (see GetPresence).
func (s *service) Presence(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]domain.PresenceStatus, error) {
	caller, err := s.caller(ctx)
	if err != nil {
		return nil, err
	}
	if s.presence == nil || len(ids) == 0 {
		return map[uuid.UUID]domain.PresenceStatus{}, nil
	}
	allowed, err := s.users.FilterSameOrg(ctx, *caller.OrgID, ids)
	if err != nil {
		return nil, err
	}
	return s.presence.Get(ctx, allowed)
}

package conversations

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

// broadcaster is the realtime port (Phase 2 supplies the WS+Redis impl; nil = no-op).
type broadcaster interface {
	ToConversation(ctx context.Context, convID uuid.UUID, eventType string, data any)
	ToUser(ctx context.Context, userID uuid.UUID, eventType string, data any)
}

// notifier is the notifications port (Phase 3 supplies impl; nil = no-op).
type notifier interface {
	NotifyMessage(ctx context.Context, conv *domain.Conversation, msg *domain.ConversationMessage, recipientIDs []uuid.UUID) error
}

type service struct {
	convRepo     domain.ConversationRepository
	memberRepo   domain.ConversationMemberRepository
	messageRepo  domain.ConversationMessageRepository
	reactionRepo domain.ConversationReactionRepository
	mentionRepo  domain.ConversationMentionRepository
	transactor   domain.Transactor
	logger       *slog.Logger
	rt           broadcaster // may be nil
	notif        notifier    // may be nil
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
) domain.ConversationService {
	return &service{convRepo, memberRepo, messageRepo, reactionRepo, mentionRepo, transactor, logger, rt, notif}
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
	// TODO(phase3): reject cross-org DM via userLookup port

	dk := directKey(caller.UserID, other)
	if existing, err := s.convRepo.FindDirect(ctx, *caller.OrgID, dk); err == nil {
		return existing, nil
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
			return s.convRepo.FindDirect(ctx, *caller.OrgID, dk)
		}
		return nil, err
	}
	return conv, nil
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

	var conv *domain.Conversation
	err = s.transactor.RunInTx(ctx, func(txCtx context.Context) error {
		conv = &domain.Conversation{
			OrganizationID: *caller.OrgID,
			Type:           dto.Type,
			Name:           dto.Name,
			Description:    dto.Description,
			ColorIndex:     dto.ColorIndex,
			CreatedBy:      &caller.UserID,
		}
		if cerr := s.convRepo.Create(txCtx, conv); cerr != nil {
			return cerr
		}
		now := time.Now()
		members := []domain.ConversationMember{
			{ConversationID: conv.ID, UserID: caller.UserID, Role: domain.ConversationMemberRoleAdmin, JoinedAt: now},
		}
		seen := map[uuid.UUID]bool{caller.UserID: true}
		// TODO(phase3): validate member user_ids belong to caller's org via userLookup port
		for _, idStr := range dto.MemberIDs {
			uid, perr := uuid.Parse(idStr)
			if perr != nil || seen[uid] {
				continue
			}
			seen[uid] = true
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
	return conv, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, dto domain.UpdateConversationDTO) (*domain.Conversation, error) {
	caller, err := s.caller(ctx)
	if err != nil {
		return nil, err
	}
	conv, err := s.convRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	member, err := s.memberOrNil(ctx, id, caller.UserID)
	if err != nil {
		return nil, err
	}
	if !s.canManageConversation(caller, conv, member) {
		return nil, domain.ErrForbidden
	}
	if dto.Name != nil {
		conv.Name = *dto.Name
	}
	if dto.Description != nil {
		conv.Description = *dto.Description
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
	caller, err := s.caller(ctx)
	if err != nil {
		return err
	}
	conv, err := s.convRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	member, err := s.memberOrNil(ctx, id, caller.UserID)
	if err != nil {
		return err
	}
	if !s.canManageConversation(caller, conv, member) {
		return domain.ErrForbidden
	}
	return s.convRepo.Delete(ctx, id)
}

// ListForCaller populates the computed UnreadCount + LastMessage fields per
// conversation. N+1 per page is acceptable for v1 (see plan Step 9).
func (s *service) ListForCaller(ctx context.Context, q domain.ListConversationsQuery) ([]domain.Conversation, int64, error) {
	caller, err := s.caller(ctx)
	if err != nil {
		return nil, 0, err
	}
	convs, total, err := s.convRepo.ListForUser(ctx, *caller.OrgID, caller.UserID, q)
	if err != nil {
		return nil, 0, err
	}
	for i := range convs {
		uc, uerr := s.memberRepo.UnreadCount(ctx, convs[i].ID, caller.UserID)
		if uerr != nil {
			return nil, 0, uerr
		}
		convs[i].UnreadCount = uc

		last, lerr := s.messageRepo.Latest(ctx, convs[i].ID)
		if lerr != nil {
			if !errors.Is(lerr, domain.ErrNotFound) {
				return nil, 0, lerr
			}
		} else {
			convs[i].LastMessage = last
		}
	}
	return convs, total, nil
}

func (s *service) AddMember(ctx context.Context, convID uuid.UUID, dto domain.AddConversationMemberDTO) (*domain.ConversationMember, error) {
	caller, err := s.caller(ctx)
	if err != nil {
		return nil, err
	}
	conv, err := s.convRepo.FindByID(ctx, convID)
	if err != nil {
		return nil, err
	}
	member, err := s.memberOrNil(ctx, convID, caller.UserID)
	if err != nil {
		return nil, err
	}
	if !s.canManageConversation(caller, conv, member) {
		return nil, domain.ErrForbidden
	}
	uid, perr := uuid.Parse(dto.UserID)
	if perr != nil {
		return nil, domain.NewValidationError(map[string]string{"user_id": "invalid uuid"})
	}
	// TODO(phase3): validate member user_ids belong to caller's org via userLookup port
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
	caller, err := s.caller(ctx)
	if err != nil {
		return err
	}
	conv, err := s.convRepo.FindByID(ctx, convID)
	if err != nil {
		return err
	}
	member, err := s.memberOrNil(ctx, convID, caller.UserID)
	if err != nil {
		return err
	}
	if !s.canManageConversation(caller, conv, member) {
		return domain.ErrForbidden
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

	// Idempotency: client-supplied id → pre-check. MUST be scoped to this
	// conversation + sender: an unscoped FindByID would let any member "send"
	// with a foreign message id and read (or hijack) messages from other
	// conversations.
	if dto.ID != nil {
		if id, perr := uuid.Parse(*dto.ID); perr == nil {
			if existing, ferr := s.messageRepo.FindByID(ctx, id); ferr == nil {
				if existing.ConversationID != convID || existing.SenderID == nil || *existing.SenderID != caller.UserID {
					return nil, domain.ErrConflict
				}
				return existing, nil
			}
		}
	}

	msg := &domain.ConversationMessage{
		ConversationID: convID,
		SenderID:       &caller.UserID,
		Content:        dto.Content,
		MediaIDs:       json.RawMessage(`[]`),
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
			existing, ferr := s.messageRepo.FindByID(ctx, id)
			if ferr != nil {
				return nil, err
			}
			if existing.ConversationID != convID || existing.SenderID == nil || *existing.SenderID != caller.UserID {
				return nil, domain.ErrConflict
			}
			return existing, nil
		}
		return nil, err
	}
	_ = s.convRepo.Touch(ctx, convID)

	// Phase 3 wires mentions + notifications; Phase 2 wires rt broadcast.
	s.afterSend(ctx, conv, msg, dto, caller)
	return msg, nil
}

// afterSend is a seam filled by Phase 2 (broadcast) + Phase 3 (mentions/notify).
// In Phase 1 it only fires the (nil-guarded) broadcast — no notify logic yet.
func (s *service) afterSend(ctx context.Context, conv *domain.Conversation, msg *domain.ConversationMessage, dto domain.SendConversationMessageDTO, caller domain.Caller) {
	if s.rt != nil {
		s.rt.ToConversation(ctx, conv.ID, "new_message", messagePayload(msg, caller))
	}
}

// messagePayload mirrors the legacy chatMessagePayload shape for the realtime
// wire format. Sender name comes from the caller, since the freshly-created
// row has no preloaded user.
func messagePayload(msg *domain.ConversationMessage, caller domain.Caller) map[string]any {
	return map[string]any{
		"id":                  msg.ID.String(),
		"conversation_id":     msg.ConversationID.String(),
		"sender_id":           caller.UserID.String(),
		"sender":              map[string]any{"id": caller.UserID.String(), "name": caller.Name},
		"content":             msg.Content,
		"reply_to_message_id": msg.ReplyToMessageID,
		"media_ids":           msg.MediaIDs,
		"created_at":          msg.CreatedAt,
	}
}

func (s *service) ListMessages(ctx context.Context, convID uuid.UUID, cur domain.MessageCursor) ([]domain.ConversationMessage, error) {
	caller, err := s.caller(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireMember(ctx, convID, caller.UserID); err != nil {
		return nil, err
	}
	return s.messageRepo.ListWindow(ctx, convID, cur)
}

func (s *service) EditMessage(ctx context.Context, msgID uuid.UUID, dto domain.UpdateConversationMessageDTO) (*domain.ConversationMessage, error) {
	caller, err := s.caller(ctx)
	if err != nil {
		return nil, err
	}
	msg, err := s.messageRepo.FindByID(ctx, msgID)
	if err != nil {
		return nil, err
	}
	conv, err := s.convRepo.FindByID(ctx, msg.ConversationID)
	if err != nil {
		return nil, err
	}
	if !s.canManageMessage(caller, conv, msg) {
		return nil, domain.ErrForbidden
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
	caller, err := s.caller(ctx)
	if err != nil {
		return err
	}
	msg, err := s.messageRepo.FindByID(ctx, msgID)
	if err != nil {
		return err
	}
	conv, err := s.convRepo.FindByID(ctx, msg.ConversationID)
	if err != nil {
		return err
	}
	if !s.canManageMessage(caller, conv, msg) {
		return domain.ErrForbidden
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
	caller, err := s.caller(ctx)
	if err != nil {
		return err
	}
	msg, err := s.messageRepo.FindByID(ctx, msgID)
	if err != nil {
		return err
	}
	conv, err := s.convRepo.FindByID(ctx, msg.ConversationID)
	if err != nil {
		return err
	}
	member, err := s.memberOrNil(ctx, msg.ConversationID, caller.UserID)
	if err != nil {
		return err
	}
	if !s.canManageConversation(caller, conv, member) {
		return domain.ErrForbidden
	}
	now := time.Now()
	return s.messageRepo.SetPinned(ctx, msgID, true, &caller.UserID, &now)
}

func (s *service) UnpinMessage(ctx context.Context, msgID uuid.UUID) error {
	caller, err := s.caller(ctx)
	if err != nil {
		return err
	}
	msg, err := s.messageRepo.FindByID(ctx, msgID)
	if err != nil {
		return err
	}
	conv, err := s.convRepo.FindByID(ctx, msg.ConversationID)
	if err != nil {
		return err
	}
	member, err := s.memberOrNil(ctx, msg.ConversationID, caller.UserID)
	if err != nil {
		return err
	}
	if !s.canManageConversation(caller, conv, member) {
		return domain.ErrForbidden
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
	return s.messageRepo.ListPinned(ctx, convID)
}

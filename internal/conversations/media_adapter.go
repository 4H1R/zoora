package conversations

import (
	"context"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

// AttachmentValidator implements the service's attachmentValidator port over
// domain.MediaRepository.
type AttachmentValidator struct {
	media domain.MediaRepository
}

func NewAttachmentValidator(media domain.MediaRepository) *AttachmentValidator {
	return &AttachmentValidator{media: media}
}

// ValidateAttachments enforces the attachment authz rule for SendMessage:
// every media id must exist, belong to the caller's org, and already be bound
// to this conversation (model_type=conversation, model_id=convID) — binding
// happens at presign time, which is the actual write-authz gate.
func (v *AttachmentValidator) ValidateAttachments(ctx context.Context, orgID, convID uuid.UUID, mediaIDs []string) error {
	for _, idStr := range mediaIDs {
		mid, err := uuid.Parse(idStr)
		if err != nil {
			return domain.NewValidationError(map[string]string{"media_ids": "invalid uuid"})
		}
		med, err := v.media.FindByID(ctx, mid)
		if err != nil {
			return domain.NewValidationError(map[string]string{"media_ids": "attachment not found"})
		}
		if med.OrganizationID == nil || *med.OrganizationID != orgID ||
			med.ModelType != domain.MediaModelConversation || med.ModelID != convID {
			return domain.NewValidationError(map[string]string{"media_ids": "attachment does not belong to this conversation"})
		}
	}
	return nil
}

package main

import (
	"context"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/chathub"
	"github.com/4H1R/zoora/internal/domain"
)

// presenceReaderAdapter bridges the concrete chathub.Presence tracker to the
// conversations service's narrow presenceReader port, converting chathub.Status
// to the domain type so the conversations package need not import chathub.
type presenceReaderAdapter struct{ p *chathub.Presence }

func (a presenceReaderAdapter) Get(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]domain.PresenceStatus, error) {
	statuses, err := a.p.Get(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]domain.PresenceStatus, len(statuses))
	for id, s := range statuses {
		out[id] = domain.PresenceStatus{Online: s.Online, LastSeen: s.LastSeen}
	}
	return out, nil
}

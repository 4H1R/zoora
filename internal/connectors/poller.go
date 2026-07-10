package connectors

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/bots"
)

// Poller long-polls one bot for /start <token> messages and completes
// connector links. One Poller per platform (telegram, bale) in the worker.
type Poller struct {
	bot    *bots.Client
	svc    domain.ConnectorService
	kind   domain.ConnectorType
	logger *slog.Logger
}

func NewPoller(bot *bots.Client, svc domain.ConnectorService, kind domain.ConnectorType, logger *slog.Logger) *Poller {
	if logger == nil {
		logger = slog.Default()
	}
	return &Poller{bot: bot, svc: svc, kind: kind, logger: logger}
}

// Run blocks until ctx is canceled. Errors back off and retry — a dead bot
// connection must not kill the worker.
func (p *Poller) Run(ctx context.Context) {
	var offset int64
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		updates, err := p.bot.GetUpdates(ctx, offset, 30)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			p.logger.Warn("bot poll failed", "bot", p.kind, "error", err)
			time.Sleep(5 * time.Second)
			continue
		}
		for _, u := range updates {
			offset = u.UpdateID + 1
			if u.Message == nil {
				continue
			}
			p.handleMessage(ctx, u.Message)
		}
	}
}

func (p *Poller) handleMessage(ctx context.Context, m *bots.Message) {
	token, ok := strings.CutPrefix(strings.TrimSpace(m.Text), "/start ")
	if !ok {
		return
	}
	chatID := strconv.FormatInt(m.Chat.ID, 10)
	res, err := p.svc.CompleteLink(ctx, p.kind, strings.TrimSpace(token), chatID)
	if err != nil {
		p.logger.Warn("link completion failed", "bot", p.kind, "error", err)
		_ = p.bot.SendMessage(ctx, chatID, "Link failed or expired. Request a new link from your Zoora settings.")
		return
	}
	_ = p.bot.SendMessage(ctx, chatID, connectedMessage(res))
}

// connectedMessage builds the /start confirmation, naming the account and org
// when available and degrading gracefully when enrichment lookups came back
// empty.
func connectedMessage(res *domain.ConnectorLinkResult) string {
	if res == nil || res.Username == "" {
		return "✅ Connected! You will now receive Zoora notifications here."
	}
	who := "@" + res.Username
	if res.Name != "" {
		who += " (" + res.Name + ")"
	}
	if res.OrgName != "" {
		return "✅ Connected as " + who + " · " + res.OrgName + ".\nYou will now receive Zoora notifications for this account here."
	}
	return "✅ Connected as " + who + ".\nYou will now receive Zoora notifications for this account here."
}

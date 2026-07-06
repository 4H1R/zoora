// Package push implements domain.PushSender over Firebase Cloud Messaging
// (web push). Credentials come from a service-account JSON file.
package push

import (
	"context"
	"fmt"
	"log/slog"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

const multicastLimit = 500 // FCM hard cap per multicast call

type FCM struct {
	client *messaging.Client
	logger *slog.Logger
}

func NewFCM(ctx context.Context, credentialsFile string, logger *slog.Logger) (*FCM, error) {
	if logger == nil {
		logger = slog.Default()
	}
	// WithCredentialsFile is marked deprecated by google.golang.org/api in favor
	// of the newer cloud.google.com/go/auth flow, but remains the supported way
	// to load a service-account file for firebase-admin-go v4. Safe here: the
	// path comes from trusted server config, not user input.
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsFile(credentialsFile)) //nolint:staticcheck // SA1019: file-based service-account creds are intentional
	if err != nil {
		return nil, fmt.Errorf("push.fcm: initializing app: %w", err)
	}
	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("push.fcm: initializing messaging: %w", err)
	}
	return &FCM{client: client, logger: logger}, nil
}

// SendMulticast implements domain.PushSender. Returns tokens FCM reports as
// permanently invalid (unregistered) so callers can prune dead connectors.
func (f *FCM) SendMulticast(ctx context.Context, tokens []string, title, body, link string) ([]string, error) {
	var invalid []string
	for start := 0; start < len(tokens); start += multicastLimit {
		end := min(start+multicastLimit, len(tokens))
		batch := tokens[start:end]

		msg := &messaging.MulticastMessage{
			Tokens: batch,
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
			Webpush: &messaging.WebpushConfig{
				FCMOptions: &messaging.WebpushFCMOptions{Link: link},
			},
		}
		br, err := f.client.SendEachForMulticast(ctx, msg)
		if err != nil {
			return invalid, fmt.Errorf("push.fcm: multicast: %w", err)
		}
		for i, resp := range br.Responses {
			if resp.Error != nil && messaging.IsUnregistered(resp.Error) {
				invalid = append(invalid, batch[i])
			}
		}
	}
	return invalid, nil
}

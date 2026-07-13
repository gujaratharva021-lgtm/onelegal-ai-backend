package services

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"

	"legaltech-backend/internal/config"
)

// globalPushService is a package-level singleton, matching this codebase's
// existing convention of a shared global handle for a cross-cutting
// infrastructure concern (see internal/database.DB). Initialized once at
// startup via InitPushService — see routes.Setup — so every service that
// wants to fire a push (via NotificationService.Create) doesn't need
// *config.Config threaded through its own constructor.
var globalPushService *PushService

// globalConfig mirrors the same pattern for *config.Config itself — services
// constructed with a zero-arg NewXxxService() (e.g. InvoiceService's internal
// EmailService) that still need config can call services.GetConfig() instead
// of every constructor threading it through by hand.
var globalConfig *config.Config

// InitPushService must be called once at startup (see routes.Setup) before
// any push is sent — it also stashes cfg for GetConfig(). Safe to call even
// when Firebase isn't configured — NewPushService handles that by returning
// a no-op service.
func InitPushService(cfg *config.Config) {
	globalConfig = cfg
	globalPushService = NewPushService(cfg)
}

// GetConfig returns the app config stashed by InitPushService.
func GetConfig() *config.Config {
	return globalConfig
}

// PushService sends real push notifications via Firebase Cloud Messaging.
// It is a safe no-op when FIREBASE_CREDENTIALS_FILE isn't configured (or the
// file/credentials are invalid) — every other feature keeps working exactly
// as before; push notifications just silently don't fire until it's set up.
// See README/setup instructions for how to create the Firebase project and
// service-account key.
type PushService struct {
	client *messaging.Client
}

func NewPushService(cfg *config.Config) *PushService {
	if cfg.FirebaseCredentialsFile == "" {
		log.Println("[push] FIREBASE_CREDENTIALS_FILE not set — push notifications disabled")
		return &PushService{}
	}

	app, err := firebase.NewApp(context.Background(), nil, option.WithCredentialsFile(cfg.FirebaseCredentialsFile))
	if err != nil {
		log.Printf("[push] Failed to initialize Firebase app (push notifications disabled): %v", err)
		return &PushService{}
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		log.Printf("[push] Failed to initialize Firebase Messaging client (push notifications disabled): %v", err)
		return &PushService{}
	}

	log.Println("[push] Firebase Cloud Messaging initialized")
	return &PushService{client: client}
}

// SendToToken pushes a single notification to one device token. Always
// best-effort: a missing token, disabled service, or FCM error is logged and
// swallowed — never propagated to the caller, matching the existing
// fire-and-forget convention for in-app notifications (see
// NotificationService.Create's callers).
func (s *PushService) SendToToken(deviceToken, title, body string, data map[string]string) {
	if s.client == nil || deviceToken == "" {
		return
	}

	msg := &messaging.Message{
		Token: deviceToken,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	if _, err := s.client.Send(context.Background(), msg); err != nil {
		log.Printf("[push] Failed to send push notification: %v", err)
	}
}

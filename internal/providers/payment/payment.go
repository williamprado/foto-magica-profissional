package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type CheckoutRequest struct {
	TenantID    uuid.UUID
	PlanCode    string
	Currency    string
	AmountCents int
	SuccessURL  string
	CancelURL   string
}

type CheckoutSession struct {
	Provider      string `json:"provider"`
	CheckoutURL   string `json:"checkoutUrl"`
	ExternalID    string `json:"externalId"`
	ExpiresAt     time.Time `json:"expiresAt"`
}

type WebhookEvent struct {
	Provider      string
	EventType     string
	ExternalRef   string
	TenantID      uuid.UUID
	Credits       int
	RawPayload    json.RawMessage
}

type Provider interface {
	Name() string
	CreateCheckoutSession(context.Context, CheckoutRequest) (CheckoutSession, error)
	ParseWebhook(context.Context, []byte, map[string]string) (WebhookEvent, error)
}

type MockProvider struct{}

func (MockProvider) Name() string { return "mock" }

func (MockProvider) CreateCheckoutSession(_ context.Context, req CheckoutRequest) (CheckoutSession, error) {
	return CheckoutSession{
		Provider:    "mock",
		CheckoutURL: fmt.Sprintf("https://billing.local/checkout/%s/%s", req.TenantID, req.PlanCode),
		ExternalID:  fmt.Sprintf("mock-%s", uuid.NewString()),
		ExpiresAt:   time.Now().Add(30 * time.Minute),
	}, nil
}

func (MockProvider) ParseWebhook(_ context.Context, payload []byte, _ map[string]string) (WebhookEvent, error) {
	event := WebhookEvent{Provider: "mock", EventType: "payment.approved", RawPayload: payload}
	return event, nil
}


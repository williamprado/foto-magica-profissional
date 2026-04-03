package billing

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/williamprado/foto-magica-profissional/internal/credits"
	"github.com/williamprado/foto-magica-profissional/internal/providers/payment"
)

type Plan struct {
	ID           uuid.UUID `json:"id"`
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	CreditAmount int       `json:"creditAmount"`
	PriceCents   int       `json:"priceCents"`
	Currency     string    `json:"currency"`
	Provider     string    `json:"provider"`
	Active       bool      `json:"active"`
}

type Service struct {
	db            *pgxpool.Pool
	credits       credits.Service
	providers     map[string]payment.Provider
	defaultProvider string
}

func NewService(db *pgxpool.Pool, credits credits.Service, providers map[string]payment.Provider, defaultProvider string) Service {
	return Service{db: db, credits: credits, providers: providers, defaultProvider: defaultProvider}
}

func (s Service) Plans(ctx context.Context) ([]Plan, error) {
	rows, err := s.db.Query(ctx, `
		select id, code, name, description, credit_amount, price_cents, currency, payment_provider, active
		from plans where active = true order by price_cents asc
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []Plan
	for rows.Next() {
		var plan Plan
		if err := rows.Scan(&plan.ID, &plan.Code, &plan.Name, &plan.Description, &plan.CreditAmount, &plan.PriceCents, &plan.Currency, &plan.Provider, &plan.Active); err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}
	return plans, nil
}

func (s Service) CreateCheckout(ctx context.Context, tenantID uuid.UUID, planCode string) (payment.CheckoutSession, error) {
	var plan Plan
	if err := s.db.QueryRow(ctx, `
		select id, code, name, description, credit_amount, price_cents, currency, payment_provider, active
		from plans where code = $1 and active = true
	`, planCode).Scan(&plan.ID, &plan.Code, &plan.Name, &plan.Description, &plan.CreditAmount, &plan.PriceCents, &plan.Currency, &plan.Provider, &plan.Active); err != nil {
		return payment.CheckoutSession{}, err
	}

	providerName := plan.Provider
	if providerName == "" {
		providerName = s.defaultProvider
	}
	provider, ok := s.providers[providerName]
	if !ok {
		return payment.CheckoutSession{}, fmt.Errorf("billing provider %s not configured", providerName)
	}

	session, err := provider.CreateCheckoutSession(ctx, payment.CheckoutRequest{
		TenantID:    tenantID,
		PlanCode:    plan.Code,
		Currency:    plan.Currency,
		AmountCents: plan.PriceCents,
		SuccessURL:  "https://app.fotomagica.local/sucesso",
		CancelURL:   "https://app.fotomagica.local/cancelado",
	})
	if err != nil {
		return payment.CheckoutSession{}, err
	}

	if _, err := s.db.Exec(ctx, `
		insert into subscriptions (tenant_id, plan_id, provider, provider_ref, status, current_period_start, current_period_end)
		values ($1, $2, $3, $4, 'pending', now(), now() + interval '30 days')
	`, tenantID, plan.ID, provider.Name(), session.ExternalID); err != nil {
		return payment.CheckoutSession{}, err
	}
	return session, nil
}

func (s Service) HandleWebhook(ctx context.Context, providerName string, headers map[string]string, payload []byte) error {
	provider, ok := s.providers[providerName]
	if !ok {
		return fmt.Errorf("provider %s not configured", providerName)
	}
	event, err := provider.ParseWebhook(ctx, payload, headers)
	if err != nil {
		return err
	}
	if _, err := s.db.Exec(ctx, `
		insert into billing_audit_logs (tenant_id, provider, event_type, payload)
		values ($1, $2, $3, $4)
	`, event.TenantID, event.Provider, event.EventType, event.RawPayload); err != nil {
		return err
	}
	if event.Credits > 0 && event.TenantID != uuid.Nil {
		if err := s.credits.CreditTopUp(ctx, event.TenantID, event.Credits, "Webhook billing credit"); err != nil {
			return err
		}
	}
	return nil
}

func SeedAuditPayload(value any) json.RawMessage {
	raw, _ := json.Marshal(value)
	return raw
}

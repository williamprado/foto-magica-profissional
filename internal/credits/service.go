package credits

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInsufficientCredits = errors.New("insufficient credits")

type WalletSummary struct {
	Balance       int `json:"balance"`
	Pending       int `json:"pending"`
	LifetimeSpent int `json:"lifetimeSpent"`
}

type Transaction struct {
	ID           uuid.UUID `json:"id"`
	TenantID     uuid.UUID `json:"tenantId"`
	Type         string    `json:"type"`
	Amount       int       `json:"amount"`
	BalanceAfter int       `json:"balanceAfter"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) Service {
	return Service{db: db}
}

func (s Service) Wallet(ctx context.Context, tenantID uuid.UUID) (WalletSummary, error) {
	var wallet WalletSummary
	if err := s.db.QueryRow(ctx, `
		select balance, pending_balance, lifetime_spent from credit_wallets where tenant_id = $1
	`, tenantID).Scan(&wallet.Balance, &wallet.Pending, &wallet.LifetimeSpent); err != nil {
		return WalletSummary{}, fmt.Errorf("wallet summary: %w", err)
	}
	return wallet, nil
}

func (s Service) DebitForGeneration(ctx context.Context, tenantID, userID uuid.UUID, amount int, description string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var balance int
	if err := tx.QueryRow(ctx, `select balance from credit_wallets where tenant_id = $1 for update`, tenantID).Scan(&balance); err != nil {
		return err
	}
	if balance < amount {
		return ErrInsufficientCredits
	}

	balance -= amount
	if _, err := tx.Exec(ctx, `
		update credit_wallets
		set balance = $2, lifetime_spent = lifetime_spent + $3, updated_at = now()
		where tenant_id = $1
	`, tenantID, balance, amount); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		insert into credit_transactions (tenant_id, type, amount, balance_after, description, created_by)
		values ($1, 'debit', $2, $3, $4, $5)
	`, tenantID, amount, balance, description, userID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s Service) CreditTopUp(ctx context.Context, tenantID uuid.UUID, amount int, description string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var balance int
	if err := tx.QueryRow(ctx, `select balance from credit_wallets where tenant_id = $1 for update`, tenantID).Scan(&balance); err != nil {
		return err
	}
	balance += amount
	if _, err := tx.Exec(ctx, `
		update credit_wallets
		set balance = $2, updated_at = now()
		where tenant_id = $1
	`, tenantID, balance); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		insert into credit_transactions (tenant_id, type, amount, balance_after, description)
		values ($1, 'credit', $2, $3, $4)
	`, tenantID, amount, balance, description); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

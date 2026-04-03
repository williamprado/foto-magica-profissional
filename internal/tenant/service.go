package tenant

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Tenant struct {
	ID        uuid.UUID `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	PlanCode  string    `json:"planCode"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return Repository{db: db}
}

func (r Repository) Create(ctx context.Context, tx pgxTx, name, slug, planCode string) (Tenant, error) {
	tenant := Tenant{}
	err := tx.QueryRow(ctx, `
		insert into tenants (name, slug, plan_code, status)
		values ($1, $2, $3, 'active')
		returning id, slug, name, plan_code, status, created_at
	`, name, slug, planCode).Scan(&tenant.ID, &tenant.Slug, &tenant.Name, &tenant.PlanCode, &tenant.Status, &tenant.CreatedAt)
	if err != nil {
		return Tenant{}, fmt.Errorf("create tenant: %w", err)
	}
	return tenant, nil
}

func (r Repository) FindBySlug(ctx context.Context, slug string) (Tenant, error) {
	tenant := Tenant{}
	err := r.db.QueryRow(ctx, `
		select id, slug, name, plan_code, status, created_at
		from tenants where slug = $1
	`, slug).Scan(&tenant.ID, &tenant.Slug, &tenant.Name, &tenant.PlanCode, &tenant.Status, &tenant.CreatedAt)
	if err != nil {
		return Tenant{}, fmt.Errorf("find tenant by slug: %w", err)
	}
	return tenant, nil
}

type pgxTx interface {
	QueryRow(ctx context.Context, sql string, args ...any) rowScanner
}

type rowScanner interface {
	Scan(dest ...any) error
}


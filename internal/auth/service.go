package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/williamprado/foto-magica-profissional/internal/rbac"
	"github.com/williamprado/foto-magica-profissional/internal/tenant"
)

type JWTManager interface {
	Issue(SessionClaims) (string, time.Time, error)
	Verify(token string) (SessionClaims, error)
}

type SessionClaims struct {
	UserID   uuid.UUID
	TenantID uuid.UUID
	Role     string
	Email    string
}

type jwtManager struct {
	secret []byte
	ttl    time.Duration
}

func NewJWTManager(secret string, ttl time.Duration) JWTManager {
	return jwtManager{secret: []byte(secret), ttl: ttl}
}

func (j jwtManager) Issue(claims SessionClaims) (string, time.Time, error) {
	expiresAt := time.Now().Add(j.ttl)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":       claims.UserID.String(),
		"tenant_id": claims.TenantID.String(),
		"role":      claims.Role,
		"email":     claims.Email,
		"exp":       expiresAt.Unix(),
	})
	signed, err := token.SignedString(j.secret)
	return signed, expiresAt, err
}

func (j jwtManager) Verify(token string) (SessionClaims, error) {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) { return j.secret, nil })
	if err != nil || !parsed.Valid {
		return SessionClaims{}, fmt.Errorf("verify token: %w", err)
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return SessionClaims{}, fmt.Errorf("invalid token claims")
	}

	userID, err := uuid.Parse(fmt.Sprint(claims["sub"]))
	if err != nil {
		return SessionClaims{}, err
	}
	tenantID, err := uuid.Parse(fmt.Sprint(claims["tenant_id"]))
	if err != nil {
		return SessionClaims{}, err
	}

	return SessionClaims{
		UserID:   userID,
		TenantID: tenantID,
		Role:     fmt.Sprint(claims["role"]),
		Email:    fmt.Sprint(claims["email"]),
	}, nil
}

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	FullName     string    `json:"fullName"`
	Role         string    `json:"role"`
	Credits      int       `json:"credits"`
	TenantID     uuid.UUID `json:"tenantId"`
	TenantSlug   string    `json:"tenantSlug"`
	TenantName   string    `json:"tenantName"`
	TenantPlan   string    `json:"tenantPlan"`
	PasswordHash string
}

type Session struct {
	Token     string      `json:"token"`
	ExpiresAt time.Time   `json:"expiresAt"`
	User      User        `json:"user"`
	Tenant    tenant.Tenant `json:"tenant"`
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return Repository{db: db}
}

type Service struct {
	db          *pgxpool.Pool
	repo        Repository
	tenants     tenant.Repository
	jwt         JWTManager
	defaultPlan string
}

func NewService(db *pgxpool.Pool, tenants tenant.Repository, jwt JWTManager, defaultPlan string) Service {
	return Service{
		db:         db,
		repo:       NewRepository(db),
		tenants:    tenants,
		jwt:        jwt,
		defaultPlan: defaultPlan,
	}
}

type RegisterInput struct {
	CompanyName string `json:"companyName" binding:"required"`
	CompanySlug string `json:"companySlug" binding:"required"`
	FullName    string `json:"fullName" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
}

type LoginInput struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required"`
	TenantSlug string `json:"tenantSlug"`
}

func (s Service) Register(ctx context.Context, input RegisterInput) (Session, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return Session{}, fmt.Errorf("hash password: %w", err)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Session{}, err
	}
	defer tx.Rollback(ctx)

	tenantRecord, err := s.tenants.Create(ctx, tx, input.CompanyName, input.CompanySlug, s.defaultPlan)
	if err != nil {
		return Session{}, err
	}

	user := User{}
	err = tx.QueryRow(ctx, `
		insert into users (email, password_hash, full_name, status)
		values ($1, $2, $3, 'active')
		returning id, email, full_name
	`, input.Email, string(hash), input.FullName).Scan(&user.ID, &user.Email, &user.FullName)
	if err != nil {
		return Session{}, fmt.Errorf("create user: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		insert into memberships (tenant_id, user_id, role)
		values ($1, $2, $3)
	`, tenantRecord.ID, user.ID, rbac.RoleOwner); err != nil {
		return Session{}, fmt.Errorf("create membership: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		insert into credit_wallets (tenant_id, balance, pending_balance, lifetime_spent)
		values ($1, 5, 0, 0)
	`, tenantRecord.ID); err != nil {
		return Session{}, fmt.Errorf("create wallet: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Session{}, err
	}

	user.Role = rbac.RoleOwner
	user.TenantID = tenantRecord.ID
	user.TenantSlug = tenantRecord.Slug
	user.TenantName = tenantRecord.Name
	user.TenantPlan = tenantRecord.PlanCode
	user.Credits = 5
	return s.newSession(user, tenantRecord)
}

func (s Service) Login(ctx context.Context, input LoginInput) (Session, error) {
	user, err := s.repo.FindByEmail(ctx, input.Email, input.TenantSlug)
	if err != nil {
		return Session{}, httpx.NewError(401, "invalid_credentials", "email or password invalid")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return Session{}, httpx.NewError(401, "invalid_credentials", "email or password invalid")
	}

	return s.newSession(user, tenant.Tenant{
		ID:       user.TenantID,
		Slug:     user.TenantSlug,
		Name:     user.TenantName,
		PlanCode: user.TenantPlan,
		Status:   "active",
	})
}

func (s Service) Me(ctx context.Context, claims SessionClaims) (Session, error) {
	user, err := s.repo.FindByID(ctx, claims.UserID, claims.TenantID)
	if err != nil {
		return Session{}, err
	}
	return s.newSession(user, tenant.Tenant{
		ID:       user.TenantID,
		Slug:     user.TenantSlug,
		Name:     user.TenantName,
		PlanCode: user.TenantPlan,
		Status:   "active",
	}), nil
}

func (s Service) newSession(user User, tenantRecord tenant.Tenant) (Session, error) {
	token, expiresAt, err := s.jwt.Issue(SessionClaims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Role:     user.Role,
		Email:    user.Email,
	})
	if err != nil {
		return Session{}, fmt.Errorf("issue token: %w", err)
	}

	return Session{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
		Tenant:    tenantRecord,
	}, nil
}

func (r Repository) FindByEmail(ctx context.Context, email, tenantSlug string) (User, error) {
	query := `
		select u.id, u.email, u.full_name, m.role, w.balance, t.id, t.slug, t.name, t.plan_code, u.password_hash
		from users u
		join memberships m on m.user_id = u.id
		join tenants t on t.id = m.tenant_id
		left join credit_wallets w on w.tenant_id = t.id
		where lower(u.email) = lower($1)
	`
	args := []any{email}
	if tenantSlug != "" {
		query += " and t.slug = $2"
		args = append(args, tenantSlug)
	}
	query += " order by m.created_at asc limit 1"

	var user User
	if err := r.db.QueryRow(ctx, query, args...).Scan(
		&user.ID, &user.Email, &user.FullName, &user.Role, &user.Credits,
		&user.TenantID, &user.TenantSlug, &user.TenantName, &user.TenantPlan, &user.PasswordHash,
	); err != nil {
		return User{}, err
	}
	return user, nil
}

func (r Repository) FindByID(ctx context.Context, userID, tenantID uuid.UUID) (User, error) {
	var user User
	if err := r.db.QueryRow(ctx, `
		select u.id, u.email, u.full_name, m.role, w.balance, t.id, t.slug, t.name, t.plan_code, u.password_hash
		from users u
		join memberships m on m.user_id = u.id
		join tenants t on t.id = m.tenant_id
		left join credit_wallets w on w.tenant_id = t.id
		where u.id = $1 and t.id = $2
	`, userID, tenantID).Scan(
		&user.ID, &user.Email, &user.FullName, &user.Role, &user.Credits,
		&user.TenantID, &user.TenantSlug, &user.TenantName, &user.TenantPlan, &user.PasswordHash,
	); err != nil {
		return User{}, err
	}
	return user, nil
}

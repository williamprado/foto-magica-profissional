package db

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(ctx context.Context, pool *pgxpool.Pool, baseDir string) error {
	if _, err := pool.Exec(ctx, `
		create table if not exists schema_migrations (
			version text primary key,
			applied_at timestamptz not null default now()
		)
	`); err != nil {
		return fmt.Errorf("ensure migrations table: %w", err)
	}

	migrationsDir := filepath.Join(baseDir, "services", "api", "migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		if err := applyMigration(ctx, pool, migrationsDir, entry); err != nil {
			return err
		}
	}

	return nil
}

func applyMigration(ctx context.Context, pool *pgxpool.Pool, migrationsDir string, entry fs.DirEntry) error {
	var exists bool
	if err := pool.QueryRow(ctx, "select exists(select 1 from schema_migrations where version = $1)", entry.Name()).Scan(&exists); err != nil {
		return fmt.Errorf("check migration %s: %w", entry.Name(), err)
	}
	if exists {
		return nil
	}

	sqlBytes, err := os.ReadFile(filepath.Join(migrationsDir, entry.Name()))
	if err != nil {
		return fmt.Errorf("read migration %s: %w", entry.Name(), err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", entry.Name(), err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, string(sqlBytes)); err != nil {
		return fmt.Errorf("execute migration %s: %w", entry.Name(), err)
	}
	if _, err := tx.Exec(ctx, "insert into schema_migrations(version) values ($1)", entry.Name()); err != nil {
		return fmt.Errorf("record migration %s: %w", entry.Name(), err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit migration %s: %w", entry.Name(), err)
	}
	return nil
}


package db

import (
    "context"
    "embed"
    "sort"
    "strings"

    "github.com/jmoiron/sqlx"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Runner struct{ DB *sqlx.DB }

func (r Runner) Migrate(ctx context.Context) error {
    if _, err := r.DB.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS stripe_sub_pkg_schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT now())`); err != nil {
        return err
    }
    entries, err := migrationsFS.ReadDir("migrations")
    if err != nil { return err }
    var names []string
    for _, e := range entries {
        if e.IsDir() { continue }
        n := e.Name()
        if !strings.HasSuffix(n, ".sql") { continue }
        names = append(names, n)
    }
    sort.Strings(names)
    for _, name := range names {
        version := strings.TrimSuffix(name, ".sql")
        var exists bool
        if err := r.DB.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM stripe_sub_pkg_schema_migrations WHERE version=$1)`, version); err != nil {
            return err
        }
        if exists { continue }
        b, err := migrationsFS.ReadFile("migrations/" + name)
        if err != nil { return err }
        if _, err := r.DB.ExecContext(ctx, string(b)); err != nil { return err }
        if _, err := r.DB.ExecContext(ctx, `INSERT INTO stripe_sub_pkg_schema_migrations(version) VALUES ($1)`, version); err != nil { return err }
    }
    return nil
}


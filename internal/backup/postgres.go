package backup

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/nardeo/nardeo-backup/internal/config"
)

// Postgres streams a plain SQL dump into a private gzip file and returns its path.
func Postgres(ctx context.Context, cfg config.Config) (string, error) {
	if err := cfg.Validate(); err != nil {
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	binary, err := exec.LookPath("pg_dump")
	if err != nil {
		return "", fmt.Errorf("pg_dump was not found in PATH; install the PostgreSQL client tools: %w", err)
	}
	if err := os.MkdirAll(cfg.OutputDir, 0700); err != nil {
		return "", fmt.Errorf("create output directory: %w", err)
	}
	stamp := time.Now().UTC().Format("20060102T150405.000000000Z")
	file, err := os.CreateTemp(cfg.OutputDir, "postgres_"+stamp+"-*.sql.gz.partial")
	if err != nil {
		return "", fmt.Errorf("create backup file: %w", err)
	}
	defer os.Remove(file.Name())
	defer file.Close()

	args, env, secrets := connection(cfg)
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Env = env
	cmd.WaitDelay = 5 * time.Second
	compressed := gzip.NewWriter(file)
	defer compressed.Close()
	cmd.Stdout = compressed
	diagnostics := &limitedBuffer{limit: 16 * 1024}
	cmd.Stderr = diagnostics
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return "", fmt.Errorf("PostgreSQL backup canceled: %w", ctx.Err())
		}
		detail := redact(strings.TrimSpace(diagnostics.String()), secrets)
		if detail != "" {
			return "", fmt.Errorf("pg_dump failed (check connection settings and database access): %w: %s", err, detail)
		}
		return "", fmt.Errorf("pg_dump failed (check connection settings and database access): %w", err)
	}
	if err := compressed.Close(); err != nil {
		return "", fmt.Errorf("finalize gzip stream: %w", err)
	}
	if err := file.Sync(); err != nil {
		return "", fmt.Errorf("sync backup file: %w", err)
	}
	if err := file.Close(); err != nil {
		return "", fmt.Errorf("close backup file: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("PostgreSQL backup canceled: %w", err)
	}
	finalPath := strings.TrimSuffix(file.Name(), ".partial")
	// Linking publishes the completed file atomically without replacing an existing backup.
	if err := os.Link(file.Name(), finalPath); err != nil {
		return "", fmt.Errorf("publish backup file: %w", err)
	}
	if detail := strings.TrimSpace(diagnostics.String()); detail != "" {
		slog.Warn("pg_dump reported a warning", "detail", redact(detail, secrets))
	}
	return filepath.Clean(finalPath), nil
}

func connection(cfg config.Config) ([]string, []string, []string) {
	args := []string{"--format=plain", "--no-password"}
	env := make([]string, 0, len(os.Environ())+5)
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		switch key {
		case "PGHOST", "PGHOSTADDR", "PGPORT", "PGUSER", "PGPASSWORD", "PGDATABASE", "PGSERVICE", "DATABASE_URL":
			continue
		}
		env = append(env, entry)
	}
	secrets := []string{cfg.Password, cfg.DatabaseURL}
	if cfg.DatabaseURL == "" {
		env = append(env, "PGHOST="+cfg.Host, "PGPORT="+cfg.Port, "PGUSER="+cfg.User, "PGPASSWORD="+cfg.Password, "PGDATABASE="+cfg.Database)
		return args, env, secrets
	}
	parsed, _ := config.ParseDatabaseURL(cfg.DatabaseURL) // Validate has checked the URL.
	secrets = append(secrets, parsed.Secrets...)
	env = append(env, "PGPASSWORD="+parsed.Password)
	args = append(args, "--dbname="+parsed.URL)
	return args, env, secrets
}

func redact(message string, secrets []string) string {
	for _, secret := range secrets {
		if secret == "" {
			continue
		}
		for _, encoded := range []string{secret, url.QueryEscape(secret), url.PathEscape(secret)} {
			message = strings.ReplaceAll(message, encoded, "[REDACTED]")
		}
	}
	return message
}

type limitedBuffer struct {
	bytes.Buffer
	limit int
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	n := len(p)
	if remaining := b.limit - b.Len(); remaining > 0 {
		if len(p) > remaining {
			p = p[:remaining]
		}
		_, _ = b.Buffer.Write(p)
	}
	return n, nil
}

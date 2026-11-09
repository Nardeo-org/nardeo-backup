package backup

import (
	"compress/gzip"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nardeo/nardeo-backup/internal/config"
)

func installDump(t *testing.T, script string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "pg_dump"), []byte("#!/bin/sh\n"+script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
}

func testConfig(t *testing.T) config.Config {
	t.Helper()
	return config.Config{DatabaseURL: "postgresql://backup:top-secret@localhost/app?sslmode=require", OutputDir: filepath.Join(t.TempDir(), "backups")}
}

func TestPostgresStreamsGzipAndProtectsCredentials(t *testing.T) {
	installDump(t, `case "$*" in *top-secret*) exit 20;; esac
test "$PGPASSWORD" = 'top-secret' || exit 21
test "$PGHOST" = '' || exit 22
test "$PGDATABASE" = '' || exit 23
test "$PGSERVICE" = '' || exit 24
printf 'CREATE TABLE example (id integer);\n'
`)
	t.Setenv("PGHOST", "wrong-host")
	t.Setenv("PGDATABASE", "wrong-db")
	t.Setenv("PGSERVICE", "wrong-service")
	cfg := testConfig(t)
	first, err := Postgres(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Postgres(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if first == second || !strings.HasSuffix(first, ".sql.gz") {
		t.Fatalf("invalid backup filenames: %q %q", first, second)
	}
	file, err := os.Open(first)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	content, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "CREATE TABLE example (id integer);\n" {
		t.Fatalf("unexpected dump %q", content)
	}
	info, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("backup permissions %o", info.Mode().Perm())
	}
	entries, err := os.ReadDir(cfg.OutputDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("unexpected files: %v", entries)
	}
}

func TestPostgresUsesSeparateParameters(t *testing.T) {
	installDump(t, `test "$PGHOST" = 'db.internal' || exit 20
test "$PGPORT" = '5433' || exit 21
test "$PGUSER" = 'backup' || exit 22
test "$PGPASSWORD" = 'secret' || exit 23
test "$PGDATABASE" = 'app' || exit 24
case "$*" in *--no-password*) ;; *) exit 25;; esac
printf 'SELECT 1;\n'
`)
	cfg := config.Config{Host: "db.internal", Port: "5433", User: "backup", Password: "secret", Database: "app", OutputDir: t.TempDir()}
	if _, err := Postgres(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresFailureRemovesPartialAndRedactsSecrets(t *testing.T) {
	installDump(t, "printf 'partial SQL'\nprintf 'connection failed: top-secret' >&2\nexit 1\n")
	cfg := testConfig(t)
	_, err := Postgres(context.Background(), cfg)
	if err == nil || !strings.Contains(err.Error(), "connection failed") {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(err.Error(), "top-secret") {
		t.Fatal("password leaked")
	}
	entries, readErr := os.ReadDir(cfg.OutputDir)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("partial files remain: %v", entries)
	}
}

func TestPostgresMissingBinary(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	_, err := Postgres(context.Background(), testConfig(t))
	if err == nil || !strings.Contains(err.Error(), "pg_dump") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPostgresCancellation(t *testing.T) {
	installDump(t, "while :; do :; done\n")
	cfg := testConfig(t)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err := Postgres(ctx, cfg)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline error: %v", err)
	}
	entries, readErr := os.ReadDir(cfg.OutputDir)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if len(entries) != 0 {
		t.Fatal("partial files remain")
	}
}

func TestPostgresOutputDirectoryFailure(t *testing.T) {
	installDump(t, "printf 'SELECT 1;'\n")
	cfg := testConfig(t)
	if err := os.WriteFile(cfg.OutputDir, []byte("occupied"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Postgres(context.Background(), cfg); err == nil {
		t.Fatal("expected output directory error")
	}
}

func TestPostgresURLPreservesPlusInQueryPassword(t *testing.T) {
	installDump(t, `test "$PGPASSWORD" = 'a+b' || exit 20
case "$*" in *a+b*|*a%2Bb*) exit 21;; esac
printf 'SELECT 1;'
`)
	cfg := testConfig(t)
	cfg.DatabaseURL = "postgresql://backup@localhost/app?password=a+b"
	if _, err := Postgres(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresSocketURLPreservesEncodedHost(t *testing.T) {
	installDump(t, `case "$*" in *--dbname=postgresql://backup@%2Fvar%2Frun%2Fpostgresql/app*) ;; *) exit 20;; esac
printf 'SELECT 1;'
`)
	cfg := testConfig(t)
	cfg.DatabaseURL = "postgresql://backup@%2Fvar%2Frun%2Fpostgresql/app"
	if _, err := Postgres(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresURLPreservesEncodedQueryOptions(t *testing.T) {
	installDump(t, `case "$*" in *options=-c%20search_path%3Dpublic*) ;; *) exit 20;; esac
printf 'SELECT 1;'
`)
	cfg := testConfig(t)
	cfg.DatabaseURL = "postgresql://backup@localhost/app?options=-c%20search_path%3Dpublic"
	if _, err := Postgres(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
}

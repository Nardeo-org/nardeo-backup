//go:build integration

// Copyright (c) 2026 Nardeo.

package integration

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestPostgresBackupRestore(t *testing.T) {
	if os.Getenv("NARDEO_INTEGRATION") != "1" {
		t.Fatal("run this suite through bash tests/integration/run.sh to provision an isolated PostgreSQL container")
	}
	binary := os.Getenv("NARDEO_BACKUP_BINARY")
	password := os.Getenv("NARDEO_INTEGRATION_PASSWORD")
	if binary == "" || password == "" {
		t.Fatal("the integration runner must provide a CLI binary and temporary database credentials")
	}
	ctx, cancel := context.WithCancel(context.Background())
	if deadline, ok := t.Deadline(); ok {
		cancel()
		ctx, cancel = context.WithDeadline(context.Background(), deadline)
	}
	t.Cleanup(cancel)
	env := []string{
		"PATH=" + os.Getenv("PATH"),
		"PGHOST=127.0.0.1", "PGPORT=5432", "PGUSER=nardeo_test",
		"PGPASSWORD=" + password, "PGDATABASE=postgres",
		"PGSSLMODE=disable", "PGCONNECT_TIMEOUT=5", "PGCLIENTENCODING=UTF8",
	}
	createDatabase(t, ctx, env, "nardeo_source", password)
	psql(t, ctx, env, "nardeo_source", strings.NewReader(`
CREATE SCHEMA backup_fixture;
CREATE TABLE backup_fixture.records (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    label text NOT NULL,
    note text,
    active boolean NOT NULL
);
INSERT INTO backup_fixture.records (label, note, active) VALUES
    ('Nardeo''s sample', E'first line\nsecond line', true),
    ('Zażółć gęślą jaźń', NULL, false),
    ('Comma, quote "', E'tab\tvalue', true);
`), password)
	want := []fixtureRecord{
		{ID: 1, Label: "Nardeo's sample", Note: "first line\nsecond line", Active: true},
		{ID: 2, Label: "Zażółć gęślą jaźń", Note: nil, Active: false},
		{ID: 3, Label: `Comma, quote "`, Note: "tab\tvalue", Active: true},
	}
	assertRecords(t, ctx, env, "nardeo_source", password, want)

	for _, mode := range []string{"environment", "database_url"} {
		t.Run(mode, func(t *testing.T) {
			destination := "nardeo_restore_" + mode
			createDatabase(t, ctx, env, destination, password)
			output := filepath.Join(t.TempDir(), "backups")
			backupEnv := append([]string(nil), env...)
			if mode == "database_url" {
				connection := &url.URL{
					Scheme: "postgresql", Host: "127.0.0.1:5432", Path: "/nardeo_source",
					User: url.UserPassword("nardeo_test", password), RawQuery: "sslmode=disable",
				}
				backupEnv = append(backupEnv, "DATABASE_URL="+connection.String())
			} else {
				for i, entry := range backupEnv {
					if strings.HasPrefix(entry, "PGDATABASE=") {
						backupEnv[i] = "PGDATABASE=nardeo_source"
					}
				}
			}
			run(t, ctx, backupEnv, nil, password, binary, "pg", "--output-dir", output)
			entries, err := os.ReadDir(output)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 1 || !strings.HasSuffix(entries[0].Name(), ".sql.gz") {
				t.Fatalf("expected exactly one completed gzip backup, got %v", entries)
			}
			file, err := os.Open(filepath.Join(output, entries[0].Name()))
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			info, err := file.Stat()
			if err != nil {
				t.Fatal(err)
			}
			if info.Mode().Perm() != 0600 {
				t.Fatalf("backup permissions = %o, want 600", info.Mode().Perm())
			}
			compressed, err := gzip.NewReader(file)
			if err != nil {
				t.Fatalf("open gzip backup: %v", err)
			}
			defer compressed.Close()
			psql(t, ctx, env, destination, compressed, password)
			assertRecords(t, ctx, env, destination, password, want)
			nextID := psql(t, ctx, env, destination, strings.NewReader(
				"INSERT INTO backup_fixture.records (label, active) VALUES ('after restore', true) RETURNING id;"), password)
			if nextID != "4" {
				t.Fatalf("restored identity sequence returned %q, want 4", nextID)
			}
		})
	}
}

type fixtureRecord struct {
	ID     int    `json:"id"`
	Label  string `json:"label"`
	Note   any    `json:"note"`
	Active bool   `json:"active"`
}

func createDatabase(t *testing.T, ctx context.Context, env []string, database, password string) {
	t.Helper()
	run(t, ctx, env, nil, password, "createdb", "--no-password", database)
	t.Cleanup(func() {
		run(t, ctx, env, nil, password, "dropdb", "--no-password", "--if-exists", "--force", database)
	})
}

func assertRecords(t *testing.T, ctx context.Context, env []string, database, password string, want []fixtureRecord) {
	t.Helper()
	data := psql(t, ctx, env, database, strings.NewReader(
		"SELECT json_agg(records ORDER BY id) FROM backup_fixture.records AS records;"), password)
	var got []fixtureRecord
	if err := json.Unmarshal([]byte(data), &got); err != nil {
		t.Fatalf("decode fixture records: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("records differ:\ngot  %#v\nwant %#v", got, want)
	}
}

func psql(t *testing.T, ctx context.Context, env []string, database string, input io.Reader, password string) string {
	t.Helper()
	return run(t, ctx, env, input, password, "psql", "--no-psqlrc", "--no-password", "--quiet",
		"--tuples-only", "--no-align", "--set=ON_ERROR_STOP=1", "--dbname="+database)
}

func run(t *testing.T, ctx context.Context, env []string, input io.Reader, password, program string, args ...string) string {
	t.Helper()
	cmd := exec.CommandContext(ctx, program, args...)
	cmd.Env, cmd.Stdin = env, input
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		detail := stderr.String()
		for _, secret := range []string{password, url.QueryEscape(password), url.PathEscape(password)} {
			detail = strings.ReplaceAll(detail, secret, "[REDACTED]")
		}
		t.Fatalf("%s failed: %v\n%s", filepath.Base(program), err, detail)
	}
	return strings.TrimSpace(stdout.String())
}

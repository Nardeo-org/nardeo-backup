package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRootAndHelpShowBannerWithoutSecrets(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgresql://backup:help-secret@localhost/app")
	for _, args := range [][]string{nil, {"--help"}, {"help"}, {"pg", "--help"}} {
		var out, logs bytes.Buffer
		root := newRootCommand()
		root.SetOut(&out)
		root.SetErr(&logs)
		root.SetArgs(args)
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), "NARDEO BACKUP") || !strings.Contains(out.String(), "--output-dir") {
			t.Fatalf("missing branded help: %s", &out)
		}
		if strings.Contains(out.String(), "help-secret") {
			t.Fatal("help leaked credentials")
		}
	}
}

func TestVersionReportsBuildMetadataWithoutDatabase(t *testing.T) {
	previousVersion, previousCommit := version, commit
	version, commit = "v0.1.0", "abc123def456"
	t.Cleanup(func() { version, commit = previousVersion, previousCommit })
	t.Setenv("DATABASE_URL", "invalid://user:version-secret@localhost/app")
	var out, logs bytes.Buffer
	root := newRootCommand()
	root.SetOut(&out)
	root.SetErr(&logs)
	root.SetArgs([]string{"--version"})
	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	if out.String() != "nardeo-backup v0.1.0 (commit abc123def456)\n" {
		t.Fatalf("unexpected version output: %q", out.String())
	}
	if logs.Len() != 0 {
		t.Fatalf("unexpected version diagnostics: %s", &logs)
	}
}

func TestPGCommandTimeoutRemovesPartialBackup(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "pg_dump"), []byte("#!/bin/sh\nprintf 'partial dump'\nwhile :; do :; done\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	output := filepath.Join(t.TempDir(), "backups")
	var out, logs bytes.Buffer
	root := newRootCommand()
	root.SetOut(&out)
	root.SetErr(&logs)
	root.SetArgs([]string{"pg", "--db-url", "postgresql://backup@localhost/app", "--output-dir", output, "--timeout", "50ms"})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	started := time.Now()
	err := root.ExecuteContext(ctx)
	if !errors.Is(err, context.DeadlineExceeded) || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout error, got %v", err)
	}
	if time.Since(started) >= 2*time.Second {
		t.Fatal("the flag deadline did not interrupt the dump before the caller deadline")
	}
	entries, readErr := os.ReadDir(output)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("files remain after timeout: %v", entries)
	}
	if strings.Contains(logs.String(), "Backup completed") {
		t.Fatal("timed-out backup was reported as completed")
	}
}

func TestPGCommandPreservesCallerCancellation(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "pg_dump"), []byte("#!/bin/sh\nprintf 'SELECT 1;'\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	root := newRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"pg", "--db-url", "postgresql://backup@localhost/app", "--output-dir", t.TempDir(), "--timeout", "1h"})
	err := root.ExecuteContext(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected caller cancellation, got %v", err)
	}
	if strings.Contains(err.Error(), "timed out") {
		t.Fatal("caller cancellation was reported as a timeout")
	}
}

func TestPGCommandWritesBackupAndStructuredLogs(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "pg_dump"), []byte("#!/bin/sh\nprintf 'SELECT 1;\\n'\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	t.Setenv("DATABASE_URL", "")
	output := t.TempDir()
	var out, logs bytes.Buffer
	root := newRootCommand()
	root.SetOut(&out)
	root.SetErr(&logs)
	root.SetArgs([]string{"pg", "--db-url", "postgresql://backup:cli-secret@localhost/app", "--output-dir", output, "--verbose"})
	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(output)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || !strings.HasSuffix(entries[0].Name(), ".sql.gz") {
		t.Fatalf("unexpected outputs: %v", entries)
	}
	if !strings.Contains(logs.String(), "level=INFO") || !strings.Contains(logs.String(), "level=DEBUG") || !strings.Contains(logs.String(), "path=") {
		t.Fatalf("missing structured logs: %s", &logs)
	}
	if strings.Contains(logs.String(), "cli-secret") {
		t.Fatal("log leaked password")
	}
}

func TestPGCommandRejectsInvalidURL(t *testing.T) {
	root := newRootCommand()
	root.SetArgs([]string{"pg", "--db-url", "https://user:secret@host/db"})
	err := root.ExecuteContext(context.Background())
	if err == nil {
		t.Fatal("expected invalid URL error")
	}
	if strings.Contains(err.Error(), "secret") {
		t.Fatal("error leaked password")
	}
}

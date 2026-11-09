package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

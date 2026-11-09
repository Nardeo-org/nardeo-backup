package config

import (
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

func TestFlagsOverrideEnvironment(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgresql://env:secret@localhost/envdb")
	t.Setenv("OUTPUT_DIR", "env-backups")
	cfg := FromEnv()
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	cfg.BindFlags(flags)
	if err := flags.Parse([]string{"--db-url", "postgresql://cli:secret@localhost/app", "--output-dir", "custom", "--verbose"}); err != nil {
		t.Fatal(err)
	}
	if cfg.DatabaseURL != "postgresql://cli:secret@localhost/app" || cfg.OutputDir != "custom" || !cfg.Verbose {
		t.Fatalf("flags did not override environment")
	}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestEnvironmentDefaults(t *testing.T) {
	for _, key := range []string{"DATABASE_URL", "OUTPUT_DIR", "PGHOST", "PGPORT", "PGUSER", "PGPASSWORD", "PGDATABASE"} {
		t.Setenv(key, "")
	}
	cfg := FromEnv()
	if cfg.Host != "localhost" || cfg.Port != "5432" || cfg.OutputDir != "./backups" {
		t.Fatalf("unexpected defaults: host=%q port=%q output=%q", cfg.Host, cfg.Port, cfg.OutputDir)
	}
	t.Setenv("PGUSER", "backup")
	t.Setenv("PGDATABASE", "app")
	t.Setenv("PGPASSWORD", "secret")
	cfg = FromEnv()
	if cfg.User != "backup" || cfg.Database != "app" || cfg.Password != "secret" {
		t.Fatal("PG environment not loaded")
	}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestValidation(t *testing.T) {
	for _, tc := range []struct {
		name string
		cfg  Config
		want string
	}{
		{"missing connection", Config{OutputDir: "backups"}, "PGHOST"},
		{"bad port", Config{OutputDir: "backups", Host: "localhost", Port: "65536", User: "backup", Database: "app"}, "port"},
		{"bad URL", Config{OutputDir: "backups", DatabaseURL: "https://user:secret@localhost/app"}, "PostgreSQL"},
		{"empty output", Config{DatabaseURL: "postgresql://localhost/app"}, "output"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q error, got %v", tc.want, err)
			}
			if strings.Contains(err.Error(), "secret") {
				t.Fatal("secret leaked")
			}
		})
	}
}

func TestValidationAcceptsPostgresSocketURL(t *testing.T) {
	cfg := Config{OutputDir: "backups", DatabaseURL: "postgresql://backup@%2Fvar%2Frun%2Fpostgresql/app"}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
}

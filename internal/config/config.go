package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

type Config struct {
	Host        string
	Port        string
	User        string
	Password    string
	Database    string
	DatabaseURL string
	OutputDir   string
	Verbose     bool
	Timeout     time.Duration
}

func FromEnv() Config {
	return Config{
		Host:        envDefault("PGHOST", "localhost"),
		Port:        envDefault("PGPORT", "5432"),
		User:        os.Getenv("PGUSER"),
		Password:    os.Getenv("PGPASSWORD"),
		Database:    os.Getenv("PGDATABASE"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		OutputDir:   envDefault("OUTPUT_DIR", "./backups"),
	}
}

func (c *Config) BindFlags(flags *pflag.FlagSet) {
	flags.StringVar(&c.OutputDir, "output-dir", c.OutputDir, "Directory for compressed backups (OUTPUT_DIR)")
	flags.StringVar(&c.DatabaseURL, "db-url", c.DatabaseURL, "PostgreSQL connection URL (DATABASE_URL)")
	// Connection URLs may contain credentials; never print an environment-derived default.
	flags.Lookup("db-url").DefValue = ""
	flags.BoolVar(&c.Verbose, "verbose", c.Verbose, "Enable debug logging")
	flags.DurationVar(&c.Timeout, "timeout", c.Timeout, "Maximum backup duration, for example 15m (0 disables the limit)")
}

func (c Config) Validate() error {
	if c.Timeout < 0 {
		return fmt.Errorf("timeout must be zero or a positive duration")
	}
	if strings.TrimSpace(c.OutputDir) == "" {
		return fmt.Errorf("output directory must not be empty")
	}
	if c.DatabaseURL != "" {
		_, err := ParseDatabaseURL(c.DatabaseURL)
		return err
	}
	for _, field := range []struct{ name, value string }{
		{"PGHOST", c.Host}, {"PGUSER", c.User}, {"PGDATABASE", c.Database},
	} {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("%s is required when DATABASE_URL or --db-url is not set", field.name)
		}
	}
	port, err := strconv.Atoi(c.Port)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("PostgreSQL port must be an integer between 1 and 65535")
	}
	return nil
}

func envDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

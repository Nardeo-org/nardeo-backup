package main

import (
	"time"

	"github.com/nardeo/nardeo-backup/internal/backup"
	"github.com/nardeo/nardeo-backup/internal/config"
	"github.com/nardeo/nardeo-backup/internal/ui"
	"github.com/spf13/cobra"
)

func newRootCommand() *cobra.Command {
	cfg := config.FromEnv()
	root := &cobra.Command{
		Use:           "nardeo-backup",
		Short:         "Stream PostgreSQL backups into compressed files",
		Long:          "Nardeo Backup creates timestamped, gzip-compressed PostgreSQL SQL dumps.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE:          func(cmd *cobra.Command, args []string) error { return cmd.Help() },
	}
	root.CompletionOptions.DisableDefaultCmd = true
	cfg.BindFlags(root.PersistentFlags())
	defaultHelp := root.HelpFunc()
	root.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		ui.PrintBanner(cmd.OutOrStdout())
		defaultHelp(cmd, args)
	})
	root.AddCommand(&cobra.Command{
		Use:     "pg",
		Short:   "Back up a PostgreSQL database with pg_dump",
		Example: "  nardeo-backup pg --output-dir ./backups\n  nardeo-backup pg --db-url 'postgresql://backup@localhost/app' --verbose",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.Validate(); err != nil {
				return err
			}
			logger := ui.NewLogger(cmd.ErrOrStderr(), cfg.Verbose)
			logger.Info("Starting PostgreSQL backup")
			logger.Debug("Streaming pg_dump output through gzip")
			started := time.Now()
			path, err := backup.Postgres(cmd.Context(), cfg)
			if err != nil {
				return err
			}
			logger.Info("Backup completed", "path", path, "duration", time.Since(started))
			return nil
		},
	})
	return root
}

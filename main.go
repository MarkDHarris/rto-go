package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"rto/cmd"
	"rto/data"
)

var dataDir string

var rootCmd = &cobra.Command{
	Use:   "rto",
	Short: "Return-to-Office tracker",
	Long:  `rto tracks your office badge-in days and calculates compliance with RTO policies.`,
	PersistentPreRunE: func(c *cobra.Command, args []string) error {
		if dataDir != "" {
			data.SetDataDir(dataDir)
		}
		// Auto-init if data directory is empty or missing
		if c.Use == "init" {
			return nil // let init handle it
		}
		if dirNeedsInit(data.GetDataDir()) {
			fmt.Fprintf(os.Stderr, "Data directory not initialized. Running 'rto init'...\n")
			if err := cmd.RunInitInDir(data.GetDataDir()); err != nil {
				return fmt.Errorf("auto-init failed: %w", err)
			}
		}
		return nil
	},
	RunE: func(c *cobra.Command, args []string) error {
		return cmd.RunBubbleteaTUI()
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize data files with defaults",
	RunE: func(c *cobra.Command, args []string) error {
		return cmd.RunInit()
	},
}

var statsCmd = &cobra.Command{
	Use:   "stats [PERIOD_KEY]",
	Short: "Print statistics for a time period",
	Long:  `Print statistics for a time period (e.g., Q1_2025). Uses the current period if not specified.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		key := ""
		if len(args) > 0 {
			key = args[0]
		} else {
			td, err := data.LoadTimePeriodData()
			if err != nil {
				return err
			}
			tp, err := td.GetCurrentPeriod()
			if err != nil {
				return fmt.Errorf("cannot determine current period: %w (try specifying a period key)", err)
			}
			key = tp.Key
		}
		return cmd.RunStats(key)
	},
}

var vacationsCmd = &cobra.Command{
	Use:   "vacations",
	Short: "List all vacations",
	RunE: func(c *cobra.Command, args []string) error {
		return cmd.RunVacations()
	},
}

var holidaysCmd = &cobra.Command{
	Use:   "holidays",
	Short: "List all holidays",
	RunE: func(c *cobra.Command, args []string) error {
		return cmd.RunHolidays()
	},
}

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup data directory to git",
	RunE: func(c *cobra.Command, args []string) error {
		remote, _ := c.Flags().GetString("remote")
		dir, _ := c.Flags().GetString("dir")
		if dir == "" {
			dir = data.GetDataDir()
		}
		result := cmd.PerformGitBackup(dir, remote)
		if result.IsError {
			return fmt.Errorf("%s", result.Message)
		}
		fmt.Println(result.Message)
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&dataDir, "data-dir", "d", "", "Data directory (default: ./config)")

	backupCmd.Flags().StringP("remote", "r", "", "Git remote URL")
	backupCmd.Flags().StringP("dir", "", "", "Directory to backup (default: data-dir)")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(vacationsCmd)
	rootCmd.AddCommand(holidaysCmd)
	rootCmd.AddCommand(backupCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// dirNeedsInit returns true only if the data directory has never been initialized.
// It checks for settings.yaml as the canonical marker of initialization.
func dirNeedsInit(dir string) bool {
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return true
	}
	if !info.IsDir() {
		return true
	}
	settingsPath := filepath.Join(dir, "settings.yaml")
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return true
	}
	return false
}

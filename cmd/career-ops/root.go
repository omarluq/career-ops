package main

import (
	"errors"
	"fmt"
	"os"

	_ "github.com/charmbracelet/log"
	"github.com/samber/lo"
	_ "github.com/samber/mo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/omarluq/career-ops/internal/vinfo"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:     "career-ops",
	Short:   "AI job search pipeline CLI",
	Long: "career-ops automates pipeline tracking, offer evaluation, " +
		"CV generation, portal scanning, and batch processing.",
	Version: vinfo.String(),
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: config/profile.yml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().String("db", "career-ops.db", "path to SQLite database")
	rootCmd.PersistentFlags().Bool("legacy", false, "use legacy markdown parsing instead of SQLite")

	lo.ForEach([]string{"verbose", "db", "legacy"}, func(flag string, _ int) {
		if err := viper.BindPFlag(flag, rootCmd.PersistentFlags().Lookup(flag)); err != nil {
			fmt.Fprintf(os.Stderr, "warning: binding %s flag: %v\n", flag, err)
		}
	})

	rootCmd.AddCommand(
		verifyCmd,
		normalizeCmd,
		dedupCmd,
		mergeCmd,
		syncCheckCmd,
		pdfCmd,
		pdfBatchCmd,
		batchCmd,
		dashboardCmd,
		importCmd,
		exportCmd,
		scanCmd,
		mcpCmd,
		profileCmd,
	)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("profile")
		viper.SetConfigType("yml")
		viper.AddConfigPath("config")
	}

	viper.SetEnvPrefix("CAREER_OPS")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			fmt.Fprintf(os.Stderr, "warning: config error: %v\n", err)
		}
	}
}

package main

import (
	"fmt"
	"os"

	_ "github.com/charmbracelet/log"
	_ "github.com/samber/lo"
	_ "github.com/samber/mo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/omarluq/career-ops/internal/vinfo"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "career-ops",
	Short: "AI job search pipeline CLI",
	Long:    "career-ops automates pipeline tracking, offer evaluation, CV generation, portal scanning, and batch processing.",
	Version: vinfo.String(),
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: config/profile.yml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable verbose output")

	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	rootCmd.AddCommand(
		verifyCmd,
		normalizeCmd,
		dedupCmd,
		mergeCmd,
		syncCheckCmd,
		pdfCmd,
		batchCmd,
		dashboardCmd,
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
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "warning: config error: %v\n", err)
		}
	}
}

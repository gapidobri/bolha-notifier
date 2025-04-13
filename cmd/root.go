package cmd

import (
	"github.com/gapidobri/bolha-notifier/internal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "bolha-notifier",
	Short: "Watcher for new bolha.com posts",
	Run: func(cmd *cobra.Command, args []string) {
		internal.Run()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "configuration file path")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/bolha-notifier")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.SetDefault("CheckInterval", 300)

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err == nil {
		log.Info("using config file: ", viper.ConfigFileUsed())
	}
}

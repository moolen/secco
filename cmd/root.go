package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "secco",
	Short: "",
	Long:  ``,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		lvl, err := log.ParseLevel(viper.GetString("loglevel"))
		if err != nil {
			log.Fatal(err)
		}
		log.SetLevel(lvl)
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	viper.AutomaticEnv()
	flags := rootCmd.PersistentFlags()
	// TODO: add k8s integration
	flags.String("kubeconfig", "", "kubeconfig to use")
	flags.String("loglevel", "debug", "set the loglevel")
	viper.BindPFlags(flags)
	viper.BindEnv("loglevel", "LOGLEVEL")
	viper.BindEnv("kubeconfig", "KUBECONFIG")
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

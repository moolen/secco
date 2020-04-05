package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/moolen/secco/pkg/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	flags := clientCmd.PersistentFlags()
	flags.String("target", "dns:///localhost:3000", "specify the grpc server to ask for traces. you may specify a dns+srv based discovery")
	viper.BindPFlags(flags)
	viper.BindEnv("target", "TARGET_ADDR")
	rootCmd.AddCommand(clientCmd)
}

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "client communicates with the agent",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := client.New(
			viper.GetString("target"),
		)
		if err != nil {
			log.Fatal(err)
		}
		stopChan := make(chan os.Signal)
		signal.Notify(stopChan, os.Interrupt)
		signal.Notify(stopChan, syscall.SIGTERM)

		go func() {
			<-stopChan
		}()

		log.Infof("%v", client)
	},
}

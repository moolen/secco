package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/moolen/secco/pkg/agent"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	flags := agentCmd.PersistentFlags()
	//	flags.String("k8s-node", "", "kubernetes node name")
	viper.BindPFlags(flags)
	//	viper.BindEnv("k8s-node", "KUBERNETES_NODE")
	rootCmd.AddCommand(agentCmd)
}

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "..",
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("starting agent")

		srv, err := agent.NewAgentServer()
		if err != nil {
			log.Fatal(err)
		}

		stopChan := make(chan os.Signal)
		signal.Notify(stopChan, os.Interrupt)
		signal.Notify(stopChan, syscall.SIGTERM)
		go func() {
			<-stopChan
			log.Infof("received ctrl+c, cleaning up")
			log.Infof("shutting down")
			os.Exit(0)
		}()

		go srv.Serve(context.Background())
		// start metrics server
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	},
}

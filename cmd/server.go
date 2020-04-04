package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/moolen/secco/pkg/server"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	flags := serverCmd.PersistentFlags()
	flags.String("target", "dns:///localhost:3000", "specify the grpc server to ask for traces. you may specify a dns+srv based discovery")
	flags.Int("listen", 3001, "specify the port to listen on")
	flags.Duration("sync-interval", time.Second*60, "sync intervall for k8s resources")
	flags.Int("cache-buffer-size", 3000, "cache buffer size")
	viper.BindPFlags(flags)
	viper.BindEnv("target", "TARGET_ADDR")
	viper.BindEnv("listen", "LISTEN")
	viper.BindEnv("sync-interval", "SYNC_INTERVAL")
	viper.BindEnv("cache-buffer-size", "CACHE_BUFFER_SIZE")
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Server collects traces from agent nodes",
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("starting server")
		// kubeClient, err := newClient()
		// if err != nil {
		// 	log.Fatal(err)
		// }
		srv, err := server.New(
			viper.GetString("target"),
			viper.GetInt("listen"),
			viper.GetDuration("sync-interval"),
			viper.GetInt("cache-buffer-size"),
		)
		if err != nil {
			log.Fatal(err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		stopChan := make(chan os.Signal)
		signal.Notify(stopChan, os.Interrupt)
		signal.Notify(stopChan, syscall.SIGTERM)

		go func() {
			<-stopChan
			log.Infof("stopping server")
			srv.Stop()
			cancel()
		}()

		srv.Serve(ctx)
	},
}

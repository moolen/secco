package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/moolen/secco/pkg/client"
	pb "github.com/moolen/secco/proto"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	flags := syncProfileCmd.PersistentFlags()
	flags.Duration("duration", time.Second*10, "specify the trace duration")
	flags.String("id", "", "specify container to trace")
	viper.BindPFlags(flags)
	viper.BindEnv("duration", "TRACE_DURATION")
	viper.BindEnv("id", "CONTAINER_ID")
	clientCmd.AddCommand(syncProfileCmd)
}

var syncProfileCmd = &cobra.Command{
	Use:   "sync-profile",
	Short: "sync profile",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := client.New(
			viper.GetString("target"),
		)
		if err != nil {
			log.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("duration"))
		stopChan := make(chan os.Signal)
		signal.Notify(stopChan, os.Interrupt)
		signal.Notify(stopChan, syscall.SIGTERM)

		go func() {
			<-stopChan
			cancel()
		}()
		// TODO: parse profile
		res, err := client.SyncProfile(ctx, &pb.SyncProfileRequest{})
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("res: %v", res)
	},
}

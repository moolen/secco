package cmd

import (
	"context"
	"io"
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
	flags := clientTraceCmd.PersistentFlags()
	flags.Duration("duration", time.Second*10, "specify the trace duration")
	flags.String("id", "", "specify container to trace")
	viper.BindPFlags(flags)
	viper.BindEnv("duration", "TRACE_DURATION")
	viper.BindEnv("id", "CONTAINER_ID")
	clientCmd.AddCommand(clientTraceCmd)
}

var clientTraceCmd = &cobra.Command{
	Use:   "trace",
	Short: "trace a specific container",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := client.New(
			viper.GetString("target"),
		)
		if err != nil {
			log.Fatal(err)
		}
		dur := viper.GetDuration("duration")
		ctx, cancel := context.WithTimeout(context.Background(), dur+time.Second*10)
		stopChan := make(chan os.Signal)
		signal.Notify(stopChan, os.Interrupt)
		signal.Notify(stopChan, syscall.SIGTERM)

		go func() {
			<-stopChan
			cancel()
		}()

		rc, err := client.RunTrace(ctx, &pb.RunTraceRequest{
			Id:       viper.GetString("id"),
			Duration: dur.Nanoseconds(),
		})

		for {
			res, err := rc.Recv()
			if err != nil {
				if err != io.EOF {
					log.Error(err)
				}
				break
			}
			log.Infof("%v", res)
		}
	},
}

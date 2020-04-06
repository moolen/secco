package cmd

import (
	"context"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/moolen/secco/pkg/client"
	pb "github.com/moolen/secco/proto"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	flags := syncProfileCmd.PersistentFlags()
	flags.String("profile", "", "path to the profile")
	flags.Duration("timeout", time.Second*10, "sync timeout")
	viper.BindPFlags(flags)
	viper.BindEnv("profile", "PROFILE_PATH")
	viper.BindEnv("timeout", "TIMEOUT")
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
		ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("timeout"))
		stopChan := make(chan os.Signal)
		signal.Notify(stopChan, os.Interrupt)
		signal.Notify(stopChan, syscall.SIGTERM)

		go func() {
			<-stopChan
			cancel()
		}()
		profile, err := ioutil.ReadFile(viper.GetString("profile"))
		if err != nil {
			log.Fatal(err)
		}
		res, err := client.SyncProfile(ctx, &pb.SyncProfileRequest{
			Profiles: []*pb.SeccompProfile{
				{
					Id:      uuid.New().String(),
					Name:    path.Base(viper.GetString("profile")),
					Profile: profile,
				},
			},
		})
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("res: %v", res)
	},
}

package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/moolen/secco/pkg/tracer"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vishvananda/netns"
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
		id := viper.GetString("id")
		if id == "" {
			log.Fatalf("id can not be empty")
		}
		ns, err := netns.GetFromDocker(id)
		if err != nil {
			log.Fatal(err)
		}
		var s syscall.Stat_t
		err = syscall.Fstat(int(ns), &s)
		if err != nil {
			log.Fatalf("could not fstat ns: %s", err)
		}
		netnsid := ns.UniqueId()
		log.Infof("found netns: %s", netnsid)

		stop := make(chan struct{})
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			log.Infof("waiting for ctrl-c")
			for range c {
				log.Infof("got for ctrl-c, stopping trace")
				stop <- struct{}{}
				return
			}
		}()

		calls, err := tracer.Start(s.Ino, stop)
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("calls: %v", calls)
	},
}

func init() {
	viper.AutomaticEnv()
	flags := rootCmd.PersistentFlags()
	// TODO: add k8s integration
	//flags.String("kubeconfig", "", "kubeconfig to use")
	flags.String("loglevel", "debug", "set the loglevel")
	flags.String("id", "", "specify the container id")
	viper.BindPFlags(flags)
	viper.BindEnv("loglevel", "LOGLEVEL")
	//viper.BindEnv("kubeconfig", "KUBECONFIG")
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

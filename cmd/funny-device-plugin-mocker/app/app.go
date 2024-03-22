package app

import (
	"fmt"
	"funny-device-plugin-mocker/cmd/funny-device-plugin-mocker/app/options"

	"funny-device-plugin-mocker/pkg/log"
	"funny-device-plugin-mocker/pkg/server"

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func NewCommand(version string) *cobra.Command {
	opts := options.NewOptions()

	rootCmd := &cobra.Command{
		Use: "funny-device-plugin-mocker",
	}

	runCmd := &cobra.Command{
		Use:     "run",
		Version: version,
		Short:   "run grpc server for a mocker device plugin server for kubernetes",
		Long:    "run grpc server for a mocker device plugin server for kubernetes",
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			runServer(opts, signals.SetupSignalHandler().Done())
		},
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "print the version of this command",
		Long:  "print the version of this command",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: '%s'", version)
		},
	}

	opts.BindFlags(runCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(versionCmd)
	return rootCmd
}

func runServer(opts *options.Option, stopChan <-chan struct{}) {
	c := make(chan struct{}, 1)

	go func() {
		<-stopChan
		log.Logger.Info("Received termination, signaling shutdown.")
		close(c)
	}()

	log.Logger.Info("Attempt to init server")

	servcie, err := server.NewDevicePlugin(opts.RootPath, opts.Prefix, opts.ResourceName)
	if err != nil {
		panic(err)
	}

	servcie.Serve()
	log.Logger.Info("Server started")

	<-c

}

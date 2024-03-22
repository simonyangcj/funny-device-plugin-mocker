package main

import (
	app "funny-device-plugin-mocker/cmd/funny-device-plugin-mocker/app"
	"funny-device-plugin-mocker/pkg/log"
	"os"
)

var version string

func main() {
	log.InitializeLogger()
	cmd := app.NewCommand(version)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

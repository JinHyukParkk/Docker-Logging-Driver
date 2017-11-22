package main

import (
	"log"

	"github.com/docker/go-plugins-helpers/sdk"
)

func main() {
	// TODO: Write a REAME detailing how to develop, build, configure, & deploy.
	var (
		err error
	)

	pluginName := "delogplugin"

	sdkhandler := sdk.NewHandler(`{"Implements": ["LoggingDriver"]}`)
	driver, err := NewFileDriver()
	if err != nil {
		log.Fatal(err)
	}
	inithandlers(&sdkhandler, driver)

	if err = sdkhandler.ServeUnix(pluginName, 0); err != nil {
		log.Fatal(err)
	}

}

package main

import (
	"log"

	"github.com/docker/go-plugins-helpers/sdk"
)

func main() {
	pluginName := "Testplugin"
	pluginHandler := sdk.NewHandler(`{"Implements": ["LoggingDriver"]}`)
	driver, err := NewFileDriver()
	if err != nil {
		log.Fatal(err)
	}
	inithandlers(&pluginHandler, driver)

	if err = pluginHandler.ServeUnix(pluginName, 0); err != nil {
		log.Fatal(err)
	}

}

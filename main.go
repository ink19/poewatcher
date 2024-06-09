package main

import (
	"flag"
	"os"

	"github.com/ink19/poewatcher/config"
	"github.com/ink19/poewatcher/internal"
	log "github.com/sirupsen/logrus"
)

var configFileName string

func init() {
	flag.StringVar(&configFileName, "config", "config.json", "config file")
}

func main() {
	flag.Parse()
	config.Init(configFileName)
	
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
	internal.RunServer()
}

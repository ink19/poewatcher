package main

import (
	"flag"
	"os"

	"github.com/ink19/poewatcher/config"
	"github.com/ink19/poewatcher/logic/server"
	log "github.com/sirupsen/logrus"

	_ "github.com/mattn/go-sqlite3"
)

var configFileName string

func init() {
	flag.StringVar(&configFileName, "config", "config.yaml", "config file")
}

func main() {
	flag.Parse()
	config.Init(configFileName)

	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	if err := server.New().Run(); err != nil {
		log.Fatal(err)
	}
}

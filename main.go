package main

import (
	"flag"
	"os"
	"os/signal"

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

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	s := server.New()
	go func() {
		<-interrupt
		_ = s.Stop()
	}()

	if err := s.Run(); err != nil {
		log.Fatal(err)
	}
}

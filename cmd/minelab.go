package main

import (
	"github.com/racerxdl/minelab-bot/bot"
	"github.com/racerxdl/minelab-bot/config"
	"github.com/sirupsen/logrus"
	"os"
)

var log = logrus.New()

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("error loading config: %s\n", err)
	}

	lab := bot.MakeMinelab(cfg)
	c := make(chan os.Signal)
	go func() {
		<-c
		log.Infof("Received stop Ctrl+C")
		lab.Stop()
	}()

	err = lab.Start()
	if err != nil {
		log.Fatalf("error starting minelab bot: %s", err)
	}
}
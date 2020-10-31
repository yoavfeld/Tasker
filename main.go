package main

import (
	"log"
	"github.com/yoavfeld/tasker/lib"
)

const configPath = "./"

func main() {
	conf, err := lib.LoadConf(configPath)
	if err != nil {
		log.Fatalf("Failed loading config: %+v", err)
	}
	server := lib.NewServer(conf)
	if err := server.Start(); err != nil {
		log.Fatalf("Error while starting: %v", err)
	}
}
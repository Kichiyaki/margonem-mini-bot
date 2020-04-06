package main

import (
	"bot/margonem"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	conn, err := margonem.Connect(&margonem.Config{
		Username: "",
		Password: "",
		Proxy:    "",
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := conn.UseWholeStamina("", ""); err != nil {
		log.Fatal(err)
	}

	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)
	<-channel

	os.Exit(0)
}

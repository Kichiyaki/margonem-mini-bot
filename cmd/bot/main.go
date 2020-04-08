package main

import (
	"bot/margonem"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/denisbrodbeck/machineid"

	"github.com/robfig/cron/v3"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/sirupsen/logrus"
)

var Mode = "development"
var MachineID = ""

type config struct {
	Accounts []struct {
		Username   string `json:"username"`
		Password   string `json:"password"`
		Proxy      string `json:"proxy"`
		Characters []struct {
			ID    string `json:"id"`
			MapID string `json:"map_id"`
		} `json:"characters"`
	} `json:"accounts"`
	Debug bool `json:"debug"`
}

func main() {
	basePath := "./"
	if Mode == "development" {
		basePath = "./../../"
	}
	logrus.SetOutput(&lumberjack.Logger{
		Filename:   basePath + "logs/bot.log",
		MaxSize:    5, // megabytes
		MaxBackups: 3,
		MaxAge:     1, //days
	})
	if Mode != "development" {
		id, err := machineid.ProtectedID("margonem-mobile-app-bot")
		if err != nil {
			logrus.Fatal(err)
		}
		if id != MachineID {
			logrus.Fatal(fmt.Errorf("Wrong machine id"))
		}
	}
	dat, err := ioutil.ReadFile(basePath + "config.json")
	if err != nil {
		logrus.Fatalf("Cannot load config file: %s", err)
	}
	cfg := &config{}
	err = json.Unmarshal(dat, cfg)
	if err != nil {
		logrus.Fatalf("Cannot unmarshal config file content into config struct: %s", err)
	}

	if cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	c := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.VerbosePrintfLogger(log.New(os.Stdout, "cron: ", log.LstdFlags)))))
	c.AddFunc("* * * * *", func() {
		for _, account := range cfg.Accounts {
			if len(account.Characters) > 0 {
				conn, err := margonem.Connect(&margonem.Config{
					Username: account.Username,
					Password: account.Password,
					Proxy:    account.Proxy,
				})
				if err != nil {
					continue
				}
				for _, char := range account.Characters {
					entry := logrus.WithField("charid", char.ID).WithField("mapid", char.MapID)
					entry.Info("Running cron job")
					time.Sleep(time.Duration(random(200, 400)) * time.Millisecond)
					err := conn.UseWholeStamina(char.ID, char.MapID)
					entry.WithField("err", err).Info("Finished cron job")
				}
			}
		}
	})
	c.Start()
	defer c.Stop()

	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)
	<-channel

	os.Exit(0)
}

func random(min, max int) int {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Intn(max-min) + min
}

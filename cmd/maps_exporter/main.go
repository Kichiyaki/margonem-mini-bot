package main

import (
	"bot/margonem"
	"bot/utils"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/sirupsen/logrus"
)

var Mode = "development"
var MachineID = "*"

type config struct {
	Accounts []struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Proxy    string `json:"proxy"`
	} `json:"accounts"`
	Debug bool `json:"debug"`
}

func main() {
	basePath := "./"
	if Mode == "development" {
		basePath = "./../../"
	}
	logrus.SetOutput(&lumberjack.Logger{
		Filename:   basePath + "logs/map_exporter.log",
		MaxSize:    5, // megabytes
		MaxBackups: 3,
		MaxAge:     1, //days
	})
	if err := utils.CheckMachineID(MachineID); err != nil {
		logrus.Fatal(err)
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

	log.Print("Poczekaj aż program się zamknie...")

	if cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	if len(cfg.Accounts) > 0 {
		conn, err := margonem.Connect(&margonem.Config{
			Username: cfg.Accounts[0].Username,
			Password: cfg.Accounts[0].Password,
			Proxy:    cfg.Accounts[0].Proxy,
		})
		if err != nil {
			logrus.Fatal(err)
		}
		maps, err := conn.Maplist()
		if err != nil {
			logrus.Fatal(err)
		}
		jsonString, _ := json.Marshal(maps)
		ioutil.WriteFile(basePath+"maps.json", jsonString, os.ModePerm)
	}
}

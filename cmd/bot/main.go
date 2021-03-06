package main

import (
	"bot/margonem"
	"bot/utils"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/sirupsen/logrus"
)

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
	logrus.SetOutput(&lumberjack.Logger{
		Filename:   basePath + "logs/bot.log",
		MaxSize:    5, // megabytes
		MaxBackups: 3,
		MaxAge:     1, //days
	})

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
	cJob := &cronJob{
		running: false,
		cfg:     cfg,
	}
	c.AddFunc("* * * * *", cJob.handler)
	c.Start()
	defer c.Stop()
	go cJob.handler()
	log.Print("Uruchomiono bota.")

	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)
	<-channel

	os.Exit(0)
}

type cronJob struct {
	mutex   sync.Mutex
	running bool
	cfg     *config
}

func (cj *cronJob) handler() {
	if cj.running {
		return
	}
	cj.mutex.Lock()
	cj.running = true
	cj.mutex.Unlock()
	defer func() {
		cj.mutex.Lock()
		cj.running = false
		cj.mutex.Unlock()
	}()

	var wg sync.WaitGroup
	limit := runtime.NumCPU() * 10
	count := 0
	for _, account := range cj.cfg.Accounts {
		if count >= limit {
			wg.Wait()
			count = 0
		}
		if len(account.Characters) > 0 {
			wg.Add(1)
			count++
			acc := account
			go func() {
				defer wg.Done()
				conn, err := margonem.Connect(&margonem.Config{
					Username: acc.Username,
					Password: acc.Password,
					Proxy:    acc.Proxy,
				})
				if err != nil {
					return
				}
				for _, char := range acc.Characters {
					entry := logrus.WithField("charid", char.ID).WithField("mapid", char.MapID)
					entry.Info("Performing a cron job")
					time.Sleep(time.Duration(utils.Random(200, 400)) * time.Millisecond)
					err := conn.UseWholeStamina(char.ID, char.MapID)
					entry.WithField("err", err).Info("Performed a cron job")
				}
			}()
		}
	}

	wg.Wait()
}

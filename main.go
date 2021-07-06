package main

import (
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/fsnotify/fsnotify"
	"github.com/robfig/cron/v3"
)

//Version of rockbin
const Version = "v0.1.3"

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func main() {
	bin, mqttClient := config()

	// on launch tell home assistant that we exist
	mqttClient.SendConfig()

	// every minute send everything to the mqtt broker
	c := cron.New()
	c.AddFunc("@every 0h1m0s", func() {

		mqttClient.SendConfig()
		bin.Update()
		bin.UpdatePosition()
		binJson, errJosn := preparePayload(bin)
		if errJosn != nil {
			log.Debug("error")
		}
		mqttClient.Send(binJson)
	})
	c.Start()

	// Setup a file watcher to get instance updates on file changes
	log.Debug("Setting up file watcher")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println(err)
	}
	defer watcher.Close()
	// defer watcherPosition.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				_ = event
				time.Sleep(time.Second * 1)

				if strings.Contains(event.Name, "ChargerPos") {
					bin.UpdatePosition()
				} else {
					bin.Update()
				}

				binJson, errJosn := preparePayload(bin)
				if errJosn != nil {
					log.Debug("error")
				}
				mqttClient.Send(binJson)
			case err := <-watcher.Errors:
				log.Fatalln(err)
			}
		}
	}()

	if err := watcher.Add(bin.FilePath); err != nil {
		log.Fatalln(err)
	}

	if err := watcher.Add(bin.ChargerFilePath); err != nil {
		log.Fatalln(err)
	}

	<-done

}

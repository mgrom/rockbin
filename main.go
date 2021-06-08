package main

import (
	"os"
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
	watcherPosition, errPosition := fsnotify.NewWatcher()
	if err != nil {
		log.Println(err)
	}
	if errPosition != nil {
		log.Println(errPosition)
	}
	defer watcher.Close()
	defer watcherPosition.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				_ = event
				time.Sleep(time.Second * 1)

				bin.Update()
				binJson, errJosn := preparePayload(bin)
				if errJosn != nil {
					log.Debug("error")
				}
				mqttClient.Send(binJson)
			case err := <-watcher.Errors:
				log.Fatalln(err)

			}

			select {
			case event := <-watcherPosition.Events:
				_ = event
				time.Sleep(time.Second * 1)

				bin.Update()
				binJson, errJosn := preparePayload(bin)
				if errJosn != nil {
					log.Debug("error")
				}
				mqttClient.Send(binJson)
			case errPosition := <-watcherPosition.Errors:
				log.Fatalln(errPosition)
			}
		}
	}()

	if err := watcher.Add(bin.FilePath); err != nil {
		log.Fatalln(err)
	}

	if errPosition := watcherPosition.Add(bin.ChargerFilePath); errPosition != nil {
		log.Fatalln(errPosition)
	}

	<-done

}

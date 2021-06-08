package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"

	log "github.com/sirupsen/logrus"
)

func config() (Bin, mqttConfig) {
	var mqttServer string
	var mqttUser string
	var mqttPassword string
	var mqttStateTopic string
	var mqttStateTemplate string
	var mqttAttributesTopic string
	var mqttAttributesTemplate string
	var mqttDeviceId string
	var sensorName string
	var binFullTime float64
	var unitOfMeasurement string
	var FilePath string
	var ChargerFilePath string
	var LoggingLevel string
	flag.StringVar(&mqttServer, "mqtt_server", "mqtt://localhost:1883", "mqtt broker address")
	flag.StringVar(&mqttUser, "mqtt_user", lookUpEnv("MQTT_USER", ""), "mqtt user")
	flag.StringVar(&mqttPassword, "mqtt_password", lookUpEnv("MQTT_PASSWORD", ""), "mqtt password")
	flag.StringVar(&mqttStateTopic, "mqtt_state_topic", "rockbin/%v", "State topic (%v is replaced with the sensor_name value)")
	flag.StringVar(&mqttStateTemplate, "mqtt_state_template", "{{ value_json.value }}", "template for bin")

	flag.StringVar(&mqttAttributesTopic, "mqtt_attributes_topic", "rockbin/%v", "Attributes topic (%v is replaced with the sensor_name value)")
	flag.StringVar(&mqttAttributesTemplate, "mqtt_attributes_template", "{{ value_json.position | to_json }}", "template for charger position")

	flag.StringVar(&sensorName, "sensor_name", "vacuumbin", "Name of sensor in Home Assistant")
	flag.StringVar(&mqttDeviceId, "device_id", "VacumParams", "ID of device in Home Assistant")
	flag.Float64Var(&binFullTime, "full_time", 2400., "Amount of seconds where the bin will be considered full")
	flag.StringVar(&unitOfMeasurement, "measurement_unit", "%", "In what unit should the measurement be sent (%, sec, min)")
	flag.StringVar(&FilePath, "file_path", "/mnt/data/rockrobo/RoboController.cfg", "file path of RoboController.cfg")
	flag.StringVar(&ChargerFilePath, "charger_file_path", "/mnt/data/rockrobo/ChargerPos.data", "file path of ChargerPos.data")
	flag.StringVar(&LoggingLevel, "log_level", "Fatal", "Level of logging (trace, debug, info, warn, error, fatal, panic).")
	flag.Parse()

	setUpLogger(LoggingLevel)
	printVersion()

	bin := Bin{
		FilePath:        FilePath,
		ChargerFilePath: ChargerFilePath,
		Capacity:        binFullTime,
		Unit:            unitOfMeasurement,
	}

	mqttURL, err := url.Parse(mqttServer)
	if err != nil {
		log.Fatalln(err)
	}

	deviceMqtt := Device{
		DeviceID:     sensorName,
		Model:        "Vacuum Parameters",
		Manufacturer: "Bolt",
		Name:         sensorName,
	}

	mqttClient := mqttConfig{
		Name:              sensorName,
		UnitOfMeasurement: unitOfMeasurement,
		StateTopic:        fmt.Sprintf(mqttStateTopic, sensorName),
		StateTemplate:     mqttStateTemplate,
		AttributesTopic:   fmt.Sprintf(mqttAttributesTopic, sensorName),
		// AttributesTemplate: mqttAttributesTemplate,
		ConfigTopic: fmt.Sprintf("homeassistant/sensor/%v/config", sensorName),
		Device:      deviceMqtt,
		UniqueID:    sensorName,
	}
	mqttClient.Connect(mqttURL, mqttUser, mqttPassword)
	return bin, mqttClient
}

func printVersion() {
	if len(os.Args) > 1 {
		if os.Args[1] == "version" {
			fmt.Println(Version)
			os.Exit(0)
		}
	}
}

func setUpLogger(level string) {
	loglevel, _ := log.ParseLevel(level)
	log.SetLevel(loglevel)
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.Info("Starting rockbin service")
	log.WithFields(log.Fields{"loglevel": log.GetLevel()}).Debug("Setup logger with log level")
}

func lookUpEnv(variable, defaultVariable string) string {
	if envVariable, ok := os.LookupEnv(variable); ok {
		return envVariable
	}
	return defaultVariable
}

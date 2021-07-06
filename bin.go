package main

import (
	"bufio"
	"fmt"
	"time"

	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

var defaultCapacity = 60 * 40.

// Bin a type to represent the bin
type Bin struct {
	FilePath        string   `json:"-"`
	ChargerFilePath string   `json:"-"`
	Capacity        float64  `json:"-"`
	Seconds         float64  `json:"-"`
	Unit            string   `json:"-"`
	Value           string   `json:"value"`
	Position        Position `json:"position"`
}

type BinState struct {
	Value    string   `json:"value"`
	Position Position `json:"position"`
}
type Position struct {
	PositionX string `json:"position_x"`
	PositionY string `json:"position_y"`
	Angle     string `json:"angle"`
}

func (b *Bin) UpdatePosition() {
	filePosition, errPosition := os.Open(b.ChargerFilePath)
	if errPosition != nil {
		log.Fatalln(errPosition)
	}
	defer filePosition.Close()
	scannerPosittion := bufio.NewScanner(filePosition)
	linePosition := ""
	xResult := ""
	yResult := ""
	angleResult := ""
	for scannerPosittion.Scan() {
		linePosition = scannerPosittion.Text()
		if strings.Contains(linePosition, "x") {
			xResult = strings.Split(linePosition, "=")[1]
			xResult = strings.Trim(xResult, " ;")
		}
		if strings.Contains(linePosition, "y") {
			yResult = strings.Split(linePosition, "=")[1]
			yResult = strings.Trim(yResult, " ;")
		}
		if strings.Contains(linePosition, "angle") {
			angleResult = strings.Split(linePosition, "=")[1]
			angleResult = strings.Trim(angleResult, " ;")
		}
	}
	filePosition.Close()
	tmpX := .0
	tmpX, errPosition = strconv.ParseFloat(xResult, 32)
	b.Position.PositionX = fmt.Sprintf("%.0f", tmpX)
	tmpY := .0
	tmpY, errPosition = strconv.ParseFloat(yResult, 32)
	b.Position.PositionY = fmt.Sprintf("%.0f", tmpY)
	tmpAngle := .0
	tmpAngle, errPosition = strconv.ParseFloat(angleResult, 32)
	b.Position.Angle = fmt.Sprintf("%.0f", tmpAngle)
	log.WithFields(log.Fields{"xPosition": b.Position.PositionX}).Info("Parsed x position")
	log.WithFields(log.Fields{"yPosition": b.Position.PositionY}).Info("Parsed y position")
	if errPosition != nil {
		log.Fatalln(errPosition)
	}
}

// Update update the bin values
func (b *Bin) Update() {
	file, err := os.Open(b.FilePath)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	line := ""
	for scanner.Scan() {
		line = scanner.Text()
		if strings.Contains(line, "bin_in_time") {
			line = strings.Split(line, "=")[1]
			line = strings.Trim(line, " ;")
			break
		}
	}
	file.Close()
	b.Seconds, err = strconv.ParseFloat(line, 32)
	log.WithFields(log.Fields{"bin_time": b.Seconds}).Info("Parsed bin time")
	if err != nil {
		log.Fatalln(err)
	}
	b.convert()
	if err != nil {
		log.Fatalln(err)
	}
}

// Convert convert the observed value to desired value
func (b *Bin) convert() {

	switch b.Unit {
	case "%":
		b.Value = fmt.Sprintf("%.2f", b.Seconds/b.Capacity*100.)
	case "sec":
		b.Value = fmt.Sprintf("%.0f", (time.Duration(b.Seconds) * time.Second).Seconds())
	case "min":
		b.Value = fmt.Sprintf("%.0f", (time.Duration(b.Seconds) * time.Second).Minutes())
	default:
		b.Value = fmt.Sprintf("%.2f", b.Seconds/b.Capacity*100.)
	}
	log.WithFields(log.Fields{"bin_time": b.Seconds, "value": b.Value}).Info("Converted Value")

}

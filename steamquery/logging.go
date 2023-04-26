package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
)

const (
	defaultLogPath = "./logs"
)

var (
	infSign  = color.WhiteString("i")
	warnSign = color.YellowString("!")
	errSign  = color.RedString("x")
	sucSign  = color.GreenString("✓")

	day, month, year = time.Now().Date()

	logFile *os.File
)

func writeInfo(message interface{}) {
	log.Printf("[%s] %v\n", infSign, message)

	_, err := logFile.WriteString(fmt.Sprintf("%v", message))
	if err != nil {
		log.Printf("[%s] Error writing to log file: %s\n", errSign, err.Error())
	}
}

func writeWarning(message interface{}) {
	log.Printf("[%s] %v\n", warnSign, message)

	_, err := logFile.WriteString(fmt.Sprintf("%v", message))
	if err != nil {
		log.Printf("[%s] Error writing to log file: %s\n", errSign, err.Error())
	}
}

func writeError(message interface{}) {
	log.Printf("[%s] %v\n", errSign, message)

	_, err := logFile.WriteString(fmt.Sprintf("%v", message))
	if err != nil {
		log.Printf("[%s] Error writing to log file: %s\n", errSign, err.Error())
	}
}

func writeSuccess(message interface{}) {
	log.Printf("[%s] %v\n", sucSign, message)

	_, err := logFile.WriteString(fmt.Sprintf("%v", message))
	if err != nil {
		log.Printf("[%s] Error writing to log file: %s\n", errSign, err.Error())
	}
}

func createLogFile(dir string) error {
	if dir == defaultLogPath {
		err := createDefaultLogDirectory()
		if err != nil {
			return err
		}
	}

	logFileName := fmt.Sprintf("%s/steamquery_%d_%d_%d.log", defaultLogPath, year, int(month), day)
	f, err := os.OpenFile(logFileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	logFile = f
	return nil
}

// In case user does not specify different log directory.
func createDefaultLogDirectory() error {
	if _, err := os.Stat(defaultLogPath); os.IsNotExist(err) {
		if err := os.Mkdir(defaultLogPath, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
)

var (
	lastQueryRunFile *os.File
)

// Blocks current scope until CTRL+C is hit.
func listenForCTRLC() {
	writeInfo("Press CTRL+C to cancel any time...")
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done
	fmt.Println("")
}

// Function to create lastQuery file.
func createLastQueryRunFile() error {
	f, err := os.OpenFile(fmt.Sprintf("%s/last_query.json", defaultLogPath),
		os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	lastQueryRunFile = f
	return nil
}

// Function to write to lastQuery file.
func writeToQueryLogFile(message interface{}) error {
	// Truncate file before.
	if err := os.Truncate(lastQueryRunFile.Name(), 0); err != nil {
		return err
	}
	// Write new and old values to file .
	_, err := lastQueryRunFile.WriteString(fmt.Sprintf("%v", message))
	return err
}

// Function to read from lastQuery file.
func readFromQueryLogFile() (*lastQueryRunFormat, error) {
	f, err := os.Open(lastQueryRunFile.Name())
	if err != nil {
		return nil, err
	}

	input, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	// In case the file is empty.
	if len(input) == 0 {
		return &lastQueryRunFormat{}, nil
	}

	var l lastQueryRunFormat

	err = json.Unmarshal(input, &l)

	return &l, err
}

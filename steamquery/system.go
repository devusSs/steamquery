package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

const (
	ipLocationURL = "http://ip-api.com/json/"
)

var (
	lastQueryRunFile *os.File
)

// Blocks current scope until CTRL+C is hit.
func listenForCTRLC() {
	writeInfo("Press CTRL+C to cancel any time...")
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	<-done
	fmt.Println("")
}

// Function to create lastQuery file.
func createLastQueryRunFile() error {
	f, err := os.OpenFile(fmt.Sprintf("%s/last_query.json", defaultLogPath), os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	lastQueryRunFile = f
	return nil
}

// Function to write to lastQuery file.
func writeToQueryLogFile(message interface{}) error {
	// Truncate file before.
	if err := lastQueryRunFile.Truncate(0); err != nil {
		return err
	}
	// Jump to beginning of file before writing.
	_, err := lastQueryRunFile.Seek(0, 0)
	if err != nil {
		return err
	}
	// Write new and old values to file .
	_, err = lastQueryRunFile.WriteString(fmt.Sprintf("%v", message))
	return err
}

// Function to read from lastQuery file.
func readFromQueryLogFile() (*lastQueryRunFormat, error) {
	input, err := io.ReadAll(lastQueryRunFile)
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

// Function to truncate lastQuery file in case it is too old or invalid.
func truncateLastQueryRunFile() error {
	if err := lastQueryRunFile.Truncate(0); err != nil {
		return err
	}
	_, err := lastQueryRunFile.Seek(0, 0)
	return err
}

// Function to return own IP.
func getOwnIPAddress() (string, error) {
	resp, err := http.Get("https://ifconfig.me")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("got unwated ip response: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// Function to get country from IP address.
func getIPLocation(ip string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s%s", ipLocationURL, ip))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("got unwated ip response: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var ipLoc ipLocationInfo

	if err := json.Unmarshal(body, &ipLoc); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s.%s", ipLoc.RegionName, ipLoc.Country), nil
}

// Functions which restarts the app with same flags.
func restartApp() {
	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Start(); err != nil {
		log.Fatalf("[%s] Failed to restart: %s\n", errSign, err.Error())
	}
	os.Exit(0)
}

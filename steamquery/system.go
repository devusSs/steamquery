package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

const (
	communityURL = "steamcommunity.com"
)

// Blocks current scope until CTRL+C is hit.
func listenForCTRLC() {
	writeInfo("Press CTRL+C to cancel any time...")
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done
	fmt.Println("")
}

// Checks if Steamcommunity is reachable.
func pingSteamOnline() error {
	pinger, err := probing.NewPinger(communityURL)
	if err != nil {
		return err
	}

	// Elevate ping so the program does not crash.
	if runtime.GOOS == "windows" {
		pinger.SetPrivileged(true)
	}

	// Elevate ping and instruct user to set elevation via ssh / shell.
	if runtime.GOOS == "linux" {
		pinger.SetPrivileged(true)
		writeWarning("Detected you are running Linux. Please make sure to enable following setting:")
		ex, err := os.Executable()
		if err != nil {
			return err
		}

		writeWarning(fmt.Sprintf("setcap cap_net_raw=+ep %s", ex))

		fmt.Printf("Did you enter that command (y/n)? ")

		inputReader := bufio.NewReader(os.Stdin)

		userInput, err := inputReader.ReadString('\n')
		if err != nil {
			return err
		}

		if strings.TrimSpace(strings.ReplaceAll(userInput, "\n", "")) != "y" {
			return errors.New("cannot use this tool without setting that command")
		}
	}

	pinger.Timeout = 3 * time.Second

	pinger.Count = 3

	err = pinger.Run()
	if err != nil {
		return err
	}

	stats := pinger.Statistics()

	if stats.AvgRtt.Milliseconds() > 500 {
		writeWarning(fmt.Sprintf("RTT to %s exceeds 500 ms (measured: %d ms)", communityURL, stats.AvgRtt.Milliseconds()))
		writeWarning(fmt.Sprintf("Calls to %s may therefor be delayed", communityURL))
	} else {
		writeInfo(fmt.Sprintf("Ping to %s succeeded, measured %d ms", communityURL, stats.AvgRtt.Milliseconds()))
	}

	return nil
}

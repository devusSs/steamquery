package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/go-ping/ping"
)

const (
	communityURL = "steamcommunity.com"
)

// Blocks current scope until CTRL+C is hit.
func listenForCTRLC() {
	log.Println("[INFO] Press CTRL+C to cancel any time...")
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done
	fmt.Println("")
}

// Checks if Steam is reachable.
func pingSteamOnline() error {
	pinger, err := ping.NewPinger(communityURL)
	if err != nil {
		return err
	}

	pinger.Count = 3

	err = pinger.Run()
	if err != nil {
		return err
	}

	stats := pinger.Statistics()

	if stats.AvgRtt.Milliseconds() > 500 {
		log.Printf("[WARN] Average RTT to %s: %d ms\n", communityURL, stats.AvgRtt.Milliseconds())
		log.Printf("[WARN] Calls to %s may be delayed\n", communityURL)
	}

	return nil
}

func determineOS() string {
	return runtime.GOOS
}

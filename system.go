package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	probing "github.com/prometheus-community/pro-bing"
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

// Checks if Steamcommunity is reachable.
func pingSteamOnline() error {
	pinger, err := probing.NewPinger(communityURL)
	if err != nil {
		return err
	}

	pinger.Timeout = 3 * time.Second

	pinger.Count = 3

	err = pinger.Run()
	if err != nil {
		return err
	}

	stats := pinger.Statistics()

	if stats.AvgRtt.Milliseconds() > 500 {
		log.Printf("[WARN] RTT to %s exceeds 500 ms (measured: %d ms)\n", communityURL, stats.AvgRtt.Milliseconds())
		log.Printf("[WARN] Calls to %s may therefor be delayed\n", communityURL)
	} else {
		log.Printf("[INFO] Ping to %s succeeded, measured %d ms\n", communityURL, stats.AvgRtt.Milliseconds())
	}

	return nil
}

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// Blocks current scope until CTRL+C is hit.
func listenForCTRLC() {
	writeInfo("Press CTRL+C to cancel any time...")
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done
	fmt.Println("")
}

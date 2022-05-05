package signals

import (
	"os"
	"os/signal"
	"syscall"

	"kusionstack.io/kusion/pkg/log"
)

var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}

// listen for interrupts or the SIGTERM signal and
// executing clean job
func HandleInterrupt() {
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, shutdownSignals...)
	go func() {
		<-stopCh
		log.Info("Received termination, signaling shutdown, executing clean job")
	}()
}

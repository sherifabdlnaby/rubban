package rubban

import (
	"os"
	"os/signal"
	"syscall"
)


// Main is the main function of the application, it will be run by cobra's root command.
func Main() {

	// Create App
	rubban := New()

	// Handle Termination Signals
	shutdownSignal := make(chan struct{}, 1)
	go onTerminationSignal(func() {
		rubban.Stop()
		shutdownSignal <- struct{}{}
	})

	// Initialize App
	err := rubban.Initialize()
	if err != nil {
		panic("Failed to Initialize Rubban. Error: " + err.Error())
		return
	}

	// Start
	go rubban.Start()

	// Wait to Shutdown
	<-shutdownSignal

	os.Exit(0)
}

func onTerminationSignal(callback func()) {
	// Signal Channels
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-signalChan
	callback()
}

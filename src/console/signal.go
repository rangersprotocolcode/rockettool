package console

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func getExitSignalsChan() <-chan bool {
	quit := make(chan bool)

	go func() {
		signals := make(chan os.Signal)
		defer close(signals)

		signal.Notify(signals, syscall.SIGQUIT, syscall.SIGTERM, os.Interrupt)
		defer signal.Stop(signals)

		<-signals
		quit <- true
	}()

	return quit
}

func handleExit(ctrlC <-chan bool) {
	<-ctrlC
	fmt.Println("exiting...")
	os.Exit(0)
}

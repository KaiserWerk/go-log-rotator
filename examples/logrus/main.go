package main

import (
	"log"
	"sync"

	rotator "github.com/KaiserWerk/go-log-rotator"

	"github.com/sirupsen/logrus"
)

func main() {
	// this creates a new Rotator with a maximum file size of 2KB and 15 rotated files are to be kept on disk
	// logrus DOES take care of thread-safe writes, so supply 'false' as last parameter to avoid unnecessary overhead
	rotator, err := rotator.New(".", "logrus-logger.log", 2<<10, 0644, 15, false)
	if err != nil {
		log.Fatal("could not create rotator:", err.Error())
	}

	logger := logrus.New()
	logger.SetOutput(rotator) // use the rotator here
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.TextFormatter{})

	var wg sync.WaitGroup
	wg.Add(3)
	go func(w *sync.WaitGroup) {
		for i := 0; i < 100; i++ {
			logger.Info("Hello World!")
		}
		w.Done()
	}(&wg)
	go func(w *sync.WaitGroup) {
		for i := 0; i < 100; i++ {
			logger.Warn("Goodbye...")
		}
		w.Done()
	}(&wg)
	go func(w *sync.WaitGroup) {
		for i := 0; i < 100; i++ {
			logger.Error("How's it going?")
		}
		w.Done()
	}(&wg)
	wg.Wait()

	// done? then close up
	_ = rotator.Close()

	// you should see 16 files by now
}

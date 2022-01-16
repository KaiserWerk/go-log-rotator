package main

import (
	"log"
	"sync"

	rotator "github.com/KaiserWerk/go-log-rotator"
)

func main() {
	// this creates a new Rotator with a maximum file size of 10KB and 3 rotated files are to be kept on disk
	// the default logger does NOT take care of thread-safe writes, so supply 'true' as last parameter
	rotator, _ := rotator.New(".", "standard-test.log", 10<<10, 0644, 3, true)

	logger := log.New(rotator, "", 0)

	var wg sync.WaitGroup
	wg.Add(3)
	go func(w *sync.WaitGroup) {
		for i := 0; i < 100; i++ {
			logger.Println("Hello World!")
		}
		w.Done()
	}(&wg)
	go func(w *sync.WaitGroup) {
		for i := 0; i < 100; i++ {
			logger.Println("Goodbye...")
		}
		w.Done()
	}(&wg)
	go func(w *sync.WaitGroup) {
		for i := 0; i < 100; i++ {
			logger.Println("How's it going?")
		}
		w.Done()
	}(&wg)

	wg.Wait()

	// done? close it
	_ = rotator.Close()

	// the log file 'standard-test.log' should now have 300 entries
}

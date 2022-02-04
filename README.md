# Go Log Rotator

[![Go Reference](https://pkg.go.dev/badge/github.com/KaiserWerk/go-log-rotator.svg)](https://pkg.go.dev/github.com/KaiserWerk/go-log-rotator)

The Log Rotator provides a simple io.WriteCloser to be used as output for your logger of choice.

### Usage examples (refer to examples folder)

Standard package logger:

```golang
package main

import (
	"fmt"
	"log"
	"sync"

	logRotator "github.com/KaiserWerk/go-log-rotator"
)

func main() {
	// this creates a new Rotator with a maximum file size of 10KB and 3 rotated files are to be kept on disk
	// the default logger does NOT take care of thread-safe writes, so supply 'true' as last parameter
	rotator, err := logRotator.New(".", "standard-logger.log", 10<<10, 0644, 3, true)
	if err != nil {
		log.Fatalf("could not create rotator: %s", err.Error())
	}
	defer rotator.Close()
	logger := log.New(rotator, "", 0)

	var wg sync.WaitGroup
	wg.Add(3)
	go func(w *sync.WaitGroup) {
		for i := 0; i < 3000; i++ {
			logger.Println("Hello World!")
		}
		w.Done()
	}(&wg)
	go func(w *sync.WaitGroup) {
		for i := 0; i < 3000; i++ {
			logger.Println("Goodbye...")
		}
		w.Done()
	}(&wg)
	go func(w *sync.WaitGroup) {
		for i := 0; i < 3000; i++ {
			logger.Println("How's it going?")
		}
		w.Done()
	}(&wg)

	wg.Wait()

	// there should be 4 rotator files by now
}
```

Logrus:

```golang
package main

import (
	"log"
	"sync"

	logRotator "github.com/KaiserWerk/go-log-rotator"

	"github.com/sirupsen/logrus"
)

func main() {
	// this creates a new Rotator with a maximum file size of 2KB and 15 rotated files are to be kept on disk
	// logrus DOES take care of thread-safe writes, so supply 'false' as last parameter to avoid unnecessary overhead
	rotator, err := logRotator.New(".", "logrus-logger.log", 2<<10, 0644, 15, false)
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
```

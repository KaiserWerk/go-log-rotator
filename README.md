# Go Log Rotator

The Log Rotator provides a simple io.WriteCloser to be used as output for your logger of choice.

### Usage examples (refer to examples folder)

Standard package logger:

```golang
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
```

Logrus:

```golang
package main

import (
	"sync"

	rotator "github.com/KaiserWerk/go-log-rotator"

	"github.com/sirupsen/logrus"
)

func main() {
	// this creates a new Rotator with a maximum file size of 2MB and 15 rotated files are to be kept on disk
	// logrus DOES take care of thread-safe writes, so supply 'false' as last parameter to avoid unnecessary overhead
	rotator, _ := rotator.New(".", "logrus-test.log", 2<<20, 0644, 15, false)

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

	// the log file 'logrus-test.log' should contain 300 entries by now
}

```
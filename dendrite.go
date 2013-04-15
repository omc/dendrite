package main

import (
	"./src/dendrite"
	"bufio"
	"flag"
	"github.com/fizx/logs"
	"io"
	"log"
	"os"
	"runtime"
	"time"
)

var configFile = flag.String("f", "/etc/dendrite/config.yaml", "location of the config file")
var debug = flag.Bool("d", false, "log at DEBUG")
var logFile = flag.String("l", "/var/log/dendrite.log", "location of the log file")
var cpus = flag.Int("c", runtime.NumCPU(), "number of cpus to possibly use")
var quitAfter = flag.Float64("q", -1, "quit after this many seconds (useful for tests)")

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(*cpus)

	// set the logger path
	handle, err := os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		logs.Warn("Unable to open log file %s, using stderr: %s", *logFile, err)
	} else {
		logs.Logger = log.New(handle, "", log.LstdFlags|log.Lshortfile)
	}

	// Check whether we're in debug mode
	if *debug {
		logs.SetLevel(logs.DEBUG)
		logs.Debug("logging at DEBUG")
	} else {
		logs.SetLevel(logs.INFO)
	}

	// Read the config files
	config, err := dendrite.NewConfig(*configFile)
	if err != nil {
		logs.Fatal("Can't read configuration: %s", err)
	}

	// Link up all of the objects
	ch := make(chan dendrite.Record, 100)
	logs.Debug("original %s", ch)
	dests := config.CreateDestinations()
	groups := config.CreateAllTailGroups(ch)

	// If any of our destinations talk back, log it.
	go func() {
		reader := bufio.NewReader(dests.Reader())
		for {
			str, err := reader.ReadString('\n')
			if err == io.EOF {
				logs.Debug("eof")
				time.Sleep(1 * time.Second)
			} else if err != nil {
				logs.Error("error reading: %s", err)
			} else {
				logs.Info("received: %s", str)
			}
		}
	}()

	// Do the event loop
	finished := make(chan bool, 0)
	go dests.Consume(ch, finished)
	if *quitAfter >= 0 {
		start := time.Now()
		logs.Debug("starting the poll")
		for {
			groups.Poll()
			if time.Now().Sub(start) >= time.Duration((*quitAfter)*float64(time.Second)) {
				break
			}
		}
	} else {
		logs.Debug("starting the loop")
		groups.Loop()
	}
	logs.Info("Closing...")
	close(ch)
	<-finished
	logs.Info("Goodbye!")
}

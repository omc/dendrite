package main

import (
	"./src/dendrite"
	"strings"
	"bufio"
	"flag"
	"fmt"
	"github.com/fizx/logs"
	"github.com/kylelemons/go-gypsy/yaml"
	"io/ioutil"
	"net"
	"os"
	"path"
	"time"
)

var configFile = flag.String("f", "/etc/dendrite/config.yaml", "location of the config file")
var debug = false

func main() {
	flag.BoolVar(&debug, "d", false, "log at DEBUG")
	flag.Parse()
	if debug {
		logs.SetLevel(logs.DEBUG)
		logs.Debug("logging at DEBUG")
	} else {
		logs.SetLevel(logs.INFO)
	}
	doc, err := yaml.ReadFile(*configFile)
	if err != nil {
		logs.Error("Can't read the config file, error was: ", err)
	}

	config := new(dendrite.Config)
	config.Populate(doc)

	confd := path.Dir(*configFile) + "/conf.d"
	entries, err := ioutil.ReadDir(confd)
	if err != nil {
		logs.Warn("Can't read relevant conf.d: %s", err)
	} else {
		for _, entry := range entries {
			path := confd + "/" + entry.Name()
			doc, err := yaml.ReadFile(path)
			if err != nil {
				logs.Warn("Can't read relevant conf.d: %s", err)
			} else {
			  base := strings.Replace(entry.Name(), ".yaml", "", 1)
				config.AddGroup(base, doc.Root)
			}
		}
	}

	out := make(chan string, 0)
	groups := make([]*dendrite.TailGroup, 0)
	for _, cg := range config.Groups {
		groups = append(groups, dendrite.NewTailGroup(cg, out))
	}
	go func() {
		for {
			for _, g := range groups {
				g.Poll()
				time.Sleep(time.Second)
			}
		}
	}()
	switch config.Protocol {
	case "file":
		file, err := os.OpenFile(config.Address, 0, 0777)
		if err != nil {
			logs.Error("Can't write to: ", err)
		}
		for {
			line := <-out
			fmt.Fprintln(file, line)
		}
	case "udp", "tcp":
		conn, err := net.Dial(config.Protocol, config.Address)
		if err != nil {
			logs.Error("Can't write to: ", err)
		} else {
			if config.Protocol == "tcp" {
				go func() {
					reader := bufio.NewReader(conn)
					for {
						str, err := reader.ReadString('\n')
						if err != nil {
							logs.Warn("connection closed with %s", err)
							os.Exit(0)
						} else {
							logs.Info("received: %s", str)
						}
					}
				}()
			}

			for {
				line := <-out
				fmt.Fprintln(conn, line)
			}
		}
	default:
		logs.Error("Unknown protocol: ", config.Protocol)
	}
}

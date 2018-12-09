package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/orktes/homeautomation-homekit/homekit"
)

var (
	configFile = flag.String("config", "", "path to the config file")
)

func main() {
	flag.Parse()

	if *configFile == "" {
		flag.PrintDefaults()
		return
	}

	//log.Debug.Enable()

	file, err := os.Open(*configFile)
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(file)
	conf, err := homekit.ParseConfig(reader)
	if err != nil {
		panic(err)
	}

	b, _ := json.MarshalIndent(conf, "", "  ")
	fmt.Printf("%s\n", string(b))

	hk, err := homekit.New(conf)
	if err != nil {
		panic(err)
	}
	defer hk.Close(context.Background())

	go func() {
		err := hk.Start()
		if err != nil {
			panic(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)

	<-c
	fmt.Println("Shutting down")
}

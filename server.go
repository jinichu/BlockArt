package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"

	"./server"
)

var config server.Config

func readConfigOrDie(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	buffer, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(buffer, &config); err != nil {
		return err
	}
	return err
}

func main() {
	path := flag.String("c", "", "Path to the JSON config")

	log.SetFlags(log.Flags() | log.Lshortfile)
	flag.Parse()

	if *path == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err := readConfigOrDie(*path); err != nil {
		log.Fatal(err)
	}

	rand.Seed(time.Now().UnixNano())

	s, err := server.New(config)
	if err != nil {
		log.Fatal(err)
	}
	if err := s.Listen(); err != nil {
		log.Fatal(err)
	}
}

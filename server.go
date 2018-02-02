package main

import (
	"flag"
	"log"

	"./server"
)

var bind = flag.String("bind", ":5001", "address to bind to")

func main() {
	log.SetFlags(log.Flags() | log.Lshortfile)
	flag.Parse()

	s, err := server.New()
	if err != nil {
		log.Fatal(err)
	}
	if err := s.Listen(*bind); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"github.com/stoex/go-lock/internal/config"
	"log"
)

var (
	// Version denotes the program version
	Version string
	// BuildDate denotes the build date
	BuildDate     string
	configuration = config.NewManager()
)

func main() {
	log.Printf("go-lock :: version %s :: build date %s", Version, BuildDate)
	log.Println(configuration.Redlock.Clients)
}

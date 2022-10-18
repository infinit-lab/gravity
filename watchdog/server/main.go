package main

import (
	"github.com/infinit-lab/gravity/config"
	"github.com/infinit-lab/gravity/printer"
	"github.com/infinit-lab/gravity/server"
	"os"
)

func main() {
	os.Args = append(os.Args, "server.port=10000")
	config.LoadArgs()

	printer.Trace("Watchdog run...")
	err := server.Run()
	if err != nil {
		printer.Error(err)
	}
}

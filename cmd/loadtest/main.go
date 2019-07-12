package main

import (
	"github.com/jessevdk/go-flags"
	"ws/cmd/loadtest/simulator"
)

func main() {
	syncCmd := simulator.SyncCmd{}

	parser := flags.NewParser(nil, flags.Default)
	parser.AddCommand("sync", "entity sync test scenario", "operations to have all the users join the same space and test entity sync with heavy traffic", &syncCmd)

	parser.Parse()
}

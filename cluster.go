package main

import (
	"flag"
	"github.com/rivo/tview"
)

func main() {

	rpcPath := flag.String("rpc", "./lightning-rpc", "Path to lightning-rpc socket")
	flag.Parse()

	ui := &UI{
		tview.NewApplication(),
		tview.NewPages(),
		make(map[string]tview.Primitive),
		nil,
		NewLog(),
		*rpcPath,
	}

	ui.Run()
}

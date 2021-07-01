package main

import (
	"flag"
	"fmt"
	"github.com/rivo/tview"
)

func main() {

	rpcPath := flag.String("rpc", "./lightning-rpc", "Path to lightning-rpc socket")
	flag.Parse()

	fmt.Println("rpc:", *rpcPath)

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

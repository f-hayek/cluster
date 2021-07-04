package main

import (
	"github.com/fiatjaf/lightningd-gjson-rpc"
	"github.com/tidwall/gjson"
	"log"
	"os"
	"time"
)

type LnClient struct {
	*lightning.Client
	ui *UI
}

var ln *LnClient

func NewClient(ui *UI) *LnClient {
	if ln != nil {
		return ln
	} else {
		// check if the socket exists
		if _, err := os.Stat(ui.rpcPath); os.IsNotExist(err) {
			ui.log.Warn("RPC socket " + ui.rpcPath + " does't seem to exist.\n")
			ui.log.Info("You can pass the path to the socket using --rpc=/path/to/lightning-rpc\n")
		}
		client := &lightning.Client{
			Path:        ui.rpcPath,
			CallTimeout: 30 * time.Second, // optional, defaults to 5 seconds
		}
		ln = &LnClient{
			client,
			ui,
		}
		return ln
	}
}
func getInfo(ui *UI) gjson.Result {
	client := NewClient(ui)

	ui.log.Info("getinfo ")
	getinfo, err := client.Call("getinfo")
	if err != nil {
		ui.log.Warn("error: " + err.Error())
	}
	ui.log.Ok("OK\n")
	return getinfo
}
func getConfig(ui *UI) gjson.Result {
	client := NewClient(ui)

	ui.log.Info("listconfigs ")
	listconfigs, err := client.Call("listconfigs")
	if err != nil {
		ui.log.Warn("error: " + err.Error())
	}
	ui.log.Ok("OK\n")
	return listconfigs
}
func getFeerates(ui *UI) gjson.Result {
	client := NewClient(ui)

	ui.log.Info("feerates ")
	feerates, err := client.Call("feerates", "perkb")
	if err != nil {
		ui.log.Warn("error: " + err.Error())
	}
	ui.log.Ok("OK\n")
	return feerates
}

func getNewAddr(ui *UI) gjson.Result {
	client := NewClient(ui)
	newAddr, err := client.Call("newaddr")
	if err != nil {
		log.Fatal("newaddr error: " + err.Error())
	}
	return newAddr
}
func getInvoice(ui *UI, params map[string]interface{}) gjson.Result {
	client := NewClient(ui)

	ui.log.Info("invoice ")
	invoice, err := client.Call("invoice", params)
	if err != nil {
		ui.log.Warn("invoice error: " + err.Error())
	}
	ui.log.Ok("OK\n")
	ui.log.Info("bolt11: [white]" + invoice.Get("bolt11").String() + "\n")
	ui.log.Info("payment_hash: [white]" + invoice.Get("payment_hash").String() + "\n")
	return invoice

}
func getNode(ui *UI, id string) Node {
	client := NewClient(ui)

	ui.log.Info("listnodes [white]" + id + " ")
	node, err := client.Call("listnodes", id)
	if err != nil {
		ui.log.Warn("error: " + err.Error() + "\n")
	}
	ui.log.Ok("OK\n")

	return Node{
		id:    node.Get("nodes.0.id").String(),
		alias: node.Get("nodes.0.alias").String(),
		color: node.Get("nodes.0.color").String(),
	}

}
func getChannel(ui *UI, chanID string) gjson.Result {
	client := NewClient(ui)

	channel, err := client.Call("listchannels", chanID)

	if err != nil {
		log.Fatal("listchannels error: " + err.Error())
	}

	return channel
}

func getForwards(ui *UI, params map[string]interface{}) gjson.Result {
	client := NewClient(ui)

	ui.log.Info("listforwards ")

	forwards, err := client.Call("listforwards", params)

	if err != nil {
		ui.log.Warn("error: " + err.Error())
	}
	ui.log.Ok("OK\n")

	return forwards
}

func getFunds(ui *UI, spent bool) gjson.Result {
	client := NewClient(ui)

	ui.log.Info("listfunds ")

	funds, err := client.Call("listfunds", spent)

	if err != nil {
		ui.log.Warn("error: " + err.Error())
	}
	ui.log.Ok("OK\n")

	return funds
}

func getTransactions(ui *UI) gjson.Result {
	client := NewClient(ui)

	ui.log.Info("listtransactions ")

	transactions, err := client.Call("listtransactions")

	if err != nil {
		ui.log.Warn("error: " + err.Error())
	}
	ui.log.Ok("OK\n")

	return transactions
}
func getPays(ui *UI) gjson.Result {
	client := NewClient(ui)

	ui.log.Info("listpays ")
	pays, err := client.Call("listpays")
	if err != nil {
		ui.log.Warn("error: " + err.Error())
	}
	ui.log.Ok("OK\n")
	return pays
}

func decodePay(ui *UI, bolt11 string) gjson.Result {
	client := NewClient(ui)

	ui.log.Info("decodepay ")
	decoded, err := client.Call("decodepay", bolt11)
	if err != nil {
		ui.log.Warn("error: " + err.Error())
	}
	ui.log.Ok("OK\n")
	return decoded
}
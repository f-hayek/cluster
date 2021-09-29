package main

import (
	"fmt"
	"github.com/fiatjaf/lightningd-gjson-rpc"
	"github.com/tidwall/gjson"
	"os"
	"strings"
	"time"
)

type LnClient struct {
	*lightning.Client
	ui *UI
}

var ln *LnClient
var NodeCache = make(map[string]Node)
var lastCacheLookup time.Time
const cacheFor = time.Second * 60

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
			CallTimeout: 60 * time.Second, // optional, defaults to 5 seconds
		}
		ln = &LnClient{
			client,
			ui,
		}
		return ln
	}
}

func call(ui *UI, method string, params ...interface{}) gjson.Result {
	client := NewClient(ui)

	ui.log.Info(method + " ")

	start := time.Now()

	results, err := client.Call(method, params...)

	if err != nil {
		ui.log.Warn("error: " + err.Error())
	}

	finish := time.Now()

	ui.log.Ok(fmt.Sprintf("[%dms]\n", (finish.Sub(start)).Milliseconds()))

	return results
}

func getInfo(ui *UI) gjson.Result {

	return call(ui, "getinfo")

}
func getConfig(ui *UI) gjson.Result {

	return call(ui, "listconfigs")

}
func getFeerates(ui *UI) gjson.Result {

	return call(ui, "feerates", "perkb")

}

func getNewAddr(ui *UI) gjson.Result {

	return call(ui, "newaddr")

}
func getInvoices(ui *UI) gjson.Result {

	return call(ui, "listinvoices")

}
func getInvoice(ui *UI, params map[string]interface{}) gjson.Result {

	invoice := call(ui, "invoice", params)

	ui.log.Ok("OK\n")
	ui.log.Info("bolt11: [white]" + invoice.Get("bolt11").String() + "\n")
	ui.log.Info("payment_hash: [white]" + invoice.Get("payment_hash").String() + "\n")
	return invoice

}

func wrapNode(results gjson.Result) Node {
	leaseFeeBaseMsat, err := Mstoi(results.Get("option_will_fund.lease_fee_base_msat").String())
	skipOptionWillFund := false
	if err != nil {
		skipOptionWillFund = true
	}
	channelFeeMaxBaseMsat, err := Mstoi(results.Get("option_will_fund.channel_fee_max_base_msat").String())
	if err != nil {
		skipOptionWillFund = true
	}

	node := Node{
		id:    results.Get("nodeid").String(),
		alias: results.Get("alias").String(),
		color: results.Get("color").String(),
	}
	if !skipOptionWillFund {
		leaseFeeBasis := results.Get("option_will_fund.lease_fee_basis").Int()
		fundingWeight := results.Get("option_will_fund.funding_weight").Int()
		channelFeeMaxProportionalThousandths := results.Get("option_will_fund.channel_fee_max_proportional_thousandths").Int()
		compactLease := results.Get("option_will_fund.compact_lease").String()

		node.optionWillFund = &OptionWillFund{
			leaseFeeBaseMsat,
			leaseFeeBasis,
			fundingWeight,
			channelFeeMaxBaseMsat,
			channelFeeMaxProportionalThousandths,
			compactLease,
		}
	}
	return node
}

func listNode(ui *UI, id string) Node {

	n, exists := NodeCache[id]
	if exists {
		return n
	}

	results := call(ui, "listnodes", id)

	node := wrapNode(results.Get("nodes.0"))

	NodeCache[id] = node
	return NodeCache[id]

}

func listNodes(ui *UI) (results []Node) {

	// cache
	if lastCacheLookup.After(time.Now().Add(- cacheFor)) {
		for _, node := range NodeCache {
			results = append(results, node)
		}
		return results
	}
	nodes := call(ui, "listnodes")

	for _, data := range nodes.Get("nodes").Array() {
		node := wrapNode(data)
		results = append(results, node)
		NodeCache[node.id] = node
	}

	lastCacheLookup = time.Now()
	return results
}
func listNodesThatWillFund(ui *UI) []Node {
	var results []Node
	nodes := listNodes(ui)
	for _, node := range nodes {
		if node.optionWillFund != nil {
			results = append(results, node)
		}
	}
	return results
}
func listNodesByAliasOrID(ui *UI, term string) []Node {
	var results []Node
	nodes := listNodes(ui)
	for _, node := range nodes {
		alias := strings.ToLower(node.alias)
		t := strings.Trim(term, " ")
		if strings.Contains(alias, strings.ToLower(t)) || strings.Contains(node.id, t) {
			results = append(results, node)
		}
	}
	return results
}


func listChannels(ui *UI) gjson.Result {

	return call(ui, "listchannels")

}
func getChannel(ui *UI, chanID string) gjson.Result {

	return call(ui, "listchannels", chanID)

}

func getForwards(ui *UI, params map[string]interface{}) gjson.Result {

	return call(ui, "listforwards", params)

}

func listFunds(ui *UI, spent bool) gjson.Result {

	return call(ui, "listfunds", spent)

}

func getTransactions(ui *UI) gjson.Result {

	return call(ui, "listtransactions")

}
func getPays(ui *UI) gjson.Result {

	return call(ui, "listpays")
}

func decodePay(ui *UI, bolt11 string) gjson.Result {

	return call(ui, "decodepay", bolt11)

}

func setChannelFee(ui *UI, scid string, base, rate int) gjson.Result {
	params := map[string]interface{} {
		"id": scid,
		"base": base,
		"ppm": rate,
	}
	return call(ui, "setchannelfee", params)
}

func fundChannel(ui *UI, nodeID string, channelSize int, feerate string, announce bool) gjson.Result {
	params := map[string]interface{} {
		"id": nodeID,
		"amount": fmt.Sprintf("%dsat", channelSize),
		"feerate": feerate,
		"announce": announce,
	}
	return call(ui, "fundchannel", params)
}

func offer(ui *UI, amount int, description string) gjson.Result {
	params := map[string]interface{} {
		"amount": fmt.Sprintf("%dsat", amount),
		"description": description,
	}

	return call(ui, "offer", params)
}

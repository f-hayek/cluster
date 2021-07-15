package main

import (
	"errors"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tidwall/gjson"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
)

type InfoLine struct {
	label string
	value string
}
type InfoColumn struct {
	labelColor string
	valueColor string
	rows []*InfoLine
	topPadding int
}

func NewInfoLine(label, value string) *InfoLine {
	return &InfoLine{label, value}
}
func NewInfoColumn(labelColor, valueColor string) *InfoColumn {
	c := &InfoColumn{
		labelColor: labelColor,
		valueColor: valueColor,
		rows: nil,
		topPadding: 1,
	}
	return c
}
func (ic *InfoColumn) AddRow(label, value string) *InfoColumn {
	ic.rows = append(ic.rows, NewInfoLine(label, value))
	return ic
}
func (ic *InfoColumn) Print(w io.Writer) {
	for i := 0; i < ic.topPadding; i++  {
		fmt.Fprint(w, "\n")
	}
	for _, row := range ic.rows {
		fmt.Fprintf(w, "%s%25v: %s%s\n", ic.labelColor, row.label, ic.valueColor, row.value)
	}
}

func findOutput(outputs []gjson.Result, txid string, outputIdx int64) (int64, error) {
	for _, output := range outputs {
		oTxid := output.Get("txid").String()
		oIdx := output.Get("output").Int()
		if oTxid == txid && oIdx == outputIdx {
			return output.Get("value").Int(), nil
		}
	}
	return 0, errors.New("output not found")
}
func calculateSpentFees(transactions, funds gjson.Result) int64 {
	fees := int64(0)
	for _, tx := range transactions.Get("transactions").Array() {
		vin := int64(0)
		for _, input := range tx.Get("inputs").Array() {
			value, err := findOutput(funds.Get("outputs").Array(), input.Get("txid").String(), input.Get("index").Int())

			if err == nil {
				vin += value
			}

		}
		if vin > 0 {
			vout := int64(0)
			for _, output := range tx.Get("outputs").Array() {
				sats, err := strconv.Atoi(strings.Replace(output.Get("satoshis").String(), "msat", "", 1))

				if err == nil {
					vout += int64(sats) / 1000
				}
			}
			fees += vin - vout
		}
	}
	return fees
}
func dashPage(ui *UI) tview.Primitive {

	// Node Info
	infoPane := tview.NewTextView()
	infoPane.SetBorder(true).SetBorderColor(MainColor).SetTitle(" Node Info ")
	infoPane.SetDynamicColors(true)

	info := getInfo(ui)
	config := getConfig(ui)

	ic := NewInfoColumn("[deepskyblue]", "[white]")
	ic.AddRow("Node alias", info.Get("alias").String())
	ic.AddRow("Node pubkey", info.Get("id").String())
	ic.AddRow("Network", info.Get("network").String())
	ic.AddRow("Blockheight", info.Get("blockheight").String())
	ic.AddRow("Bound to", info.Get("binding.0.address").String() + ":" + info.Get("binding.0.port").String())
	for _, announce := range info.Get("address").Array() {
		ic.AddRow("Announce " + announce.Get("type").String(), announce.Get("address").String() + ":" + announce.Get("port").String())
	}
	ic.AddRow("Peers", info.Get("num_peers").String())
	activeChannels := info.Get("num_active_channels").String()
	ic.AddRow("Active channels", activeChannels)
	ic.AddRow("Offline channels", info.Get("num_inactive_channels").String())
	ic.AddRow("Pending channels", info.Get("num_pending_channels").String())
	var largeChannels string
	if config.Get("large-channels").Bool() {
		largeChannels = "Supported"
	} else {
		largeChannels = "[red]Not supported"
	}
	ic.AddRow("Large channels", largeChannels)
	ic.AddRow("Mininum capacity", formatSats(config.Get("min-capacity-sat").Int()))

	collectedFees := info.Get("msatoshi_fees_collected").Int() / 1000

	ic.AddRow("Fees collected (sats)", "[yellow]" + formatSats(collectedFees))


	// Available Funds
	fundsPane := tview.NewTextView()
	fundsPane.SetBorder(true).SetBorderColor(MainColor).SetTitle(" Available funds ")
	fundsPane.SetDynamicColors(true)

	funds := getFunds(ui, true) // list both confirmed and spent funds

	transactions := getTransactions(ui)

	spentFees := calculateSpentFees(transactions, funds)

	ic.AddRow("Fees spent on-chain", "[yellow]" + formatSats(spentFees))
	profitLoss := collectedFees - spentFees
	var plPrefix string
	if profitLoss > 0 {
		plPrefix = "[green]"
	} else {
		plPrefix = "[red]"
	}
	ic.AddRow("Profit/Loss", plPrefix + formatSats(profitLoss))
	ic.Print(infoPane)

	fc := NewInfoColumn("[deepskyblue]", "[yellow]")
	onChainFunds := int64(0)
	numUtxo := int64(0)

	for _, output := range funds.Get("outputs").Array() {
		if output.Get("status").String() == "confirmed" {
			onChainFunds += output.Get("value").Int()
			numUtxo += 1
		}
	}

	outboundFunds := int64(0)
	totalChannelFunds := int64(0)
	minChan := float64(math.MaxInt64)
	var maxChan float64

	for _, output := range funds.Get("channels").Array() {
		outboundFunds += output.Get("channel_sat").Int()
		chanSize := output.Get("channel_total_sat").Int()
		totalChannelFunds += chanSize
		minChan = math.Min(float64(chanSize), minChan)
		maxChan = math.Max(float64(chanSize), maxChan)
	}

	fc.AddRow("On-chain capacity", formatSats(onChainFunds) + " [white]in [yellow]" + fmt.Sprintf("%d [white]UTXOs", numUtxo))
	fc.AddRow("Outbound LN capacity", formatSats(outboundFunds) + " [white]in [yellow]" + fmt.Sprintf("%s [white]channels", activeChannels))
	fc.AddRow("Total capacity", formatSats(onChainFunds + outboundFunds))
	fc.AddRow("Inbound LN capacity", formatSats(totalChannelFunds - outboundFunds))
	fc.AddRow("Smallest channel", formatSats(int64(minChan)))
	fc.AddRow("Biggest channel", formatSats(int64(maxChan)))
	fc.Print(fundsPane)


	// Current fees

	feesPane := tview.NewTextView()
	feesPane.SetBorder(true).SetBorderColor(MainColor).SetTitle(" Current fee rates ")
	feesPane.SetDynamicColors(true)

	rates := getFeerates(ui)

	fec := NewInfoColumn("[deepskyblue]", "[orange]")
	fec.AddRow("Opening", formatSats(rates.Get("perkb.opening").Int() / 1024) + " sat/vB")
	fec.AddRow("Mutual close", "[green]" + formatSats(rates.Get("perkb.mutual_close").Int() / 1024) + " sat/vB")
	fec.AddRow("Unilateral close", formatSats(rates.Get("perkb.unilateral_close").Int() / 1024) + " sat/vB")
	fec.AddRow("Delayed to us", formatSats(rates.Get("perkb.delayed_to_us").Int() / 1024) + " sat/vB")
	fec.AddRow("HTLC resolution", formatSats(rates.Get("perkb.htlc_resolution").Int() / 1024) + " sat/vB")
	fec.AddRow("Penalty", formatSats(rates.Get("perkb.penalty").Int() / 1024) + " sat/vB")
	fec.AddRow("Min acceptable", formatSats(rates.Get("perkb.min_acceptable").Int() / 1024) + " sat/vB")
	fec.AddRow("Max acceptable", formatSats(rates.Get("perkb.max_acceptable").Int() / 1024) + " sat/vB")

	fec.Print(feesPane)


	// Recent activity

	activity := NewTable()
	activity.SetBorder(true).SetBorderColor(MainColor)
	activity.SetTitle(" Recent activity ")

	activity.Select(4, 0).SetFixed(4, 12)
	activity.AddColumnHeader("\n[bold]date", tview.AlignRight)
	activity.AddColumnHeader("\noperation", tview.AlignCenter)
	activity.AddColumnHeader("\ndestination", tview.AlignRight)
	activity.AddColumnHeader("\namount", tview.AlignRight)
	activity.AddColumnHeader("\n description", tview.AlignRight)
	activity.Separator()

	activity.SetDoneFunc(func(key tcell.Key) {
		ui.FocusMenu()
	})
	rowOffset := 4
	idx := 0
	pays := getPays(ui)


	for _, pay := range pays.Get("pays").Array() {
		if pay.Get("status").String() == "complete" {

			createdAt := pay.Get("created_at").Int()
			ts := time.Unix(createdAt,0)
			amount, _ := Mstoi(pay.Get("amount_sent_msat").String())

			decoded := decodePay(ui, pay.Get("bolt11").String())
			destination := decoded.Get("payee").String()
			payee := getNode(ui, destination)
			description := decoded.Get("description").String()
			activity.SetCell(idx + rowOffset, 0,
				tview.NewTableCell( "[grey]" + ts.Format("2 Jan 2006 - 15:04")).SetAlign(tview.AlignRight))
			activity.SetCell(idx + rowOffset, 1,
				tview.NewTableCell("[green]sent to").SetAlign(tview.AlignCenter))
			activity.SetCell(idx + rowOffset, 2,
				tview.NewTableCell("[grey]" + payee.alias).SetAlign(tview.AlignRight))
			activity.SetCell(idx + rowOffset, 3,
				tview.NewTableCell("[yellow]" + formatSats(amount / 1000)).SetAlign(tview.AlignRight))
			activity.SetCell(idx + rowOffset, 4,
				tview.NewTableCell("[white]" + description).SetAlign(tview.AlignRight))

			idx += 1
		}
	}
	dash := tview.NewFlex()
	dashLeft := tview.NewFlex()

	dashLeft.SetDirection(tview.FlexRow)
	dashLeft.AddItem(infoPane, 0, 2, false)
	dashLeft.AddItem(fundsPane, 0, 1, false)
	dashLeft.AddItem(feesPane, 0, 1, false)

	dash.AddItem(dashLeft, 0, 1, false)
	dash.AddItem(activity, 0, 1, true)

	return dash
}

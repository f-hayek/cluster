package main

import (
	"errors"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tidwall/gjson"
	"io"
	"math"
	"sort"
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
	rows       []*InfoLine
	topPadding int
}

func NewInfoLine(label, value string) *InfoLine {
	return &InfoLine{label, value}
}
func NewInfoColumn(labelColor, valueColor string) *InfoColumn {
	c := &InfoColumn{
		labelColor: labelColor,
		valueColor: valueColor,
		rows:       nil,
		topPadding: 1,
	}
	return c
}
func (ic *InfoColumn) AddRow(label, value string) *InfoColumn {
	ic.rows = append(ic.rows, NewInfoLine(label, value))
	return ic
}
func (ic *InfoColumn) Print(w io.Writer) {
	for i := 0; i < ic.topPadding; i++ {
		fmt.Fprint(w, "\n")
	}
	for _, row := range ic.rows {
		fmt.Fprintf(w, "%s%25v: %s%s\n", ic.labelColor, row.label, ic.valueColor, row.value)
	}
}

type Activity struct {
	date        time.Time
	amount      int64
	fees        int64
	operation   string
	description string
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

func formatDesc(desc string) string {
	descLen := len(desc)
	if descLen > 50 {
		return " " + desc[0:30] + " (...) " + desc[descLen-14:]
	} else {
		return desc
	}
}
func dashPage(ui *UI) tview.Primitive {

	// Node Info
	infoPane := tview.NewTextView()
	infoPane.SetBorder(true).SetBorderColor(BorderColor).SetTitle(" Node Info ")
	infoPane.SetDynamicColors(true)

	info := getInfo(ui)
	config := getConfig(ui)

	ic := NewInfoColumn("[deepskyblue]", "[white]")
	ic.AddRow("Node alias", info.Get("alias").String())
	ic.AddRow("Node pubkey", info.Get("id").String())
	ic.AddRow("Network", info.Get("network").String())
	ic.AddRow("Blockheight", info.Get("blockheight").String())
	ic.AddRow("Bound to", info.Get("binding.0.address").String()+":"+info.Get("binding.0.port").String())
	for _, announce := range info.Get("address").Array() {
		ic.AddRow("Announce "+announce.Get("type").String(), announce.Get("address").String()+":"+announce.Get("port").String())
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

	ic.AddRow("Fees collected (sats)", "[yellow]"+formatSats(collectedFees))

	// Available Funds
	fundsPane := tview.NewTextView()
	fundsPane.SetBorder(true).SetBorderColor(BorderColor).SetTitle(" Available funds ")
	fundsPane.SetDynamicColors(true)

	funds := getFunds(ui, true) // list both confirmed and spent funds

	transactions := getTransactions(ui)

	spentFees := calculateSpentFees(transactions, funds)

	ic.AddRow("Fees spent on-chain", "[yellow]"+formatSats(spentFees))
	profitLoss := collectedFees - spentFees
	var plPrefix string
	if profitLoss > 0 {
		plPrefix = "[green]"
	} else {
		plPrefix = "[red]"
	}
	ic.AddRow("Profit/Loss", plPrefix+formatSats(profitLoss))
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

	fc.AddRow("On-chain capacity", formatSats(onChainFunds)+" [white]in [yellow]"+fmt.Sprintf("%d [white]UTXOs", numUtxo))
	fc.AddRow("Outbound LN capacity", formatSats(outboundFunds)+" [white]in [yellow]"+fmt.Sprintf("%s [white]channels", activeChannels))
	fc.AddRow("Total node worth", formatSats(onChainFunds+outboundFunds))
	fc.AddRow("Inbound LN capacity", formatSats(totalChannelFunds-outboundFunds))
	fc.AddRow("Smallest channel", formatSats(int64(minChan)))
	fc.AddRow("Biggest channel", formatSats(int64(maxChan)))
	fc.Print(fundsPane)

	// Off-chain fees
	// On-chain fees
	offChainFeesPane := tview.NewTextView()
	offChainFeesPane.SetBorder(true).SetBorderColor(BorderColor).SetTitle(" Default off-chain channel fees  ")
	offChainFeesPane.SetDynamicColors(true)

	l2Fees := NewInfoColumn("[deepskyblue]", "[orange]")
	l2Fees.AddRow("Default base fee", formatSats(config.Get("fee-base").Int()))
	l2Fees.AddRow("Default fee rate", formatSats(config.Get("fee-per-satoshi").Int()))

	l2Fees.Print(offChainFeesPane)

	// On-chain fees
	onChainFeesPane := tview.NewTextView()
	onChainFeesPane.SetBorder(true).SetBorderColor(BorderColor).SetTitle(" Current on-chain fee rates ")
	onChainFeesPane.SetDynamicColors(true)

	rates := getFeerates(ui)

	l1Fees := NewInfoColumn("[deepskyblue]", "[orange]")
	l1Fees.AddRow("Opening", formatSats(rates.Get("perkb.opening").Int()/1024)+" sat/vB")
	l1Fees.AddRow("Mutual close", "[green]"+formatSats(rates.Get("perkb.mutual_close").Int()/1024)+" sat/vB")
	l1Fees.AddRow("Unilateral close", formatSats(rates.Get("perkb.unilateral_close").Int()/1024)+" sat/vB")
	l1Fees.AddRow("Delayed to us", formatSats(rates.Get("perkb.delayed_to_us").Int()/1024)+" sat/vB")
	l1Fees.AddRow("HTLC resolution", formatSats(rates.Get("perkb.htlc_resolution").Int()/1024)+" sat/vB")
	l1Fees.AddRow("Penalty", formatSats(rates.Get("perkb.penalty").Int()/1024)+" sat/vB")
	l1Fees.AddRow("Min acceptable", formatSats(rates.Get("perkb.min_acceptable").Int()/1024)+" sat/vB")
	l1Fees.AddRow("Max acceptable", formatSats(rates.Get("perkb.max_acceptable").Int()/1024)+" sat/vB")

	l1Fees.Print(onChainFeesPane)

	// Recent activityTable

	activityTable := NewTable()
	activityTable.SetBorder(true).SetBorderColor(BorderColor)
	activityTable.SetTitle(" Recent LN activity ")

	activityTable.AddColumnHeader("\n[bold]date", tview.AlignCenter)
	activityTable.AddColumnHeader("\noperation", tview.AlignRight)
	activityTable.AddColumnHeader("\namount", tview.AlignRight)
	activityTable.AddColumnHeader("\nfees\n(sats)", tview.AlignRight)
	activityTable.AddColumnHeader("\n description", tview.AlignLeft)
	activityTable.Separator(16)

	activityTable.SetDoneFunc(func(key tcell.Key) {
		ui.FocusMenu()
	})
	activityTable.SetSelectable(true, false)

	rowOffset := activityTable.GetRowCount()
	activityTable.Select(rowOffset, 0).SetFixed(rowOffset, 4)

	// Do not allow to select the header
	activityTable.SetSelectionChangedFunc(func(row, column int) {
		if row < rowOffset {
			activityTable.Select(row+1, column)
		}
	})

	var activities []*Activity

	// pays
	pays := getPays(ui).Get("pays").Array()

	month := 24 * time.Hour * 31

	lastMonth := time.Now().Add(-month)

	for _, pay := range pays {
		// date
		date := time.Unix(pay.Get("created_at").Int(), 0)

		if date.After(lastMonth) {

			status := pay.Get("status").String()
			// only completed or pending pays for the past week

			if status == "complete" || status == "pending" {
				// amount
				amount, _ := Mstoi(pay.Get("amount_msat").String())
				amountSent, _ := Mstoi(pay.Get("amount_sent_msat").String())

				// operation
				destination := pay.Get("destination").String()
				payee := listNode(ui, destination)
				var operation string
				if destination == info.Get("id").String() {
					operation = "[greenyellow]rebalance"
				} else {
					if status == "pending" {
						operation = "[violet]pending to " + payee.alias
					} else {
						operation = "[darkviolet]sent to " + payee.alias
					}
				}

				// description
				bolt11 := pay.Get("bolt11").String()
				var bolt11Decoded gjson.Result
				var description string

				if bolt11 != "" {
					bolt11Decoded = decodePay(ui, bolt11)
					description = bolt11Decoded.Get("description").String()
				} else {
					description = pay.Get("label").String()
				}

				description = " " + formatDesc(description)

				activities = append(activities, &Activity{
					date,
					amount / 1000,
					(amountSent - amount) / 1000,
					operation,
					description,
				})
			}
		}
	}

	// invoices

	invoices := getInvoices(ui).Get("invoices").Array()

	for _, invoice := range invoices {
		// only paid invoices for the past month
		paidAt := invoice.Get("paid_at").Int()

		date := time.Unix(paidAt, 0)
		if date.After(lastMonth) {

			// amount
			amount := invoice.Get("msatoshi_received").Int()

			// operation
			operation := "[green]received"

			// description
			description := invoice.Get("description").String()

			description = " " + formatDesc(description)

			activities = append(activities, &Activity{
				date,
				amount / 1000,
				0,
				operation,
				description,
			})
		}
	}

	sort.Slice(activities, func(i, j int) bool {
		a1 := activities[i]
		a2 := activities[j]
		return a2.date.Before(a1.date)
	})
	totalFees := int64(0)

	for idx, activity := range activities {

		activityTable.SetCell(idx+rowOffset, 0,
			tview.NewTableCell("[grey] "+activity.date.Format("2006-01-02 15:04")).SetAlign(tview.AlignCenter))
		activityTable.SetCell(idx+rowOffset, 1,
			tview.NewTableCell(activity.operation).SetAlign(tview.AlignRight))
		var amountColor string
		if activity.operation == "[green]received" {
			amountColor = "[green]"
		} else {
			amountColor = "[red]"
		}
		var feesFormatted string
		if activity.fees == 0 {
			feesFormatted = ""
		} else {
			feesFormatted = "[red]" + formatSats(activity.fees)
		}
		activityTable.SetCell(idx+rowOffset, 2,
			tview.NewTableCell(amountColor+formatSats(activity.amount)).SetAlign(tview.AlignRight))
		activityTable.SetCell(idx+rowOffset, 3,
			tview.NewTableCell(feesFormatted).SetAlign(tview.AlignRight))
		activityTable.SetCell(idx+rowOffset, 4,
			tview.NewTableCell("[white]"+activity.description).SetAlign(tview.AlignLeft))

		totalFees += activity.fees
	}

	activityTable.Separator(16)

	currentRow := activityTable.GetRowCount()

	// Total inbound
	activityTable.SetCell(currentRow, 3,
		tview.NewTableCell("[red]" + formatSats(totalFees)).SetAlign(tview.AlignRight))

	dash := tview.NewFlex()
	dashLeft := tview.NewFlex()

	dashLeft.SetDirection(tview.FlexRow)
	dashLeft.AddItem(infoPane, 0, 2, false)
	dashLeft.AddItem(fundsPane, 0, 1, false)
	feesPane := tview.NewFlex()
	feesPane.SetDirection(tview.FlexColumn)
	feesPane.AddItem(offChainFeesPane, 0, 1, false)
	feesPane.AddItem(onChainFeesPane, 0, 1, false)

	dashLeft.AddItem(feesPane, 0, 1, false)

	dash.AddItem(dashLeft, 0, 1, false)
	dash.AddItem(activityTable, 0, 1, true)

	return dash
}

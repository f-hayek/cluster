package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func dualFundingPage(ui *UI) tview.Primitive {

	// Dual Funding Info
	infoPane := tview.NewTextView()
	infoPane.SetBorder(true).SetBorderColor(BorderColor).SetTitle(" Dual Funding ")
	infoPane.SetDynamicColors(true)


	config := getConfig(ui)
	configPath := config.Get("conf").String()

	dfEnabled := config.Get("experimental-dual-fund").Bool()
	var dfEnabledLabel string
	if dfEnabled {
		dfEnabledLabel = "Yes"
	} else {
		dfEnabledLabel = "No"
	}
	ic := NewInfoColumn("[deepskyblue]", "[white]")
	ic.AddRow("Dual-funding enabled", dfEnabledLabel)

	if dfEnabled {

	} else {
		// Instructions
		ic.AddRow("Instructions", "\nTo enable dual-funding add\nexperimental-dual-fund\noption to your \n" + configPath + "\nand restart c-lightning.")
	}
	ic.Print(infoPane)

	liquidityTable := NewTable()
	liquidityTable.SetBorder(true).SetBorderColor(BorderColor)
	liquidityTable.SetTitle(" Liquidity Ads ")

	liquidityTable.AddColumnHeader("\n[greenyellow]alias", tview.AlignRight)
	liquidityTable.AddColumnHeader("\n[bold]lease fee base\n(sats)", tview.AlignRight)
	liquidityTable.AddColumnHeader("\nlease fee basis", tview.AlignRight)
	liquidityTable.AddColumnHeader("\nfunding\nweight", tview.AlignRight)
	liquidityTable.AddColumnHeader("\nchannel fee\nmax base", tview.AlignRight)
	liquidityTable.AddColumnHeader("\nchannel fee\n max proportional", tview.AlignRight)
	liquidityTable.AddColumnHeader("\nLease ID", tview.AlignLeft)
	liquidityTable.Separator(10)

	liquidityTable.SetDoneFunc(func(key tcell.Key) {
		ui.FocusMenu()
	})
	liquidityTable.SetSelectable(true, false)

	rowOffset := liquidityTable.GetRowCount()
	liquidityTable.Select(rowOffset, 0).SetFixed(rowOffset, 6)

	// Do not allow to select the header
	liquidityTable.SetSelectionChangedFunc(func(row, column int) {
		if row < rowOffset {
			liquidityTable.Select(row+1, column)
		}
	})

	ads := listNodesThatWillFund(ui)

	for idx, ad := range ads {
		liquidityTable.SetCell(idx+rowOffset, 0,
			tview.NewTableCell("[greenyellow]" + ad.alias).SetAlign(tview.AlignRight))
		liquidityTable.SetCell(idx+rowOffset, 1,
			tview.NewTableCell(formatSats(ad.optionWillFund.leaseFeeBaseMsat / 1000)).SetAlign(tview.AlignRight))
		liquidityTable.SetCell(idx+rowOffset, 2,
			tview.NewTableCell(formatSats(ad.optionWillFund.leaseFeeBasis)).SetAlign(tview.AlignRight))
		liquidityTable.SetCell(idx+rowOffset, 3,
			tview.NewTableCell(formatSats(ad.optionWillFund.fundingWeight)).SetAlign(tview.AlignRight))
		liquidityTable.SetCell(idx+rowOffset, 4,
			tview.NewTableCell(formatSats(ad.optionWillFund.channelFeeMaxBaseMsat / 1000)).SetAlign(tview.AlignRight))
		liquidityTable.SetCell(idx+rowOffset, 5,
			tview.NewTableCell(formatSats(ad.optionWillFund.channelFeeMaxProportionalThousandths)).SetAlign(tview.AlignRight))
		liquidityTable.SetCell(idx+rowOffset, 6,
			tview.NewTableCell(ad.optionWillFund.compactLease).SetAlign(tview.AlignLeft))

	}

	dash := tview.NewFlex()
	dashLeft := tview.NewFlex()

	dashLeft.SetDirection(tview.FlexRow)
	dashLeft.AddItem(infoPane, 0, 2, false)

	dash.AddItem(dashLeft, 0, 1, false)
	dash.AddItem(liquidityTable, 0, 1, true)
	return dash

}
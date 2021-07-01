package main

import (
	//	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tidwall/gjson"
	"log"
	"strconv"
	"time"
)

var (
	qr *tview.TextView
)

func receivePage(ui *UI) tview.Primitive {

	qr = tview.NewTextView()
	qr.SetDynamicColors(true)
	//qr.SetRegions(true)
	qr.SetBorder(true)
	qr.SetBorderColor(MainColor)
	qr.SetTitle("QR Code")
	qr.SetTextAlign(tview.AlignCenter)

	form := tview.NewForm()
	form.AddInputField("Satoshi", "", 30, tview.InputFieldInteger, nil).
		AddInputField("Memo", "", 30, nil, nil).
		AddInputField("Timeout (seconds)", "3600", 30, tview.InputFieldInteger, nil).
		AddCheckbox("Receive on-chain", false, nil).
		AddButton("Receive", func() {
			ui.handleCreateInvoice(form, qr)
		}).
		AddButton("Cancel", func() {
			ui.pages.SwitchToPage("dash")
			ui.FocusMenu()
		})
	form.SetBorderColor(MainColor)
	form.SetBorder(true).SetTitle(" Receive funds ")

	flex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(form, 0, 2, true).
		AddItem(qr, 0, 3, false)
	return flex
}

func generateLabel() string {
	return "cluster_" + strconv.FormatInt(time.Now().Unix(), 10)
}

func (ui *UI) handleCreateInvoice(form *tview.Form, qr *tview.TextView) {
	satoshiField := form.GetFormItemByLabel("Satoshi").(*tview.InputField)
	onChain := form.GetFormItemByLabel("Receive on-chain").(*tview.Checkbox).IsChecked()
	descField := form.GetFormItemByLabel("Memo").(*tview.InputField).GetText()
	timeoutField := form.GetFormItemByLabel("Timeout (seconds)").(*tview.InputField).GetText()

	sats, err := strconv.Atoi(satoshiField.GetText())
	if err != nil {
		log.Fatal("wrong amount")
	}

	timeout, err := strconv.Atoi(timeoutField)
	if err != nil || timeout <= 0 {
		ui.log.Warn("Timeout value " + timeoutField + " is incorrect\n")
	}
	if onChain {
		newAddr := getNewAddr(ui).Get("bech32").String()
		qrs, err := QRCode(newAddr)
		if err != nil {
			ui.log.Warn("Error generating newaddr QR code: " + err.Error() + "\n")
		}
		qr.SetText("\n" + qrs)

	} else {

		inv := getInvoice(ui, map[string]interface{}{
			"msatoshi":    sats * 1000,
			"label":       generateLabel(),
			"description": descField,
			"expiry":      timeout})

		bolt11 := inv.Get("bolt11").String()
		paymentHash := inv.Get("payment_hash").String()

		qrs, err := QRCode(bolt11)
		if err != nil {
			ui.log.Warn("Error generating invoice QR code: " + err.Error() + "\n")
		}
		qr.SetText("\n" + qrs)
		// listen for invoices
		ln := NewClient(ui)
		ln.PaymentHandler = func(res gjson.Result) {
			if res.Get("payment_hash").String() == paymentHash {
				ui.log.Info("Invoice [white]" + paymentHash + " [green]PAID\n")
				ln.PaymentHandler = nil
			}
		}
		ln.ListenForInvoices()
	}

	return
}

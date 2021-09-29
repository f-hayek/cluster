package main

import (
	"fmt"
	//	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tidwall/gjson"
	"strconv"
	"time"
)

const (
	defaultTimeout = "7"
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
	qr.SetTitle(" QR Code ")
	qr.SetTextAlign(tview.AlignCenter)

	typeOptions := []string{
		"bolt11",
		"bolt12",
		"onchain",}

	initialType := 0
	form := tview.NewForm()
	form.AddDropDown("Type", typeOptions, initialType, nil).
		AddInputField("Satoshi", "", 30, tview.InputFieldInteger, nil).
		AddInputField("Memo", "", 30, nil, nil).
		AddInputField("Expires in (days)", defaultTimeout, 30, tview.InputFieldInteger, nil).
		AddButton("Receive", func() {
			ui.handleCreateInvoice(form, qr)
		}).
		AddButton("Cancel", func() {
			ui.pages.SwitchToPage("dash")
			ui.FocusMenu()
		})
	form.SetBorderColor(BorderColor)
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
	typeField := form.GetFormItemByLabel("Type").(*tview.DropDown)
	descField := form.GetFormItemByLabel("Memo").(*tview.InputField)
	timeoutField := form.GetFormItemByLabel("Expires in (days)").(*tview.InputField)

	sats, err := strconv.Atoi(satoshiField.GetText())
	if err != nil {
		ui.log.Warn("Incorrect satoshi amount: " + err.Error())
	}

	timeout, err := strconv.Atoi(timeoutField.GetText())
	if err != nil || timeout <= 0 {
		ui.log.Warn("Timeout value " + timeoutField.GetText() + " is incorrect\n")
	}

	_, selectedType := typeField.GetCurrentOption()

	switch selectedType {
	case "onchain":
		newAddr := getNewAddr(ui).Get("bech32").String()
		qrs, err := QRCode(newAddr)
		if err != nil {
			ui.log.Warn("Error generating newaddr QR code: " + err.Error() + "\n")
		}
		qr.SetText("\n" + qrs)

	case "bolt11":
		inv := getInvoice(ui, map[string]interface{}{
			"msatoshi":    sats * 1000,
			"label":       generateLabel(),
			"description": descField.GetText(),
			"expiry":      timeoutField.GetText() + "d"})

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
				ui.log.Info("Invoice [white]" + paymentHash + " ")
				ui.log.Ok("PAID\n")
				satoshiField.SetText("")
				descField.SetText("")
				timeoutField.SetText(defaultTimeout)
				ln.PaymentHandler = nil
			}
		}
		ln.ListenForInvoices()
	case "bolt12":
		ui.log.Info("Bolt12 selected\n")

		o := offer(ui, sats, descField.GetText())

		bolt12 := o.Get("bolt12").String()
		offerID := o.Get("offer_id").String()

		qrs, err := QRCode(bolt12)
		if err != nil {
			ui.log.Warn("Error generating offer QR code: " + err.Error() + "\n")
		}
		qr.SetText("\n" + qrs)
		ui.log.Info(fmt.Sprintf("Offer ID: %s", offerID))
		ui.log.Info(fmt.Sprintf("Bolt12: %s", bolt12))
	}
	return
}

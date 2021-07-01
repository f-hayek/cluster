package main

import (
	"fmt"
	decodepay "github.com/fiatjaf/ln-decodepay"
	"github.com/rivo/tview"
)

func payPage(ui *UI) *Form {
	f := NewForm().
		AddTextArea("Invoice / Payment request", "", 30).
		AddButton("Send", nil).
		AddButton("Cancel", func() {
			ui.pages.SwitchToPage("dash")
			ui.FocusMenu()
		})
	f.SetBorder(true).SetTitle("Enter invoice data").SetTitleAlign(tview.AlignLeft)
	return f
}

func validatePay(bolt11 string, lastChar rune) bool {
	_, err := decodepay.Decodepay(bolt11)
	if err != nil {
		//fmt.Println("invalid invoice")
		return true
	}
	return true
}

func sendPay(bolt11 string) {
	fmt.Println(bolt11)
}

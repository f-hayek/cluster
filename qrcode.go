package main

import (
	"fmt"
	"github.com/skip2/go-qrcode"
)

// QRCode generates a QR code and prints it on the command line.
func QRCode(content string) (string, error) {
	qr, err := qrcode.New(content, qrcode.Low)

	qr.DisableBorder = true

	if err != nil {
		fmt.Print("error")
		return "", err
	}
	return qr.ToSmallString(false), nil
}

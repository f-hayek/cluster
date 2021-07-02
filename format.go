package main

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func formatSats(v int64) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%d", v)
}

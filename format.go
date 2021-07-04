package main

import (
	"errors"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"strconv"
	"strings"
)

func formatSats(v int64) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%d", v)
}

func Mstoi(msats string) (int64, error) {
	s1 := strings.Replace(msats, "msat", "", 1)
	s2, err := strconv.Atoi(s1)
	if err != nil {
		return 0, errors.New("Can't convert " + msats + " to int")
	}
	return int64(s2), nil
}
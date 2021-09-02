package main

import (
	"fmt"
	"github.com/rivo/tview"
)

type Log struct {
	view *tview.TextView
	buf  []string
	c    chan string
}

func NewLog() *Log {

	v := tview.NewTextView()
	v.SetTitle(" Activity ")
	v.SetBorder(true)
	v.SetDynamicColors(true)
	v.SetBorderColor(MainColor)
	v.SetTextColor(TextColor)
	v.SetScrollable(false)
	log := &Log{
		view: v,
		buf:  []string{"", "", "", "", ""},
		c:    make(chan string),
	}
	go log.Start()
	return log
}

func (l *Log) Info(message string) {
	l.c <- "[deepskyblue]" + message
}

func (l *Log) Warn(message string) {
	l.c <- "[red]" + message
}

func (l *Log) Ok(message string) {
	l.c <- "[green]" + message
}

func (l *Log) Start() {
	for m := range l.c {
		fmt.Fprintf(l.view, "%s", m)
	}
}

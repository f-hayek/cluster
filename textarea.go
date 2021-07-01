package main

import (
	//	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TextArea struct {
	*tview.Box
	// the text that was entered
	text string
	// the text to be displayed before the text area
	label string
	// placeholder
	placeholder string

	// The label color.
	labelColor tcell.Color

	// The background color of the input area.
	fieldBackgroundColor tcell.Color

	// The text color of the input area.
	fieldTextColor tcell.Color

	// The text color of the placeholder.
	placeholderTextColor tcell.Color

	// The screen width of the label area. A value of 0 means use the width of
	// the label text.
	labelWidth int

	// The screen width of the input area. A value of 0 means extend as much as
	// possible.
	fieldWidth int

	// cursor pos
	cursorPos int

	// A callback function set by the Form class and called when the user leaves
	// this form item.
	finished func(tcell.Key)
}

func NewTextArea() *TextArea {
	return &TextArea{
		Box:                  tview.NewBox().SetBorderPadding(1, 1, 1, 1),
		labelColor:           tcell.ColorWhite,
		fieldBackgroundColor: tcell.ColorBlack,
		fieldTextColor:       tcell.ColorWhite,
		placeholderTextColor: tcell.ColorGrey,
	}
}

func (ta *TextArea) GetLabel() string {
	return ta.label
}

func (ta *TextArea) SetFormAttributes(labelWidth int, labelColor, bgColor, fieldTextColor, fieldBgColor tcell.Color) tview.FormItem {
	ta.labelWidth = labelWidth
	ta.labelColor = labelColor
	//ta.backgroundColor = bgColor
	ta.fieldTextColor = fieldTextColor
	ta.fieldBackgroundColor = fieldBgColor
	return ta
}

func (ta *TextArea) GetFieldWidth() int {
	return 20
}

func (ta *TextArea) SetFinishedFunc(handler func(key tcell.Key)) tview.FormItem {
	ta.finished = handler
	return ta
}

func (ta *TextArea) SetLabel(label string) *TextArea {
	ta.label = label
	return ta
}
func (ta *TextArea) SetText(value string) *TextArea {
	ta.text = value
	return ta
}
func (ta *TextArea) SetFieldWidth(fieldWidth int) *TextArea {
	ta.fieldWidth = fieldWidth
	return ta
}

// Draw draws this primitive onto the screen.
func (ta *TextArea) Draw(screen tcell.Screen) {
	ta.Box.DrawForSubclass(screen, ta)
}

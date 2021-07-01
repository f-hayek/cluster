package main

import (
	"github.com/rivo/tview"
)

type Form struct {
	*tview.Form
	items []tview.FormItem
}

func NewForm() *Form {
	f := &Form{
		Form: tview.NewForm(),
	}

	return f
}

func (f *Form) AddTextArea(label, value string, fieldWidth int) *Form {
	f.items = append(f.items, NewTextArea().
		SetLabel(label).
		SetText(value).
		SetFieldWidth(fieldWidth))
	return f
}

func (f *Form) AddButton(text string, selected func()) *Form {
	f.Form.AddButton(text, selected)
	return f
}

func (f *Form) SetBorder(flag bool) *Form {
	f.Form.SetBorder(flag)
	return f
}
func (f *Form) SetTitle(title string) *Form {
	f.Form.SetTitle(title)
	return f
}

func (f *Form) SetTitleAlign(align int) *Form {
	f.Form.SetTitleAlign(align)
	return f
}

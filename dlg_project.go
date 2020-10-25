package main

import (
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
	"strings"
)

type DlgProject struct {
	*walk.Dialog

	isModify bool
	updated  func(item *CfgItem)
	cfg      *Cfg
	cfgItem  *CfgItem

	oldText    string
	saveButton *walk.PushButton
	nameEdit   *walk.LineEdit
}

func (s *DlgProject) Init(owner walk.Form) error {
	if s.cfgItem == nil {
		s.cfgItem = NewCfgItem()
	}
	s.oldText = s.cfgItem.Name

	fontEdit := declarative.Font{PointSize: 12}
	dlg := declarative.Dialog{
		AssignTo: &s.Dialog,
		Layout:   declarative.VBox{},
		Children: []declarative.Widget{
			declarative.Composite{
				Layout: declarative.HBox{},
				Children: []declarative.Widget{
					declarative.TextLabel{
						Text: "项目名称:",
					},
					declarative.LineEdit{
						MinSize:       declarative.Size{Width: 150},
						Font:          fontEdit,
						Text:          s.cfgItem.Name,
						AssignTo:      &s.nameEdit,
						OnTextChanged: s.OnInputChanged,
					},
				},
			},

			declarative.PushButton{
				Text: "保存",
				Font: declarative.Font{
					PointSize: 16,
				},
				Enabled:   false,
				AssignTo:  &s.saveButton,
				OnClicked: s.OnSave,
			},
		},
	}

	if s.isModify {
		dlg.Title = "修改项目"
	} else {
		dlg.Title = "新建项目"
	}

	return dlg.Create(owner)
}

func (s *DlgProject) ShowModal() {
	s.Run()
}

func (s *DlgProject) OnInputChanged() {
	text := s.nameEdit.Text()
	if len(text) == 0 || 0 == strings.Compare(text, s.oldText) {
		s.saveButton.SetEnabled(false)
	} else {
		if s.cfg.GetItem(text) == nil {
			s.saveButton.SetEnabled(true)
		} else {
			s.saveButton.SetEnabled(false)
		}
	}
}

func (s *DlgProject) OnSave() {
	s.cfgItem.Name = s.nameEdit.Text()

	if !s.isModify {
		s.cfg.Items = append(s.cfg.Items, s.cfgItem)
	}

	if s.updated != nil {
		s.updated(s.cfgItem)
	}

	s.Close(0)
}

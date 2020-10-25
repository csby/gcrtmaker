package main

import (
	"fmt"
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
	"strings"
)

type DlgHost struct {
	*walk.Dialog

	cfgItem      *CfgItem
	listBox      *walk.ListBox
	LineEdit     *walk.LineEdit
	removeButton *walk.PushButton
	model        *HostModel
}

func (s *DlgHost) Init(owner walk.Form) error {
	fontEdit := declarative.Font{PointSize: 12}
	s.model = &HostModel{cfgItem: s.cfgItem}

	dlg := declarative.Dialog{
		AssignTo: &s.Dialog,
		Title:    "主机信息",
		MinSize:  declarative.Size{Width: 600, Height: 420},
		MaxSize:  declarative.Size{Width: 0, Height: 450},
		Size:     declarative.Size{Width: 820, Height: 420},
		Layout:   declarative.VBox{},
		Children: []declarative.Widget{
			// tool
			declarative.Composite{
				MaxSize: declarative.Size{Width: 0, Height: 120},
				Layout:  declarative.Grid{Rows: 0, Columns: 3},
				Children: []declarative.Widget{
					declarative.LineEdit{
						AssignTo:    &s.LineEdit,
						Row:         0,
						Column:      0,
						Font:        fontEdit,
						ToolTipText: "域名或IP地址",
						Text:        "localhost",
					},
					declarative.PushButton{
						Row:       0,
						Column:    1,
						Text:      "添加",
						OnClicked: s.OnAddItem,
					},
					declarative.PushButton{
						AssignTo:  &s.removeButton,
						Row:       0,
						Column:    2,
						Text:      "删除",
						Enabled:   false,
						OnClicked: s.OnRemoveItem,
					},
				},
			},

			// list
			declarative.Composite{
				Layout: declarative.VBox{},
				Children: []declarative.Widget{
					declarative.ListBox{
						AssignTo:              &s.listBox,
						Font:                  fontEdit,
						Model:                 s.model,
						OnCurrentIndexChanged: s.OnIndexChanged,
					},
				},
			},
		},
	}

	return dlg.Create(owner)
}

func (s *DlgHost) ShowModal() {
	s.Run()
}

func (s *DlgHost) OnIndexChanged() {
	index := s.listBox.CurrentIndex()
	if index < 0 {
		return
	}
	s.removeButton.SetEnabled(true)
}

func (s *DlgHost) OnAddItem() {
	text := s.LineEdit.Text()
	err := s.model.AddItem(text)
	if err != nil {
		walk.MsgBox(&s.FormBase, "添加主机失败", err.Error(), walk.MsgBoxIconError)
		return
	}

	s.listBox.SetModel(s.model)
	s.removeButton.SetEnabled(false)
}

func (s *DlgHost) OnRemoveItem() {
	if s.model.RemoveItem(s.listBox.CurrentIndex()) > 0 {
		s.listBox.SetModel(s.model)
		s.removeButton.SetEnabled(false)
	}
}

type HostModel struct {
	walk.ListModelBase
	cfgItem *CfgItem
}

func (s *HostModel) ItemCount() int {
	if s.cfgItem == nil {
		return 0
	}
	return len(s.cfgItem.Crt.Hosts)
}

func (s *HostModel) Value(index int) interface{} {
	return s.cfgItem.Crt.Hosts[index]
}

func (s *HostModel) AddItem(item string) error {
	host := strings.TrimSpace(item)
	if len(host) < 1 {
		return fmt.Errorf("主机不能为空")
	}

	if s.cfgItem.Crt.Hosts == nil {
		s.cfgItem.Crt.Hosts = make([]string, 0)
	}
	count := len(s.cfgItem.Crt.Hosts)
	for index := 0; index < count; index++ {
		if strings.ToLower(host) == strings.ToLower(s.cfgItem.Crt.Hosts[index]) {
			return fmt.Errorf("主机 '%s' 已存在", host)
		}
	}
	s.cfgItem.Crt.Hosts = append(s.cfgItem.Crt.Hosts, host)

	return nil
}

func (s *HostModel) RemoveItem(index int) int {
	if index < 0 {
		return 0
	}
	count := s.ItemCount()
	if index >= count {
		return 0
	}

	removeCount := 0
	items := make([]string, 0)
	for i := 0; i < count; i++ {
		if i == index {
			removeCount++
			continue
		}
		items = append(items, s.cfgItem.Crt.Hosts[i])
	}
	if removeCount > 0 {
		s.cfgItem.Crt.Hosts = items
	}

	return removeCount
}

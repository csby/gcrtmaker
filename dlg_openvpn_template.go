package main

import (
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
	"io/ioutil"
)

type DlgOpenVpnTemplate struct {
	*walk.Dialog

	cfg     *Cfg
	cfgPath string
	cfgItem *CfgItem

	dbOpenVpn *walk.DataBinder
}

func (s *DlgOpenVpnTemplate) Init(owner walk.Form) error {
	dlg := declarative.Dialog{
		AssignTo: &s.Dialog,
		Title:    "OpenVPN参数",
		MinSize:  declarative.Size{Width: 860, Height: 620},
		Size:     declarative.Size{Width: 960, Height: 420},
		Layout:   declarative.VBox{Alignment: declarative.AlignHNearVNear},
		Children: []declarative.Widget{
			declarative.Composite{
				DataBinder: declarative.DataBinder{
					AssignTo:       &s.dbOpenVpn,
					Name:           "db-open-vpn",
					DataSource:     &s.cfgItem.Vpn,
					ErrorPresenter: declarative.ToolTipErrorPresenter{},
				},
				Layout: declarative.Grid{Columns: 6},
				Children: []declarative.Widget{
					declarative.Label{
						Row:        0,
						Column:     0,
						ColumnSpan: 2,
						Text:       "DH(openssl dhparam -out dh2048.pem 2048)",
					},
					declarative.PushButton{
						Row:     0,
						Column:  2,
						MaxSize: declarative.Size{Width: 50},
						Text:    "从文件加载...",
						OnClicked: func() {
							dlg := &walk.FileDialog{
								Title:  "请选择哈夫曼参数文件",
								Filter: "diffie hellman parameters file (*.pem)|*.pem",
							}
							accepted, err := dlg.ShowOpen(&s.FormBase)
							if accepted && err == nil {
								data, err := ioutil.ReadFile(dlg.FilePath)
								if err != nil {
									walk.MsgBox(&s.FormBase, "读取文件内容", err.Error(), walk.MsgBoxIconError)
									return
								}
								s.cfgItem.Vpn.DiffieHellmanParameters = string(data)
								s.dbOpenVpn.Reset()
							}
						},
					},
					declarative.TextEdit{
						Row:        1,
						Column:     0,
						ColumnSpan: 3,
						VScroll:    true,
						Text:       declarative.Bind("DiffieHellmanParameters"),
					},

					declarative.Label{
						Row:        0,
						Column:     3,
						ColumnSpan: 2,
						Text:       "TA(openvpn --genkey --secret ta.key)",
					},
					declarative.PushButton{
						Row:     0,
						Column:  5,
						MaxSize: declarative.Size{Width: 50},
						Text:    "从文件加载...",
						OnClicked: func() {
							dlg := &walk.FileDialog{
								Title:  "请选择安全参数文件",
								Filter: "secret file (*.key)|*.key",
							}
							accepted, err := dlg.ShowOpen(&s.FormBase)
							if accepted && err == nil {
								data, err := ioutil.ReadFile(dlg.FilePath)
								if err != nil {
									walk.MsgBox(&s.FormBase, "读取文件内容", err.Error(), walk.MsgBoxIconError)
									return
								}
								s.cfgItem.Vpn.TlsAuth = string(data)
								s.dbOpenVpn.Reset()
							}
						},
					},
					declarative.TextEdit{
						Row:        1,
						Column:     3,
						ColumnSpan: 3,
						VScroll:    true,
						Text:       declarative.Bind("TlsAuth"),
					},

					declarative.Label{
						Row:        2,
						Column:     0,
						ColumnSpan: 3,
						Text:       "服务端配置",
					},
					declarative.TextEdit{
						Row:        3,
						Column:     0,
						ColumnSpan: 3,
						VScroll:    true,
						Text:       declarative.Bind("ServerTemplate"),
					},

					declarative.Label{
						Row:        2,
						Column:     3,
						ColumnSpan: 3,
						Text:       "客户端配置",
					},
					declarative.TextEdit{
						Row:        3,
						Column:     3,
						ColumnSpan: 3,
						VScroll:    true,
						Text:       declarative.Bind("ClientTemplate"),
					},

					declarative.PushButton{
						Row:        4,
						Column:     0,
						ColumnSpan: 6,
						Font: declarative.Font{
							PointSize: 15,
						},
						Text: "保存",
						OnClicked: func() {
							s.dbOpenVpn.Submit()
							err := s.cfg.SaveToFile(s.cfgPath)
							if err != nil {
								walk.MsgBox(&s.FormBase, "保存失败", err.Error(), walk.MsgBoxIconError)
							} else {
								s.Close(0)
							}
						},
					},
				},
			},
		},
	}

	return dlg.Create(owner)
}

func (s *DlgOpenVpnTemplate) ShowModal() {
	s.Run()
}

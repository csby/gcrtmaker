package main

import (
	"fmt"
	"github.com/csby/gsecurity/gcrt"
	"github.com/csby/gsecurity/grsa"
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
	"os"
	"path/filepath"
	"strings"
)

type DlgOpenVpnCrt struct {
	*walk.Dialog

	isServer bool
	cfgItem  *CfgItem
	caCrt    *gcrt.Crt
	caKey    *grsa.Private

	mainComposite *walk.Composite
	createButton  *walk.PushButton

	db          *walk.DataBinder
	Folder      string
	Name        string
	Password    string
	ExpiredDays int64
}

func (s *DlgOpenVpnCrt) Init(owner walk.Form) error {
	fontEdit := declarative.Font{PointSize: 12}
	dlg := declarative.Dialog{
		AssignTo: &s.Dialog,
		MinSize:  declarative.Size{Width: 400, Height: 200},
		Layout:   declarative.VBox{Alignment: declarative.AlignHNearVNear},
		Children: []declarative.Widget{
			declarative.Composite{
				AssignTo: &s.mainComposite,
				DataBinder: declarative.DataBinder{
					AssignTo:       &s.db,
					Name:           "db",
					DataSource:     s,
					ErrorPresenter: declarative.ToolTipErrorPresenter{},
				},
				Layout: declarative.Grid{Columns: 2},
				Children: []declarative.Widget{
					declarative.Label{
						Row:    0,
						Column: 0,
						Text:   "目录:",
					},
					declarative.LineEdit{
						Row:      0,
						Column:   1,
						Font:     fontEdit,
						ReadOnly: true,
						Text:     declarative.Bind("Folder"),
					},

					declarative.Label{
						Row:    1,
						Column: 0,
						Text:   "名称:",
					},
					declarative.LineEdit{
						Row:    1,
						Column: 1,
						Font:   fontEdit,
						Text:   declarative.Bind("Name"),
					},

					declarative.Label{
						Row:    2,
						Column: 0,
						Text:   "密码:",
					},
					declarative.LineEdit{
						Row:          2,
						Column:       1,
						Font:         fontEdit,
						PasswordMode: true,
						Text:         declarative.Bind("Password"),
					},

					declarative.Label{
						Row:    3,
						Column: 0,
						Text:   "过期:",
					},
					declarative.NumberEdit{
						Row:      3,
						Column:   1,
						Font:     fontEdit,
						Decimals: 0,
						Suffix:   " 天后",
						Value:    declarative.Bind("ExpiredDays"),
					},

					declarative.PushButton{
						Row:        4,
						Column:     1,
						ColumnSpan: 1,
						Text:       "创建",
						Font: declarative.Font{
							PointSize: 15,
						},
						AssignTo:  &s.createButton,
						OnClicked: s.onCreateCrt,
					},
				},
			},
		},
	}

	if s.isServer {
		dlg.Title = "新建服务器证书"
	} else {
		dlg.Title = "新建客户端证书"
	}

	return dlg.Create(owner)
}

func (s *DlgOpenVpnCrt) ShowModal() {
	s.Run()
}

func (s *DlgOpenVpnCrt) onCreateCrt() {
	s.db.Submit()

	name := strings.TrimSpace(s.Name)
	if len(name) < 1 {
		walk.MsgBox(&s.FormBase, "新建证书失败", "名称为空", walk.MsgBoxIconError)
		return
	}

	folder := s.Folder
	crtPath := filepath.Join(folder, fmt.Sprintf("%s.crt", name))
	_, err := os.Stat(crtPath)
	if err == nil || os.IsExist(err) {
		sv := walk.MsgBox(&s.FormBase, "新建证书", fmt.Sprintf("证书\r\n%s\r\n已存在\r\n\r\n是否继续创建并覆盖当前已存在的证书？", crtPath), walk.MsgBoxYesNo)
		if sv != 6 {
			return
		}
	}

	s.mainComposite.SetEnabled(false)
	s.createButton.SetText("创建中...")
	go func() {
		err := s.createCrt(s.Folder, name)
		if err != nil {
			walk.MsgBox(&s.FormBase, "新建证书失败", err.Error(), walk.MsgBoxIconError)
		} else {
			walk.MsgBox(&s.FormBase, "新建证书成功", crtPath, walk.MsgBoxIconInformation)
		}

		s.createButton.SetText("创建")
		s.mainComposite.SetEnabled(true)
	}()
}

func (s *DlgOpenVpnCrt) createCrt(folder, name string) error {
	caText, err := s.caCrt.ToMemory()
	if err != nil {
		return err
	}

	private := &grsa.Private{}
	err = private.Create(2048)
	if err != nil {
		return err
	}
	public, err := private.Public()
	if err != nil {
		return err
	}
	keyText, err := private.ToMemory(s.Password)
	if err != nil {
		return err
	}

	crtTemplate := &gcrt.Template{
		OrganizationalUnit: s.Name,
		CommonName:         s.Name,
		ExpiredDays:        s.ExpiredDays,
	}
	if s.isServer {
		crtTemplate.Organization = "server"
	} else {
		crtTemplate.Organization = "client"
	}
	template, err := crtTemplate.Template()
	if err != nil {
		return err
	}
	crt := &gcrt.Crt{}
	err = crt.Create(template, s.caCrt.Certificate(), public, s.caKey)
	if err != nil {
		return err
	}

	err = os.MkdirAll(folder, 0777)
	if err != nil {
		return err
	}
	crtPath := filepath.Join(folder, fmt.Sprintf("%s.crt", name))
	err = crt.ToFile(crtPath)
	if err != nil {
		return err
	}
	crtText, err := crt.ToMemory()
	if err != nil {
		return err
	}

	tb := &TextBuilder{}
	path := filepath.Join(folder, fmt.Sprintf("%s.ovpn", s.Name))
	if s.isServer {
		keyPath := filepath.Join(folder, fmt.Sprintf("%s.key", s.Name))
		err = private.ToFile(keyPath, s.Password)
		if err != nil {
			return err
		}
		path = filepath.Join(folder, fmt.Sprintf("%s.conf", s.Name))
		tb.AddLine(s.cfgItem.Vpn.ServerTemplate)
	} else {
		tb.AddLine(s.cfgItem.Vpn.ClientTemplate)
	}

	if len(s.cfgItem.Vpn.TlsAuth) > 2 {
		tb.AddLine("")
		if s.isServer {
			tb.AddLine("key-direction 0")
		} else {
			tb.AddLine("key-direction 1")
		}
		tb.AddLine("<tls-auth>")
		tb.WriteString(s.cfgItem.Vpn.TlsAuth)
		tb.AddLine("</tls-auth>")
	}

	tb.AddLine("")
	tb.AddLine("<ca>")
	tb.WriteString(string(caText))
	tb.AddLine("</ca>")

	tb.AddLine("")
	tb.AddLine("<cert>")
	tb.WriteString(string(crtText))
	tb.AddLine("</cert>")

	tb.AddLine("")
	tb.AddLine("<key>")
	tb.WriteString(string(keyText))
	tb.AddLine("</key>")

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = fmt.Fprint(file, tb.String())

	return err
}

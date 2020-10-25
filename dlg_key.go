package main

import (
	"fmt"
	"github.com/csby/gsecurity/grsa"
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
	"path/filepath"
	"strings"
)

type DlgKey struct {
	*walk.Dialog

	cfg     *Cfg
	cfgPath string
	cfgItem *CfgItem
	dbCa    *walk.DataBinder
	dbKey   *walk.DataBinder

	mainComposite *walk.Composite
	createButton  *walk.PushButton
	outFolder     *walk.LineEdit
	outName       *walk.LineEdit
	keyLength     *walk.NumberEdit
	keyPassword   *walk.LineEdit
}

func (s *DlgKey) Init(owner walk.Form) error {
	fontEdit := declarative.Font{PointSize: 12}

	dlg := declarative.Dialog{
		AssignTo: &s.Dialog,
		Title:    "新建CA私钥",
		MinSize:  declarative.Size{Width: 600, Height: 320},
		MaxSize:  declarative.Size{Width: 0, Height: 450},
		Size:     declarative.Size{Width: 850, Height: 350},
		Layout:   declarative.VBox{},
		Children: []declarative.Widget{
			declarative.Composite{
				Enabled:  true,
				AssignTo: &s.mainComposite,
				Layout:   declarative.VBox{},
				Children: []declarative.Widget{
					// key
					declarative.Composite{
						Background: declarative.SolidColorBrush{
							Color: walk.RGB(250, 250, 250),
						},
						MaxSize: declarative.Size{Width: 0, Height: 160},
						Layout:  declarative.Grid{Rows: 0, Columns: 3},
						Children: []declarative.Widget{
							newLabel("输出目录:", 0, 0),
							declarative.LineEdit{
								AssignTo: &s.outFolder,
								Row:      0,
								Column:   1,
								Font:     fontEdit,
								ReadOnly: true,
							},
							declarative.PushButton{
								Row:    0,
								Column: 2,
								Text:   "浏览...",
								OnClicked: func() {
									dlg := &walk.FileDialog{
										Title: "请选择私钥文件保存目录",
										//InitialDirPath: s.outFolder.Text(),
									}
									accepted, err := dlg.ShowBrowseFolder(&s.FormBase)
									if accepted && err == nil {
										s.outFolder.SetText(dlg.FilePath)
									}
								},
							},
							newLabel("文件名称:", 1, 0),
							declarative.LineEdit{
								AssignTo: &s.outName,
								Row:      1,
								Column:   1,
								Font:     fontEdit,
								Text:     "ca",
							},
							newLabel("密钥长度:", 2, 0),
							declarative.NumberEdit{
								AssignTo:  &s.keyLength,
								Row:       2,
								Column:    1,
								Font:      fontEdit,
								Decimals:  0,
								MinValue:  1024,
								MaxValue:  10240,
								Increment: 1024,
								Suffix:    " 位",
							},
							newLabel("私钥密码:", 3, 0),
							declarative.LineEdit{
								AssignTo:     &s.keyPassword,
								Row:          3,
								Column:       1,
								Font:         fontEdit,
								PasswordMode: true,
							},
						},
					},
					// button
					declarative.Composite{
						Layout: declarative.HBox{MarginsZero: true},
						Children: []declarative.Widget{
							declarative.HSpacer{},
							declarative.PushButton{
								Text: "创建",
								Font: declarative.Font{
									PointSize: 15,
								},
								AssignTo:  &s.createButton,
								OnClicked: s.OnCreateKey,
							},
							declarative.PushButton{
								Text: "关闭",
								Font: declarative.Font{
									PointSize: 15,
								},
								OnClicked: func() {
									s.Close(0)
								},
							},
						},
					},
				},
			},
		},
	}

	return dlg.Create(owner)
}

func (s *DlgKey) ShowModal(outputFolder string) {
	folder := ""
	if s.cfg != nil {
		if s.cfgItem.Ca.KeyFile != "" {
			folder = filepath.Dir(s.cfgItem.Ca.KeyFile)
		} else if s.cfgItem.Ca.CrtFile != "" {
			folder = filepath.Dir(s.cfgItem.Ca.CrtFile)
		}
	}

	if folder == "" || folder == "." {
		folder = outputFolder
	}

	if folder == "" {
		if s.cfgPath != "" {
			folder = filepath.Join(filepath.Dir(s.cfgPath), "crt")
		}
	}
	s.outFolder.SetText(folder)
	s.keyLength.SetValue(2048)

	s.Run()
}

func (s *DlgKey) OnCreateKey() {
	s.mainComposite.SetEnabled(false)
	s.createButton.SetText("创建中...")

	go func() {
		path, password, err := s.createKey()
		if err != nil {
			walk.MsgBox(&s.FormBase, "新建CA私钥失败", err.Error(), walk.MsgBoxIconError)
		} else {
			s.cfgItem.Ca.KeyFile = path
			s.cfgItem.Ca.Password().Set(password)
			s.dbKey.Reset()
			s.dbCa.Reset()
			s.cfg.SaveToFile(s.cfgPath)
			walk.MsgBox(&s.FormBase, "新建CA私钥成功", path, walk.MsgBoxIconInformation)
		}

		s.createButton.SetText("创建")
		s.mainComposite.SetEnabled(true)
	}()

}

func (s *DlgKey) createKey() (string, string, error) {
	folder, err := filepath.Abs(s.outFolder.Text())
	if err != nil {
		return "", "", fmt.Errorf("输出目录无效: %v", err)
	}
	name := strings.TrimSpace(s.outName.Text())
	if name == "" {
		return "", "", fmt.Errorf("文件名称为空")
	}
	path := filepath.Join(folder, fmt.Sprintf("%s.key", name))
	length := int(s.keyLength.Value())
	password := s.keyPassword.Text()

	rsaPrivate := &grsa.Private{}
	err = rsaPrivate.Create(length)
	if err != nil {
		return "", "", err
	}
	err = rsaPrivate.ToFile(path, password)
	if err != nil {
		return "", "", err
	}

	return path, password, nil
}

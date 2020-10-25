package main

import (
	"fmt"
	"github.com/csby/gsecurity/gcrt"
	"github.com/csby/gsecurity/grsa"
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
	"path/filepath"
	"strings"
)

type DlgCa struct {
	*walk.Dialog

	cfg     *Cfg
	cfgPath string
	cfgItem *CfgItem
	dbCa    *walk.DataBinder
	dbKey   *walk.DataBinder

	mainComposite      *walk.Composite
	outFolder          *walk.LineEdit
	outName            *walk.LineEdit
	organizationalUnit *walk.LineEdit
	commonName         *walk.LineEdit
	password           *walk.LineEdit
	expiredDays        *walk.NumberEdit
	createButton       *walk.PushButton
}

func (s *DlgCa) Init(owner walk.Form) error {
	fontEdit := declarative.Font{PointSize: 12}

	dlg := declarative.Dialog{
		AssignTo: &s.Dialog,
		Title:    "新建CA证书",
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
						DataBinder: declarative.DataBinder{
							AssignTo:       &s.dbKey,
							Name:           "db-ca-key",
							DataSource:     &s.cfgItem.Ca,
							ErrorPresenter: declarative.ToolTipErrorPresenter{},
						},
						Background: declarative.SolidColorBrush{
							Color: walk.RGB(250, 250, 250),
						},
						MaxSize: declarative.Size{Width: 0, Height: 80},
						Layout:  declarative.Grid{Rows: 0, Columns: 3},
						Children: []declarative.Widget{
							newLabel("私钥文件:", 0, 0),
							declarative.LineEdit{
								Row:      0,
								Column:   1,
								Font:     fontEdit,
								ReadOnly: true,
								Text:     declarative.Bind("KeyFile"),
							},
							declarative.PushButton{
								Row:    0,
								Column: 2,
								Text:   "浏览...",
								OnClicked: func() {
									dlg := &walk.FileDialog{
										Title:  "请选择私钥文件",
										Filter: "key file (*.key)|*.key|pem file (*.pem)|*.pem",
									}
									accepted, err := dlg.ShowOpen(&s.FormBase)
									if accepted && err == nil {
										s.cfgItem.Ca.KeyFile = dlg.FilePath
										s.dbKey.Reset()
									}
								},
							},
							newLabel("私钥密码:", 1, 0),
							declarative.LineEdit{
								AssignTo:     &s.password,
								Row:          1,
								Column:       1,
								Font:         fontEdit,
								PasswordMode: true,
								Text:         declarative.Bind("Password"),
							},
							declarative.PushButton{
								Row:       1,
								Column:    2,
								Text:      "新建...",
								OnClicked: s.ShowKey,
							},
						},
					},
					// crt
					declarative.Composite{
						Background: declarative.SolidColorBrush{
							Color: walk.RGB(250, 250, 250),
						},
						MaxSize: declarative.Size{Width: 0, Height: 200},
						Layout:  declarative.Grid{Rows: 0, Columns: 3},
						Children: []declarative.Widget{
							newLabel("输出目录:", 0, 0),
							declarative.LineEdit{
								Row:      0,
								Column:   1,
								AssignTo: &s.outFolder,
								Font:     fontEdit,
								ReadOnly: true,
							},
							declarative.PushButton{
								Row:    0,
								Column: 2,
								Text:   "浏览...",
								OnClicked: func() {
									dlg := &walk.FileDialog{
										Title: "请选择证书文件保存目录",
										//InitialDirPath: frame.cfg.Crt.Folder,
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
							newLabel("证书标识:", 2, 0),
							declarative.LineEdit{
								AssignTo: &s.organizationalUnit,
								Row:      2,
								Column:   1,
								Font:     fontEdit,
							},
							newLabel("显示名称:", 3, 0),
							declarative.LineEdit{
								AssignTo: &s.commonName,
								Row:      3,
								Column:   1,
								Font:     fontEdit,
							},
							newLabel("有效期:", 4, 0),
							declarative.NumberEdit{
								AssignTo: &s.expiredDays,
								Row:      4,
								Column:   1,
								Font:     fontEdit,
								Decimals: 0,
								MinValue: 1,
								MaxValue: 36500,
								Suffix:   " 天",
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
								OnClicked: s.OnCreateCrt,
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

func (s *DlgCa) ShowModal(outputFolder string) {
	folder := ""
	if s.cfg != nil {
		if s.cfgItem.Ca.CrtFile != "" {
			folder = filepath.Dir(s.cfgItem.Ca.CrtFile)
		} else if s.cfgItem.Ca.KeyFile != "" {
			folder = filepath.Dir(s.cfgItem.Ca.KeyFile)
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

	crt := &gcrt.Crt{}
	err := crt.FromFile(s.cfgItem.Ca.CrtFile)
	if err == nil {
		s.organizationalUnit.SetText(crt.OrganizationalUnit())
		s.commonName.SetText(crt.CommonName())
	}

	s.expiredDays.SetValue(3650)

	s.Run()
}

func (s *DlgCa) ShowKey() {
	dlg := &DlgKey{cfg: s.cfg, cfgPath: s.cfgPath, dbCa: s.dbCa, dbKey: s.dbKey, cfgItem: s.cfgItem}
	err := dlg.Init(&s.FormBase)
	if err != nil {
		fmt.Println(err)
		return
	}
	dlg.ShowModal(s.outFolder.Text())
}

func (s *DlgCa) OnCreateCrt() {
	s.mainComposite.SetEnabled(false)
	s.createButton.SetText("创建中...")

	go func() {
		path, password, err := s.createCrt()
		if err != nil {
			walk.MsgBox(&s.FormBase, "新建CA证书失败", err.Error(), walk.MsgBoxIconError)
		} else {
			s.cfgItem.Ca.CrtFile = path
			s.cfgItem.Ca.Password().Set(password)
			s.dbCa.Reset()
			s.cfg.SaveToFile(s.cfgPath)
			walk.MsgBox(&s.FormBase, "新建CA证书成功", path, walk.MsgBoxIconInformation)
		}

		s.createButton.SetText("创建")
		s.mainComposite.SetEnabled(true)
	}()

}

func (s *DlgCa) createCrt() (string, string, error) {
	folder, err := filepath.Abs(s.outFolder.Text())
	if err != nil {
		return "", "", fmt.Errorf("输出目录无效: %v", err)
	}
	name := strings.TrimSpace(s.outName.Text())
	if name == "" {
		return "", "", fmt.Errorf("文件名称为空")
	}
	path := filepath.Join(folder, fmt.Sprintf("%s.crt", name))

	ou := s.organizationalUnit.Text()
	if ou == "" {
		return "", "", fmt.Errorf("证书标识为空")
	}

	keyPath := s.cfgItem.Ca.KeyFile
	if keyPath == "" {
		return "", "", fmt.Errorf("私钥文件为空")
	}
	keyPassword := s.password.Text()
	rsaPrivate := &grsa.Private{}
	err = rsaPrivate.FromFile(keyPath, keyPassword)
	if err != nil {
		return "", "", fmt.Errorf("加载私钥文件错误: %v", err)
	}
	rsaPublic, err := rsaPrivate.Public()
	if err != nil {
		return "", "", err
	}
	crtTemplate := &gcrt.Template{
		Organization:       "ca",
		OrganizationalUnit: ou,
		CommonName:         s.commonName.Text(),
		ExpiredDays:        int64(s.expiredDays.Value()),
	}
	template, err := crtTemplate.Template()
	if err != nil {
		return "", "", err
	}
	crt := &gcrt.Crt{}
	err = crt.Create(template, template, rsaPublic, rsaPrivate)
	if err != nil {
		return "", "", err
	}
	err = crt.ToFile(path)
	if err != nil {
		return "", "", err
	}

	return path, keyPassword, nil
}

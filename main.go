package main

import (
	"fmt"
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
	"os"
	"path/filepath"
)

const (
	moduleName    = "gcrtmaker"
	moduleRemark  = "证书工具"
	moduleVersion = "1.0.1.0"
)

func main() {
	frame := new(Frame)
	frame.CenterScreen = true
	filePath, err := filepath.Abs(os.Args[0])
	if err != nil {
		frame.cfgPath = fmt.Sprintf("%s.json", moduleName)
	} else {
		frame.cfgPath = filepath.Join(filepath.Dir(filePath), fmt.Sprintf("%s.json", moduleName))
		frame.RootFolder = filepath.Join(filepath.Dir(filePath), "crts")
	}
	frame.dbCa = &walk.DataBinder{}
	frame.dbCrt = &walk.DataBinder{}
	frame.cfg = NewCfg()
	frame.cfg.LoadFromFile(frame.cfgPath)
	frame.cfgItem = frame.cfg.GetItem(frame.cfg.DefaultItemName)
	if frame.cfgItem == nil {
		frame.cfgItem = NewCfgItem()
		frame.cfgItem.Name = "项目1"
		if frame.cfg.Items == nil {
			frame.cfg.Items = make([]*CfgItem, 0)
		}
		frame.cfg.DefaultItemName = frame.cfgItem.Name
		frame.cfg.Items = append(frame.cfg.Items, frame.cfgItem)
	}
	if len(frame.cfgItem.Crt.RootFolder) < 1 {
		frame.cfgItem.Crt.RootFolder = filepath.Join(frame.RootFolder, frame.cfgItem.Name)
	}

	fontEdit := declarative.Font{PointSize: 12}
	mw := &declarative.MainWindow{
		AssignTo:        &frame.MainWindow,
		Title:           fmt.Sprintf("%s %s", moduleRemark, moduleVersion),
		MinSize:         declarative.Size{Width: 600, Height: 400},
		Size:            declarative.Size{Width: 800, Height: 520},
		OnBoundsChanged: frame.OnBoundsChanged,
		Layout:          declarative.VBox{},
		ToolBar: declarative.ToolBar{
			ButtonStyle: declarative.ToolBarButtonTextOnly,
			AssignTo:    &frame.toolBar,
			Items: []declarative.MenuItem{
				declarative.Menu{
					Text: "文件",
					Items: []declarative.MenuItem{
						declarative.Action{
							Text:        "新建CA证书...",
							OnTriggered: frame.ShowCa,
						},
						declarative.Action{
							Text:        "吊销证书...",
							OnTriggered: frame.ShowCrl,
						},
						declarative.Menu{
							Text: "OpenVPN",
							Items: []declarative.MenuItem{
								declarative.Action{
									Text:        "配置...",
									OnTriggered: frame.ShowOpenVpnTemplate,
								},
								declarative.Action{
									Text:        "服务器证书...",
									OnTriggered: frame.ShowOpenVpnServerCrt,
								},
								declarative.Action{
									Text:        "客户端证书...",
									OnTriggered: frame.ShowOpenVpnClientCrt,
								},
							},
						},
					},
				},
			},
		},
		Children: []declarative.Widget{
			declarative.Composite{
				Enabled:  true,
				AssignTo: &frame.mainComposite,
				Layout:   declarative.VBox{MarginsZero: true},
				Children: []declarative.Widget{
					// project
					declarative.Composite{
						Layout: declarative.Grid{MarginsZero: true},
						Children: []declarative.Widget{
							declarative.TextLabel{
								MaxSize:       declarative.Size{Width: 74},
								Text:          "项目名称:",
								TextAlignment: declarative.AlignHFarVCenter,
							},
							declarative.ComboBox{
								Column:                1,
								ColumnSpan:            2,
								Font:                  fontEdit,
								Model:                 frame.cfg.Items,
								DisplayMember:         "Name",
								CurrentIndex:          frame.cfg.IndexOf(frame.cfgItem),
								AssignTo:              &frame.projectSelector,
								OnCurrentIndexChanged: frame.OnProjectChanged,
							},
							declarative.PushButton{
								Column:    3,
								MaxSize:   declarative.Size{Width: 50},
								Text:      "修改...",
								OnClicked: frame.ModifyProject,
							},
							declarative.PushButton{
								Column:    4,
								MaxSize:   declarative.Size{Width: 50},
								Text:      "添加...",
								OnClicked: frame.AddProject,
							},
						},
					},
					// ca
					declarative.Composite{
						DataBinder: declarative.DataBinder{
							AssignTo:       &frame.dbCa,
							Name:           "db-ca",
							DataSource:     &frame.cfgItem.Ca,
							ErrorPresenter: declarative.ToolTipErrorPresenter{},
						},
						Background: declarative.SolidColorBrush{
							Color: walk.RGB(250, 250, 250),
						},
						MaxSize: declarative.Size{Width: 0, Height: 120},
						Layout:  declarative.Grid{Rows: 0, Columns: 3},
						Children: []declarative.Widget{
							newLabel("CA证书文件:", 0, 0),
							declarative.LineEdit{
								Row:      0,
								Column:   1,
								Font:     fontEdit,
								ReadOnly: true,
								Text:     declarative.Bind("CrtFile"),
							},
							declarative.PushButton{
								Row:    0,
								Column: 2,
								Text:   "浏览...",
								OnClicked: func() {
									dlg := &walk.FileDialog{
										Title:  "请选择根证书(CA)文件",
										Filter: "certificate file (*.crt)|*.crt",
									}
									accepted, err := dlg.ShowOpen(&frame.FormBase)
									if accepted && err == nil {
										frame.cfgItem.Ca.CrtFile = dlg.FilePath
										frame.dbCa.Reset()
									}
								},
							},
							newLabel("CA私钥文件:", 1, 0),
							declarative.LineEdit{
								Row:      1,
								Column:   1,
								Font:     fontEdit,
								ReadOnly: true,
								Text:     declarative.Bind("KeyFile"),
							},
							declarative.PushButton{
								Row:    1,
								Column: 2,
								Text:   "浏览...",
								OnClicked: func() {
									dlg := &walk.FileDialog{
										Title:  "请选择根证书(CA)对应的私钥文件",
										Filter: "key file (*.key)|*.key|pem file (*.pem)|*.pem",
									}
									accepted, err := dlg.ShowOpen(&frame.FormBase)
									if accepted && err == nil {
										frame.cfgItem.Ca.KeyFile = dlg.FilePath
										frame.dbCa.Reset()
									}
								},
							},
							newLabel("CA私钥密码:", 2, 0),
							declarative.LineEdit{
								Row:          2,
								Column:       1,
								Font:         fontEdit,
								PasswordMode: true,
								Text:         declarative.Bind("Password"),
							},
							declarative.PushButton{
								Row:       2,
								Column:    2,
								Text:      "验证",
								OnClicked: frame.VerifyCa,
							},
						},
					},
					// crt
					declarative.Composite{
						DataBinder: declarative.DataBinder{
							AssignTo:       &frame.dbCrt,
							Name:           "db-crt",
							DataSource:     &frame.cfgItem.Crt,
							ErrorPresenter: declarative.ToolTipErrorPresenter{},
						},
						Background: declarative.SolidColorBrush{
							Color: walk.RGB(250, 250, 250),
						},
						MaxSize: declarative.Size{Width: 0, Height: 280},
						Layout:  declarative.Grid{Rows: 0, Columns: 5},
						Children: []declarative.Widget{
							newLabel("证书类型:", 0, 0),
							declarative.ComboBox{
								Row:                   0,
								Column:                1,
								ColumnSpan:            3,
								Font:                  fontEdit,
								Model:                 Orgs,
								DisplayMember:         "DisplayName",
								BindingMember:         "Name",
								Value:                 declarative.Bind("Organization", declarative.SelRequired{}),
								AssignTo:              &frame.kindSelector,
								OnCurrentIndexChanged: frame.OnKindChanged,
							},
							declarative.PushButton{
								Row:       0,
								Column:    4,
								Text:      "主机...",
								Visible:   false,
								AssignTo:  &frame.hostButton,
								OnClicked: frame.ShowHosts,
							},
							newLabel("输出根目录:", 1, 0),
							declarative.LineEdit{
								Row:        1,
								Column:     1,
								ColumnSpan: 3,
								Font:       fontEdit,
								ReadOnly:   true,
								Text:       declarative.Bind("RootFolder"),
							},
							declarative.PushButton{
								Row:    1,
								Column: 4,
								Text:   "浏览...",
								OnClicked: func() {
									dlg := &walk.FileDialog{
										Title: "请选择证书文件保存目录",
										//InitialDirPath: frame.cfg.Crt.Folder,
									}
									accepted, err := dlg.ShowBrowseFolder(&frame.FormBase)
									if accepted && err == nil {
										frame.cfgItem.Crt.RootFolder = dlg.FilePath
										frame.dbCrt.Reset()
									}
								},
							},
							newLabel("输出子目录:", 2, 0),
							declarative.LineEdit{
								Row:        2,
								Column:     1,
								ColumnSpan: 3,
								Font:       fontEdit,
								Text:       declarative.Bind("SubFolder"),
							},
							newLabel("文件名称:", 3, 0),
							declarative.LineEdit{
								Row:      3,
								Column:   1,
								Font:     fontEdit,
								AssignTo: &frame.crtFileName,
								Text:     declarative.Bind("Name"),
							},
							newLabel("证书密码:", 4, 0),
							declarative.LineEdit{
								Row:          4,
								Column:       1,
								Font:         fontEdit,
								PasswordMode: true,
								Text:         declarative.Bind("Password"),
								ToolTipText:  "pfx格式证书密码",
							},
							newLabel("有效期:", 5, 0),
							declarative.NumberEdit{
								Row:      5,
								Column:   1,
								Font:     fontEdit,
								Decimals: 0,
								Suffix:   " 天",
								Value:    declarative.Bind("ExpiredDays"),
							},
							newLabel("证书标识:", 6, 0),
							declarative.LineEdit{
								Row:    6,
								Column: 1,
								Font:   fontEdit,
								Text:   declarative.Bind("OrganizationalUnit"),
							},

							newLabel("显示名称:", 3, 2),
							declarative.LineEdit{
								Row:    3,
								Column: 3,
								Font:   fontEdit,
								Text:   declarative.Bind("CommonName"),
							},
							newLabel("地区:", 4, 2),
							declarative.LineEdit{
								Row:    4,
								Column: 3,
								Font:   fontEdit,
								Text:   declarative.Bind("Locality"),
							},
							newLabel("省份:", 5, 2),
							declarative.LineEdit{
								Row:    5,
								Column: 3,
								Font:   fontEdit,
								Text:   declarative.Bind("Province"),
							},
							newLabel("地址:", 6, 2),
							declarative.LineEdit{
								Row:    6,
								Column: 3,
								Font:   fontEdit,
								Text:   declarative.Bind("StreetAddress"),
							},
						},
					},
					// btn
					declarative.Composite{
						Layout: declarative.HBox{MarginsZero: true},
						Children: []declarative.Widget{
							declarative.PushButton{
								Text: "创建",
								Font: declarative.Font{
									PointSize: 15,
								},
								AssignTo:  &frame.createButton,
								OnClicked: frame.OnCreateCrt,
							},
						},
					},
				},
			},
		},
	}

	code, err := mw.Run()
	if err != nil {
		fmt.Println("error:", err)
	}
	frame.Dispose()
	fmt.Println("exit code:", code)
}

package main

import (
	"fmt"
	"github.com/csby/gsecurity/gcrt"
	"github.com/csby/gsecurity/grsa"
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
	"math/big"
	"path/filepath"
	"time"
)

type DlgCrl struct {
	*walk.Dialog

	cfg     *Cfg
	cfgPath string
	cfgItem *CfgItem
	caCrt   *gcrt.Crt
	caKey   *grsa.Private

	mainComposite  *walk.Composite
	fileEdit       *walk.LineEdit
	tv             *walk.TableView
	model          *CrlModel
	reasonLabel    *walk.TextLabel
	reasonSelector *walk.ComboBox
	btDelete       *walk.PushButton
	btAdd          *walk.PushButton
	btCreate       *walk.PushButton
}

func (s *DlgCrl) Init(owner walk.Form) error {
	fontEdit := declarative.Font{PointSize: 12}
	s.model = new(CrlModel)
	s.model.items = make([]*gcrt.RevokedItem, 0)

	dlg := declarative.Dialog{
		AssignTo: &s.Dialog,
		Title:    "新建证书吊销列表",
		MinSize:  declarative.Size{Width: 680, Height: 420},
		Size:     declarative.Size{Width: 960, Height: 380},
		Layout:   declarative.VBox{},
		Children: []declarative.Widget{
			declarative.Composite{
				Enabled:  true,
				AssignTo: &s.mainComposite,
				Layout:   declarative.VBox{},
				Children: []declarative.Widget{
					// file
					declarative.Composite{
						MinSize: declarative.Size{Width: 0, Height: 40},
						MaxSize: declarative.Size{Width: 0, Height: 50},
						Layout:  declarative.Grid{Rows: 0, Columns: 3, MarginsZero: true},
						Children: []declarative.Widget{
							newLabel("列表文件:", 0, 0),
							declarative.LineEdit{
								AssignTo: &s.fileEdit,
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
										Title:    "请选择证书吊销列表文件",
										Filter:   "certificate revoke file (*.crl)|*.crl",
										FilePath: s.fileEdit.Text(),
									}
									accepted, err := dlg.ShowSave(&s.FormBase)
									if accepted && err == nil {
										path := dlg.FilePath
										ext := filepath.Ext(path)
										if ext != ".crl" {
											extLen := len(ext)
											if extLen > 0 {
												path = path[0 : len(path)-extLen]
											}
											path = fmt.Sprintf("%s.crl", path)
										}
										s.changePath(path)
									}
								},
							},
						},
					},

					// header
					declarative.Composite{
						Layout: declarative.HBox{MarginsZero: true},
						Children: []declarative.Widget{
							declarative.TextLabel{
								Text:          "证书列表",
								TextAlignment: declarative.AlignHNearVFar,
							},
							declarative.HSpacer{},
							declarative.TextLabel{
								Text:          "吊销理由:",
								TextAlignment: declarative.AlignHNearVFar,
								AssignTo:      &s.reasonLabel,
							},
							declarative.ComboBox{
								Row:                   0,
								Column:                1,
								ColumnSpan:            3,
								Font:                  fontEdit,
								Model:                 CRLReasons,
								DisplayMember:         "DisplayName",
								BindingMember:         "Code",
								AssignTo:              &s.reasonSelector,
								OnCurrentIndexChanged: s.OnReasonChanged,
							},
							declarative.PushButton{
								AssignTo:  &s.btDelete,
								Text:      "删除",
								Visible:   false,
								OnClicked: s.deleteItem,
							},
							declarative.PushButton{
								AssignTo:  &s.btAdd,
								Text:      "添加...",
								OnClicked: s.addItem,
							},
						},
					},
					// list
					declarative.TableView{
						AssignTo: &s.tv,
						Model:    s.model,
						Columns: []declarative.TableViewColumn{
							{Title: "序列号", Frozen: true, Width: 200},
							{Title: "理由"},
							{Title: "名称"},
							{Title: "类型", Width: 65},
							{Title: "标识"},
							{Title: "有效期", Width: 100},
							{Title: "地区"},
							{Title: "省份"},
							{Title: "地址"},
						},
						OnCurrentIndexChanged: s.selectedItemChanged,
					},
					// button
					declarative.Composite{
						Layout: declarative.HBox{MarginsZero: true},
						Children: []declarative.Widget{
							declarative.HSpacer{},
							declarative.PushButton{
								AssignTo: &s.btCreate,
								Text:     "创建",
								Font: declarative.Font{
									PointSize: 15,
								},
								Enabled:   false,
								OnClicked: s.OnCreateCrl,
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

func (s *DlgCrl) ShowModal() {
	path := s.cfgItem.Crl.CrlFile
	if path == "" {
		folder := filepath.Dir(s.cfgItem.Ca.CrtFile)
		path = filepath.Join(folder, "cr.crl")
	}
	s.changePath(path)

	s.Run()
}

func (s *DlgCrl) OnCreateCrl() {
	s.mainComposite.SetEnabled(false)
	s.btCreate.SetText("创建中...")

	go func() {
		path := s.cfgItem.Crl.CrlFile
		err := s.createCrl(path)
		if err != nil {
			walk.MsgBox(&s.FormBase, "新建证书吊销列表失败", err.Error(), walk.MsgBoxIconError)
		} else {
			walk.MsgBox(&s.FormBase, "新建证书吊销列表成功", path, walk.MsgBoxIconInformation)
		}

		s.btCreate.SetText("创建")
		s.mainComposite.SetEnabled(true)
	}()
}

func (s *DlgCrl) changePath(path string) {
	s.fileEdit.SetText(path)
	if path != s.cfgItem.Crl.CrlFile {
		s.cfgItem.Crl.CrlFile = path
		s.cfg.SaveToFile(s.cfgPath)
	}
	s.btCreate.SetEnabled(len(path) > 0)

	crl := &gcrt.Crl{}
	err := crl.FromFile(path)
	if err != nil {
		return
	}
	info, err := crl.Info()
	if err != nil {
		return
	}
	s.model.items = info.Items
	s.model.PublishRowsReset()
}

func (s *DlgCrl) createCrl(path string) error {
	crl := &gcrt.Crl{}
	count := len(s.model.items)
	for i := 0; i < count; i++ {
		err := crl.AddItem(s.model.items[i])
		if err != nil {
			return err
		}
	}

	return crl.ToFile(path, s.caCrt, s.caKey, nil, nil)
}

func (s *DlgCrl) addCrt(path string) (*gcrt.Crt, error) {
	crt := &gcrt.Crt{}
	err := crt.FromFile(path)
	if err != nil {
		return nil, err
	}

	err = crt.Verify(s.caCrt)
	if err != nil {
		return nil, fmt.Errorf("证书不是当前CA签发: %v", err)
	}

	return crt, nil
}

func (s *DlgCrl) deleteCrt(index int) error {
	if index < 0 {
		return fmt.Errorf("invalid index: %d", index)
	}
	count := len(s.model.items)
	if index >= count {
		return fmt.Errorf("invalid index: %d", index)
	}

	items := make([]*gcrt.RevokedItem, 0)
	for i := 0; i < count; i++ {
		if i == index {
			continue
		}
		items = append(items, s.model.items[i])
	}
	s.model.items = items

	return nil
}

func (s *DlgCrl) selectedItemChanged() {
	itemIndex := s.tv.CurrentIndex()
	if itemIndex < 0 {
		s.btDelete.SetVisible(false)
		s.reasonLabel.SetVisible(false)
		s.reasonSelector.SetVisible(false)
	} else {
		s.btDelete.SetVisible(true)
		s.reasonLabel.SetVisible(true)
		s.reasonSelector.SetVisible(true)
		item := s.model.items[itemIndex]
		s.reasonSelector.SetCurrentIndex(crlReasonIndex(item.ReasonCode))
	}
}

func (s *DlgCrl) deleteItem() {
	err := s.deleteCrt(s.tv.CurrentIndex())
	if err != nil {
		return
	}
	s.model.PublishRowsReset()
}

func (s *DlgCrl) addItem() {
	dlg := &walk.FileDialog{
		Title:  "请选择将被吊销证书文件",
		Filter: "certificate file (*.crt)|*.crt",
	}
	accepted, err := dlg.ShowOpen(&s.FormBase)
	if err != nil || !accepted {
		return
	}
	crt, err := s.addCrt(dlg.FilePath)
	if err != nil {
		walk.MsgBox(&s.FormBase, "添加证书失败", err.Error(), walk.MsgBoxIconError)
		return
	}
	index, ok := s.model.Exist(crt.SerialNumber())
	if ok {
		walk.MsgBox(&s.FormBase, "添加证书失败", "该证书已在列表中", walk.MsgBoxIconError)
		s.tv.SetCurrentIndex(index)
		return
	}

	item := &gcrt.RevokedItem{
		SerialNumber:       crt.SerialNumber(),
		RevocationTime:     time.Now(),
		Organization:       crt.Organization(),
		OrganizationalUnit: crt.OrganizationalUnit(),
		CommonName:         crt.CommonName(),
		Locality:           crt.Locality(),
		Province:           crt.Province(),
		StreetAddress:      crt.StreetAddress(),
		NotBefore:          crt.NotBefore(),
		NotAfter:           crt.NotAfter(),
	}
	if s.model.items == nil {
		s.model.items = make([]*gcrt.RevokedItem, 0)
	}
	s.model.items = append(s.model.items, item)
	s.model.PublishRowsReset()
}

func (s *DlgCrl) OnReasonChanged() {
	selIndex := s.reasonSelector.CurrentIndex()
	if selIndex < 0 {
		return
	}

	itemIndex := s.tv.CurrentIndex()
	if itemIndex < 0 {
		return
	}
	count := len(s.model.items)
	if count < 1 {
		return
	}
	item := s.model.items[itemIndex]
	reason := CRLReasons[selIndex]
	if item.ReasonCode == reason.Code {
		return
	}

	item.ReasonCode = reason.Code
	s.model.PublishRowChanged(itemIndex)
}

type CrlModel struct {
	walk.TableModelBase

	items []*gcrt.RevokedItem
}

func (s *CrlModel) RowCount() int {
	return len(s.items)
}

func (s *CrlModel) Value(row, col int) interface{} {
	item := s.items[row]
	switch col {
	case 0:
		if item.SerialNumber == nil {
			return ""
		} else {
			return item.String()
		}
	case 1:
		return crlReasonDisplayName(item.ReasonCode)
	case 2:
		return item.CommonName
	case 3:
		return orgDisplayName(item.Organization)
	case 4:
		return item.OrganizationalUnit
	case 5:
		if item.NotAfter == nil {
			return ""
		} else {
			return item.NotAfter.Format("2006-01-02")
		}
	case 6:
		return item.Locality
	case 7:
		return item.Province
	case 8:
		return item.StreetAddress
	}

	return ""
}

func (s *CrlModel) Exist(sno *big.Int) (int, bool) {
	if sno == nil {
		return -1, false
	}
	sns := sno.Text(16)
	count := len(s.items)
	for i := 0; i < count; i++ {
		if sns == s.items[i].String() {
			return i, true
		}
	}

	return -1, false
}

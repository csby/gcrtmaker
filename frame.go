package main

import (
	"fmt"
	"github.com/csby/gsecurity/gcrt"
	"github.com/csby/gsecurity/grsa"
	"github.com/lxn/walk"
	"github.com/lxn/win"
	"os"
	"path/filepath"
	"strings"
)

type Frame struct {
	*walk.MainWindow

	CenterScreen bool
	RootFolder   string

	cfg     *Cfg
	cfgItem *CfgItem
	cfgPath string
	dbCa    *walk.DataBinder
	dbCrt   *walk.DataBinder

	firstBoundsChange bool
	toolBar           *walk.ToolBar
	mainComposite     *walk.Composite
	projectSelector   *walk.ComboBox
	kindSelector      *walk.ComboBox
	crtFileName       *walk.LineEdit
	createButton      *walk.PushButton
	hostButton        *walk.PushButton
}

func (s *Frame) OnBoundsChanged() {
	if !s.firstBoundsChange {
		s.firstBoundsChange = true

		if s.CenterScreen {
			screenWidth := int(win.GetSystemMetrics(win.SM_CXSCREEN))
			screenHeight := int(win.GetSystemMetrics(win.SM_CYSCREEN))
			frameWidth := s.Size().Width
			frameHeight := s.Size().Height
			frameBound := walk.Rectangle{
				X:      (screenWidth - frameWidth) / 2,
				Y:      (screenHeight - frameHeight) / 2,
				Width:  frameWidth,
				Height: frameHeight,
			}
			s.SetBoundsPixels(frameBound)
		}
	}
}

func (s *Frame) AddProject() {
	dlg := &DlgProject{
		isModify: false,
		updated:  s.UpdateProject,
		cfgItem:  NewCfgItem(),
		cfg:      s.cfg,
	}
	err := dlg.Init(&s.FormBase)
	if err != nil {
		fmt.Println(err)
		return
	}

	dlg.ShowModal()
}

func (s *Frame) ModifyProject() {
	dlg := &DlgProject{
		isModify: true,
		updated:  s.UpdateProject,
		cfgItem:  s.cfgItem,
		cfg:      s.cfg,
	}
	err := dlg.Init(&s.FormBase)
	if err != nil {
		fmt.Println(err)
		return
	}

	dlg.ShowModal()
}

func (s *Frame) UpdateProject(item *CfgItem) {
	if item == nil {
		return
	}
	index := s.cfg.IndexOf(item)
	if index < 0 {
		return
	}

	s.cfg.DefaultItemName = item.Name
	s.projectSelector.SetModel(s.cfg.Items)
	s.projectSelector.SetCurrentIndex(index)

	s.cfg.SaveToFile(s.cfgPath)
}

func (s *Frame) ShowHosts() {
	dlg := &DlgHost{cfgItem: s.cfgItem}
	err := dlg.Init(&s.FormBase)
	if err != nil {
		fmt.Println(err)
		return
	}
	dlg.ShowModal()
}

func (s *Frame) ShowCa() {
	dlg := &DlgCa{cfg: s.cfg, cfgPath: s.cfgPath, dbCa: s.dbCa, cfgItem: s.cfgItem}
	err := dlg.Init(&s.FormBase)
	if err != nil {
		fmt.Println(err)
		return
	}
	dlg.ShowModal(s.cfgItem.Crt.RootFolder)
}

func (s *Frame) ShowCrl() {
	caCrt, caKey, err := s.verifyCa()
	if err != nil {
		walk.MsgBox(&s.FormBase, "CA无效", err.Error(), walk.MsgBoxIconError)
		return
	}

	dlg := &DlgCrl{cfg: s.cfg, cfgPath: s.cfgPath, caCrt: caCrt, caKey: caKey, cfgItem: s.cfgItem}
	err = dlg.Init(&s.FormBase)
	if err != nil {
		fmt.Println(err)
		return
	}
	dlg.ShowModal()
}

func (s *Frame) ShowOpenVpnServerCrt() {
	caCrt, caKey, err := s.verifyCa()
	if err != nil {
		walk.MsgBox(&s.FormBase, "CA无效", err.Error(), walk.MsgBoxIconError)
		return
	}

	dlg := &DlgOpenVpnCrt{
		isServer:    true,
		caCrt:       caCrt,
		caKey:       caKey,
		cfgItem:     s.cfgItem,
		Folder:      s.cfgItem.Crt.RootFolder,
		Name:        "server",
		ExpiredDays: 3650,
	}
	err = dlg.Init(&s.FormBase)
	if err != nil {
		fmt.Println(err)
		return
	}
	dlg.ShowModal()
}

func (s *Frame) ShowOpenVpnClientCrt() {
	caCrt, caKey, err := s.verifyCa()
	if err != nil {
		walk.MsgBox(&s.FormBase, "CA无效", err.Error(), walk.MsgBoxIconError)
		return
	}

	dlg := &DlgOpenVpnCrt{
		isServer:    false,
		caCrt:       caCrt,
		caKey:       caKey,
		cfgItem:     s.cfgItem,
		Folder:      s.cfgItem.Crt.RootFolder,
		ExpiredDays: 365,
	}
	err = dlg.Init(&s.FormBase)
	if err != nil {
		fmt.Println(err)
		return
	}
	dlg.ShowModal()
}

func (s *Frame) ShowOpenVpnTemplate() {
	dlg := &DlgOpenVpnTemplate{cfg: s.cfg, cfgPath: s.cfgPath, cfgItem: s.cfgItem}
	err := dlg.Init(&s.FormBase)
	if err != nil {
		fmt.Println(err)
		return
	}
	dlg.ShowModal()
}

func (s *Frame) OnProjectChanged() {
	selIndex := s.projectSelector.CurrentIndex()
	curIndex := s.cfg.IndexOf(s.cfgItem)
	if selIndex == curIndex {
		return
	}
	if selIndex < 0 || selIndex >= len(s.cfg.Items) {
		return
	}

	s.cfgItem = s.cfg.Items[selIndex]
	if len(s.cfgItem.Crt.RootFolder) < 1 {
		s.cfgItem.Crt.RootFolder = filepath.Join(s.RootFolder, s.cfgItem.Name)
	}

	s.cfg.DefaultItemName = s.cfgItem.Name
	s.dbCa.SetDataSource(&s.cfgItem.Ca)
	s.dbCrt.SetDataSource(&s.cfgItem.Crt)

	s.dbCa.Reset()
	s.dbCrt.Reset()
}

func (s *Frame) OnKindChanged() {
	selIndex := s.kindSelector.CurrentIndex()
	if selIndex < 0 {
		return
	}

	kind := Orgs[selIndex]
	if kind.Name == "server" || kind.Name == "server&client" {
		s.hostButton.SetVisible(true)
	} else {
		s.hostButton.SetVisible(false)
	}

	fileName := s.crtFileName.Text()
	if fileName == "" {
		s.cfgItem.Crt.Name = kind.Name
		s.crtFileName.SetText(kind.Name)
	} else if fileName != kind.Name {
		for _, item := range Orgs {
			if fileName == item.Name {
				s.crtFileName.SetText(kind.Name)
				break
			}
		}
	}
}

func (s *Frame) OnCreateCrt() {
	s.dbCa.Submit()
	s.dbCrt.Submit()

	folder := filepath.Join(s.cfgItem.Crt.RootFolder, s.cfgItem.Crt.SubFolder)
	pfxPath := filepath.Join(folder, fmt.Sprintf("%s.pfx", s.cfgItem.Crt.Name))
	_, err := os.Stat(pfxPath)
	if err == nil || os.IsExist(err) {
		sv := walk.MsgBox(&s.FormBase, "新建证书", fmt.Sprintf("证书\r\n%s\r\n已存在\r\n\r\n是否继续创建并覆盖当前已存在的证书？", pfxPath), walk.MsgBoxYesNo)
		if sv != 6 {
			return
		}
	}

	s.toolBar.SetEnabled(false)
	s.mainComposite.SetEnabled(false)
	s.createButton.SetText("创建中...")

	go func() {
		err := s.createCrt()
		if err != nil {
			walk.MsgBox(&s.FormBase, "新建证书失败", err.Error(), walk.MsgBoxIconError)
		} else {
			s.SaveConfig()
			walk.MsgBox(&s.FormBase, "新建证书成功", pfxPath, walk.MsgBoxIconInformation)
		}

		s.createButton.SetText("创建")
		s.mainComposite.SetEnabled(true)
		s.toolBar.SetEnabled(true)
	}()

}

func (s *Frame) SaveConfig() error {
	err := s.dbCa.Submit()
	if err != nil {
		return err
	}
	err = s.dbCrt.Submit()
	if err != nil {
		return err
	}

	return s.cfg.SaveToFile(s.cfgPath)
}

func (s *Frame) VerifyCa() {
	s.dbCa.Submit()
	crt, _, err := s.verifyCa()
	if err != nil {
		walk.MsgBox(&s.FormBase, "验证失败", err.Error(), walk.MsgBoxIconError)
	} else {
		msg := &strings.Builder{}
		msg.WriteString(fmt.Sprintf("证书类型：%s\r\n", orgDisplayName(crt.Organization())))
		msg.WriteString(fmt.Sprintf("证书标识：%s\r\n", crt.OrganizationalUnit()))
		msg.WriteString(fmt.Sprintf("显示名称：%s\r\n", crt.CommonName()))
		msg.WriteString(fmt.Sprintf("有效期：%s 至 %s", crt.NotBefore().Format("2006-01-02"), crt.NotAfter().Format("2006-01-02")))
		walk.MsgBox(&s.FormBase, "验证成功", msg.String(), walk.MsgBoxIconInformation)
		s.cfg.SaveToFile(s.cfgPath)
	}
}

func (s *Frame) createCrt() error {
	if s.cfgItem.Ca.CrtFile == "" {
		return fmt.Errorf("CA证书文件为空")
	}
	if s.cfgItem.Ca.KeyFile == "" {
		return fmt.Errorf("CA私钥文件为空")
	}
	if s.cfgItem.Crt.Organization == "" {
		return fmt.Errorf("证书类型为空")
	}
	if s.cfgItem.Crt.RootFolder == "" {
		return fmt.Errorf("输出根目录为空")
	}
	if s.cfgItem.Crt.Name == "" {
		return fmt.Errorf("文件名称为空")
	}
	if s.cfgItem.Crt.OrganizationalUnit == "" {
		return fmt.Errorf("证书标识为空")
	}

	caCrt := &gcrt.Crt{}
	err := caCrt.FromFile(s.cfgItem.Ca.CrtFile)
	if err != nil {
		return fmt.Errorf("加载CA证书错误: %v", err)
	}
	caPrivate := &grsa.Private{}
	err = caPrivate.FromFile(s.cfgItem.Ca.KeyFile, fmt.Sprint(s.cfgItem.Ca.Password().Get()))
	if err != nil {
		return fmt.Errorf("加载CA私钥错误: %v", err)
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

	crtTemplate := &gcrt.Template{
		Organization:       s.cfgItem.Crt.Organization,
		OrganizationalUnit: s.cfgItem.Crt.OrganizationalUnit,
		CommonName:         s.cfgItem.Crt.CommonName,
		Locality:           s.cfgItem.Crt.Locality,
		Province:           s.cfgItem.Crt.Province,
		StreetAddress:      s.cfgItem.Crt.StreetAddress,
		Hosts:              s.cfgItem.Crt.Hosts,
		ExpiredDays:        s.cfgItem.Crt.ExpiredDays,
	}
	template, err := crtTemplate.Template()
	if err != nil {
		return err
	}
	crt := &gcrt.Pfx{}
	err = crt.Create(template, caCrt.Certificate(), public, caPrivate)
	if err != nil {
		return err
	}
	folder := filepath.Join(s.cfgItem.Crt.RootFolder, s.cfgItem.Crt.SubFolder)
	pfxPath := filepath.Join(folder, fmt.Sprintf("%s.pfx", s.cfgItem.Crt.Name))
	pfxPassword := fmt.Sprint(s.cfgItem.Crt.Password().Get())
	err = crt.ToFile(pfxPath, caCrt, private, pfxPassword)
	if err != nil {
		return err
	}
	crtPath := filepath.Join(folder, fmt.Sprintf("%s.crt", s.cfgItem.Crt.Name))
	crt.Crt.ToFile(crtPath)

	keyPath := filepath.Join(folder, fmt.Sprintf("%s.key", s.cfgItem.Crt.Name))
	private.ToFile(keyPath, "")

	return nil
}

func (s *Frame) verifyCa() (*gcrt.Crt, *grsa.Private, error) {
	key := &grsa.Private{}
	err := key.FromFile(s.cfgItem.Ca.KeyFile, fmt.Sprint(s.cfgItem.Ca.Password().Get()))
	if err != nil {
		return nil, nil, fmt.Errorf("私钥无效: %v", err)
	}
	crt := &gcrt.Crt{}
	err = crt.FromFile(s.cfgItem.Ca.CrtFile)
	if err != nil {
		return nil, nil, fmt.Errorf("证书无效: %v", err)
	}

	return crt, key, nil
}

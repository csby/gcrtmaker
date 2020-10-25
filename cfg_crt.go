package main

type CfgCrt struct {
	RootFolder         string   `json:"rootFolder" note:"输出根目录"`
	SubFolder          string   `json:"subFolder" note:"输出子目录"`
	Name               string   `json:"name" note:"文件名称"`
	Organization       string   `json:"organization" note:"证书类型"`
	OrganizationalUnit string   `json:"organizationalUnit" note:"证书标识"`
	CommonName         string   `json:"commonName" note:"显示名称"`
	Locality           string   `json:"locality" note:"地区"`
	Province           string   `json:"province" note:"省份"`
	StreetAddress      string   `json:"streetAddress" note:"地址"`
	Hosts              []string `json:"hosts" note:"主机"`
	ExpiredDays        int64    `json:"expiredDays" note:"有效期(默认365)"`
	PfxPassword        string   `json:"pfxPassword" note:"pfx这书密码"`
}

func (s *CfgCrt) Password() *CfgPwd {
	return &CfgPwd{value: &s.PfxPassword}
}

package main

type CfgCa struct {
	CrtFile     string `json:"crtFile" note:"证书文件路径"`
	KeyFile     string `json:"keyFile" note:"私钥文件路径"`
	KeyPassword string `json:"keyPassword" note:"私钥密码"`
}

func (s *CfgCa) Password() *CfgPwd {
	return &CfgPwd{value: &s.KeyPassword}
}

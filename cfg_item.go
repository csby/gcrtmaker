package main

type CfgItem struct {
	Name string     `json:"name"`
	Ca   CfgCa      `json:"ca"`
	Crt  CfgCrt     `json:"crt"`
	Crl  CfgCrl     `json:"crl"`
	Vpn  CfgOpenVpn `json:"vpn"`
}

func NewCfgItem() *CfgItem {
	clientTemplate := &TextBuilder{}
	clientTemplate.AddLine("client")
	clientTemplate.AddLine("dev tun")
	clientTemplate.AddLine("proto udp")
	clientTemplate.AddLine("remote 192.168.123.11 1194")
	clientTemplate.AddLine("resolv-retry infinite")
	clientTemplate.AddLine("nobind")
	clientTemplate.AddLine("persist-key")
	clientTemplate.AddLine("persist-tun")
	clientTemplate.AddLine("remote-cert-tls server")
	clientTemplate.AddLine("cipher AES-256-CBC")
	clientTemplate.AddLine("verb 3")

	serverTemplate := &TextBuilder{}
	serverTemplate.AddLine("port 1194")
	serverTemplate.AddLine("proto udp")
	serverTemplate.AddLine("dev tun")
	serverTemplate.AddLine("server 10.8.0.0 255.255.255.0")
	serverTemplate.AddLine("ifconfig-pool-persist ipp.txt")
	serverTemplate.AddLine("client-to-client")
	serverTemplate.AddLine("keepalive 10 120")
	serverTemplate.AddLine("cipher AES-256-CBC")
	serverTemplate.AddLine("persist-key")
	serverTemplate.AddLine("persist-tun")
	serverTemplate.AddLine("status openvpn-status.log")
	serverTemplate.AddLine("log openvpn.log")
	serverTemplate.AddLine("verb 3")
	serverTemplate.AddLine("explicit-exit-notify 1")

	return &CfgItem{
		Crt: CfgCrt{
			Organization: "client",
			ExpiredDays:  365,
			Hosts:        make([]string, 0),
		},
		Vpn: CfgOpenVpn{
			ServerTemplate: serverTemplate.String(),
			ClientTemplate: clientTemplate.String(),
		},
	}
}

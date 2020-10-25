package main

type CfgOpenVpn struct {
	DiffieHellmanParameters string `json:"dh"` // Generate your own with: `openssl dhparam -out dh2048.pem 2048`
	TlsAuth                 string `json:"ta"` // Generate with: `openvpn --genkey --secret ta.key`
	ServerTemplate          string `json:"serverTemplate"`
	ClientTemplate          string `json:"clientTemplate"`
}

package main

import (
	"encoding/base64"
	"fmt"
	"github.com/csby/gsecurity/gaes"
)

type CfgPwd struct {
	value *string
}

func (*CfgPwd) CanSet() bool {
	return true
}

func (s *CfgPwd) Get() interface{} {
	if s.value == nil {
		return ""
	}
	if len(*s.value) < 1 {
		return ""
	}

	data, err := base64.StdEncoding.DecodeString(*s.value)
	if err != nil {
		return ""
	}
	aesEncoder := &gaes.Aes{
		Key:       "Pwd#Crt@2020",
		Algorithm: "AES-128-CBC",
	}
	decData, err := aesEncoder.Decrypt(data)
	if err != nil {
		return ""
	}

	return string(decData)
}

func (s *CfgPwd) Set(v interface{}) error {
	if s.value == nil {
		return fmt.Errorf("invalid value: nil")
	}
	value := fmt.Sprint(v)
	if len(value) < 1 {
		*s.value = ""
		return nil
	}
	aesEncoder := &gaes.Aes{
		Key:       "Pwd#Crt@2020",
		Algorithm: "AES-128-CBC",
	}
	data, err := aesEncoder.Encrypt([]byte(value))
	if err != nil {
		*s.value = ""
		return nil
	}

	*s.value = base64.StdEncoding.EncodeToString(data)
	return nil
}

func (*CfgPwd) Zero() interface{} {
	return ""
}

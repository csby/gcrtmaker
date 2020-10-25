package main

import (
	"encoding/asn1"
	"fmt"
	"github.com/csby/gsecurity/gcrt"
	"github.com/lxn/walk/declarative"
	"strings"
)

var (
	Orgs = []*Org{
		{
			Name:        "server",
			DisplayName: "服务器",
		},
		{
			Name:        "client",
			DisplayName: "客户端",
		},
		{
			Name:        "server&client",
			DisplayName: "服务器和客户端",
		},
		{
			Name:        "user",
			DisplayName: "用户",
		},
	}

	CRLReasons = []*CRLReason{
		{
			Code:        gcrt.CRLReasonCodeUnspecified,
			DisplayName: "未指定",
		},
		{
			Code:        gcrt.CRLReasonCodeKeyCompromise,
			DisplayName: "密钥泄漏",
		},
		{
			Code:        gcrt.CRLReasonCodeCACompromise,
			DisplayName: "CA泄漏",
		},
		{
			Code:        gcrt.CRLReasonCodeAffiliationChanged,
			DisplayName: "附属关系已更改",
		},
		{
			Code:        gcrt.CRLReasonCodeSuperseded,
			DisplayName: "被取代",
		},
		{
			Code:        gcrt.CRLReasonCodeCessationOfOperation,
			DisplayName: "停止操作",
		},
		{
			Code:        gcrt.CRLReasonCodeCertificateHold,
			DisplayName: "证书挂起",
		},
	}
)

type Org struct {
	Name        string
	DisplayName string
}

func orgDisplayName(name string) string {
	for _, item := range Orgs {
		if strings.ToLower(name) == item.Name {
			return item.DisplayName
		}
	}
	return name
}

type CRLReason struct {
	Code        asn1.Enumerated
	DisplayName string
}

func crlReasonDisplayName(code asn1.Enumerated) string {
	for _, item := range CRLReasons {
		if code == item.Code {
			return item.DisplayName
		}
	}
	return fmt.Sprint(code)
}

func crlReasonIndex(code asn1.Enumerated) int {
	count := len(CRLReasons)
	for index := 0; index < count; index++ {
		item := CRLReasons[index]
		if code == item.Code {
			return index
		}
	}
	return -1
}

func newLabel(text string, row, column int) declarative.Composite {
	return declarative.Composite{
		Row:    row,
		Column: column,
		Layout: declarative.VBox{
			Margins: declarative.Margins{
				Top: 1,
			},
		},
		Children: []declarative.Widget{
			declarative.TextLabel{
				Text:          text,
				TextAlignment: declarative.AlignHFarVCenter,
			},
		},
	}
}

type TextBuilder struct {
	strings.Builder
}

func (s *TextBuilder) AddLine(line string) {
	s.WriteString(line)
	s.WriteString("\r\n")
}

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Cfg struct {
	DefaultItemName string     `json:"defaultItem"`
	Items           []*CfgItem `json:"items"`
}

func NewCfg() *Cfg {
	return &Cfg{
		Items: make([]*CfgItem, 0),
	}
}

func (s *Cfg) GetItem(name string) *CfgItem {
	count := len(s.Items)
	for i := 0; i < count; i++ {
		item := s.Items[i]
		if item == nil {
			continue
		}

		if strings.Compare(item.Name, name) == 0 {
			return item
		}
	}

	return nil
}

func (s *Cfg) IndexOf(cfgItem *CfgItem) int {
	count := len(s.Items)
	for i := 0; i < count; i++ {
		item := s.Items[i]
		if item == nil {
			continue
		}

		if cfgItem == item {
			return i
		}
	}

	return -1
}

func (s *Cfg) SaveToFile(filePath string) error {
	bytes, err := json.MarshalIndent(s, "", "    ")
	if err != nil {
		return err
	}

	fileFolder := filepath.Dir(filePath)
	_, err = os.Stat(fileFolder)
	if os.IsNotExist(err) {
		os.MkdirAll(fileFolder, 0777)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = fmt.Fprint(file, string(bytes[:]))

	return err
}

func (s *Cfg) LoadFromFile(filePath string) error {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, s)
}

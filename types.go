package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

const FSBase = "/home/vahid/Desktop/temp1/tcpdumps"

type DeploymentInfo struct {
	DNS      map[string]string              `json:"dns"`
	Networks map[string]*TCPDUMPNetworkInfo `json:"networks"`
}

type TCPDUMPNetworkInfo struct {
	ID      string `json:"id"`
	ShortID string `json:"short_id"`
	Name    string `json:"name"`
	FSName  string `json:"fs_name"`
}

func (d *DeploymentInfo) Save() {
	b, err := json.MarshalIndent(d, "", "\t")
	if err != nil {
		panic(err)
	}
	direcotry_name := d.Networks["overlay"].ShortID
	if _, err := os.Stat(FSBase + "/" + direcotry_name); os.IsNotExist(err) {
		os.Mkdir(FSBase+"/"+direcotry_name, os.ModeDir|0777)
	}
	ioutil.WriteFile(FSBase+"/"+direcotry_name+"/info.json", b, 0666)
}

func LoadDeploymentInfo(name string) *DeploymentInfo {
	d := DeploymentInfo{}
	b, err := ioutil.ReadFile(FSBase + "/" + name + "/info.json")
	if err != nil {
		panic(err)
	}
	json.Unmarshal(b, &d)
	return &d
}

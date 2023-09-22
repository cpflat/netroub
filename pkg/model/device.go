package model

type Connections struct {
	SrcNode      string `json:"src_node"`
	SrcInterface string `json:"src_interface"`
	DstNode      string `json:"dst_node"`
	DstInterface string `json:"dst_interface"`
}

type ParamsInterface struct {
	Name     string `json:"name"`
	Priority string `json:"priority"`
}

type Interfaces struct {
	Name   string          `json:"name"`
	Params ParamsInterface `json:"params"`
}

type Params struct {
	As   string `json:"as"`
	Name string `json:"name"`
}

type Nodes struct {
	Name       string       `json:"name"`
	Params     Params       `json:"params"`
	Interfaces []Interfaces `json:"interfaces"`
}

type Data struct {
	Name        string        `json:"name"`
	Nodes       []Nodes       `json:"nodes"`
	Connections []Connections `json:"connections"`
}

var Devices Data

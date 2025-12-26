package model

import "fmt"

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

// LabName is the custom lab name for containerlab.
// If empty, Devices.Name (topology name) is used.
var LabName string

// GetLabName returns the lab name to use for containerlab.
// Returns custom LabName if set, otherwise returns Devices.Name.
func GetLabName() string {
	if LabName != "" {
		return LabName
	}
	return Devices.Name
}

// SetLabName sets a custom lab name for containerlab.
func SetLabName(name string) {
	LabName = name
}

// ResetLabName clears the custom lab name.
func ResetLabName() {
	LabName = ""
}

func ValidateHostNames(hosts []string) error {
	for _, host := range hosts {
		ok := false
		for _, device := range Devices.Nodes {
			if host == device.Name {
				ok = true
			}
		}
		if !ok {
			return fmt.Errorf("host %s not found in the topology", host)
		}
	}
	return nil
}

func ClabHostName(host string) string {
	return "clab-" + GetLabName() + "-" + host
}

func GetDeviceIndex(device string) int {
	for i, node := range Devices.Nodes {
		if device == node.Name {
			return i
		}
	}
	return -1
}

package main

import (
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/cloudfoundry/gosigar"
	"github.com/shunichi/go-cpuid"
)

type HddInfo struct {
	Name  string
	Total uint64
	Used  uint64
}

type HostInfo struct {
	Name     string
	Cpu      string
	Memory   uint64
	HddInfos []HddInfo
}

func makeHddInfo(fs sigar.FileSystem) HddInfo {
	usage := sigar.FileSystemUsage{}
	usage.Get(fs.DirName)
	return HddInfo{fs.DevName, usage.Total * 1024, usage.Used * 1024}
}

func makeSystemHddInfos() []HddInfo {
	fslist := sigar.FileSystemList{}
	fslist.Get()
	hddInfos := make([]HddInfo, 0, len(fslist.List))
	for _, fs := range fslist.List {
		if strings.HasPrefix(fs.DevName, "/") {
			hddInfos = append(hddInfos, makeHddInfo(fs))
		}
	}
	return hddInfos
}

func makeHostInfo() HostInfo {
	mem := sigar.Mem{}
	mem.Get()
	return HostInfo{hostname(), cpuid.BrandName(), mem.Total, makeSystemHddInfos()}
}

func makeHostInfoAsJson(indent bool) (string, error) {
	info := makeHostInfo()
	var err error
	var b []byte
	if indent {
		b, err = json.MarshalIndent(info, "", "  ")
	} else {
		b, err = json.Marshal(info)
	}
	if err == nil {
		return string(b), nil
	} else {
		return "", errors.New("json.Marshal failed")
	}
}

func hostname() string {
	if name, err := os.Hostname(); err == nil {
		return name
	} else {
		return "unknown host"
	}
}

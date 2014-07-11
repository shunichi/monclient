package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloudfoundry/gosigar"
	"github.com/shunichi/go-cpuid"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

var verbose bool = false

func main() {

	if hostname, err := os.Hostname(); err == nil {
		fmt.Printf("hostname: %s\n", hostname)
	} else {
		fmt.Println("failed to get hostname.")
	}

	mem := sigar.Mem{}
	mem.Get()

	fmt.Printf("Mem used: %.1f/%.1f GiB\n", gb(mem.Used), gb(mem.Total))

	postHostInfoByJson()
}

func hostname() string {
	if name, err := os.Hostname(); err == nil {
		return name
	} else {
		return "unknown host"
	}
}

// func postHostInfo() {
// 	mem := sigar.Mem{}
// 	mem.Get()

// 	v := url.Values{}
// 	if hostname, err := os.Hostname(); err == nil {
// 		v.Set("name", hostname)
// 	} else {
// 		return
// 	}
// 	v.Set("memory", strconv.FormatUint(mem.Total, 10))
// 	v.Set("cpu", cpuid.BrandName())
// 	if resp, err := http.PostForm("http://localhost:4567/update", v); err == nil {
// 		fmt.Printf("status = %s\n", resp.Status)
// 	} else {
// 		fmt.Printf("error = %s\n", err.Error())
// 	}
// }

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

func makeHostInfoAsJson() (string, error) {
	info := makeHostInfo()
	if b, err := json.Marshal(info); err == nil {
		return string(b), nil
	} else {
		return nil, errors.New("json.Marshal failed")
	}
}

func postHostInfoByJson() {
	if json, err = makeHostInfoAsJson(); err != nil {
		if resp, err := http.Post("http://localhost:4567/json", "application/json", bytes.NewReader(b)); err == nil {
			fmt.Printf("status = %s\n", resp.Status)
		}
	}
	// info := makeHostInfo()
	// if b, err := json.Marshal(info); err == nil {
	// 	fmt.Println(string(b))
	// 	if resp, err := http.Post("http://localhost:4567/json", "application/json", bytes.NewReader(b)); err == nil {
	// 		fmt.Printf("status = %s\n", resp.Status)
	// 	}
	// }
}

func gb(bytes uint64) float64 {
	return float64(bytes) / (1024 * 1024 * 1024)
}

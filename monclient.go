package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	//	"net/url"
	//	"strconv"
	"github.com/cloudfoundry/gosigar"
	"github.com/jessevdk/go-flags"
	"github.com/shunichi/go-cpuid"
)

var opts struct {
	Verbose bool `short:"v" long:"verbose" description:"Show verbose information"`
}

var verbose bool = false

func main() {
	if args, err := flags.Parse(&opts); err != nil {
		fmt.Errorf("%s\n", err.Error())
		os.Exit(1)
	} else {
		fmt.Printf("%+v\n", opts)
		for a := range args {
			fmt.Println(a)
		}
	}
	if opts.Verbose {
		printHostInfoAsJson()
	}
	postHostInfoAsJson()
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

func printHostInfoAsJson() {
	if json, err := makeHostInfoAsJson(true); err == nil {
		fmt.Print(json + "\n")
	}
}

func postHostInfoAsJson() {
	if json, err := makeHostInfoAsJson(false); err == nil {
		if resp, err := http.Post("http://localhost:4567/json", "application/json", strings.NewReader(json)); err == nil {
			if opts.Verbose {
				printResponse(resp)
			}
		} else {
			fmt.Errorf("%s", err.Error())
		}
	}
}

func printResponse(resp *http.Response) {
	fmt.Printf("status = %s\n", resp.Status)

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	s := buf.String()
	fmt.Printf("body = %s\n", s)
}

func gb(bytes uint64) float64 {
	return float64(bytes) / (1024 * 1024 * 1024)
}

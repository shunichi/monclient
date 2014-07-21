package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Verbose bool   `short:"v" long:"verbose" description:"Show verbose information"`
	Url     string `short:"u" long:"url" description:"Post URL" default:"http://localhost:4567/json"`
}

func usage() {
	const msg = `
Usage: monclient [options]

Options:
    -v, --verbose      Show verbose infomation
    -u, --url          Post URL
`

	os.Stderr.Write([]byte(msg))
}

func main() {
	if _, err := flags.Parse(&opts); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		usage()
		os.Exit(1)
	}

	if opts.Verbose {
		printHostInfoAsJson()
	}

	if err := postHostInfoAsJson(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func postHostInfoAsJson() error {
	json, err := makeHostInfoAsJson(false)
	if err != nil {
		return err
	}

	resp, err := http.Post(opts.Url, "application/json", strings.NewReader(json))
	if err != nil {
		return err
	}

	if opts.Verbose {
		printResponse(resp)
	}
	return nil
}

func printResponse(resp *http.Response) {
	fmt.Printf("status: %s\n", resp.Status)

	fmt.Printf("header:\n")
	for k, v := range resp.Header {
		fmt.Printf("  %s: %v\n", k, v)
	}

	s, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		return
	}
	fmt.Printf("body:\n%s\n", s)
}

func printHostInfoAsJson() {
	if json, err := makeHostInfoAsJson(true); err == nil {
		fmt.Printf("host info:\n%s\n\n", json)
	}
}

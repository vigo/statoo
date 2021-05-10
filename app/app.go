/*
Package app is the core library of statoo command-line app

Usage

	cmd := NewCLIApplication()
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
*/
package app

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const version = "0.2.2"
const defTimeout = 10

var (
	argURL                string
	optVersionInformation *bool
	optTimeout            *int
	optVerboseOutput      *bool
	optJSONOutput         *bool
	optHeaders            headersFlag
	optFind               *string

	usage = `
usage: %[1]s [-flags] URL

  flags:

  -version        display version information (%s)
  -t, -timeout    default timeout in seconds  (default: %d)
  -h, -help       display help
  -json           provides json output
  -verbose        verbose output              (default: false)
  -header         request header, multiple allowed
  -find           find text in repsonse body if -json is set
	
  examples:
  
  $ %[1]s "https://ugur.ozyilmazel.com"
  $ %[1]s -timeout 30 "https://ugur.ozyilmazel.com"
  $ %[1]s -verbose "https://ugur.ozyilmazel.com"
  $ %[1]s -json https://vigo.io
  $ %[1]s -json -find "python" https://vigo.io
  $ %[1]s -header "Authorization: Bearer TOKEN" https://vigo.io
  $ %[1]s -header "Authorization: Bearer TOKEN" -header "X-Api-Key: APIKEY" https://vigo.io
  $ %[1]s -json -find "Meetup organization" https://vigo.io

`
)

type headersFlag []string

func (h *headersFlag) String() string {
	return "headers"
}

func (h *headersFlag) Set(value string) error {
	*h = append(*h, strings.TrimSpace(value))
	return nil
}

// CLIApplication represents app structure
type CLIApplication struct {
	Out io.Writer
}

// JSONResponse represents data structure of json repsonse
type JSONResponse struct {
	URL       string    `json:"url"`
	Status    int       `json:"status"`
	CheckedAt time.Time `json:"checked_at"`
	Find      *string   `json:"find,omitempty"`
	Found     *bool     `json:"found,omitempty"`
}

// NewCLIApplication creates new CLIApplication instance
func NewCLIApplication() *CLIApplication {
	flag.Usage = func() {
		// w/o os.Stdout, you need to pipe out via
		// cmd &> /path/to/file
		fmt.Fprintf(os.Stdout, usage, os.Args[0], version, defTimeout)
		os.Exit(0)
	}

	optVersionInformation = flag.Bool("version", false, "")
	optVerboseOutput = flag.Bool("verbose", false, "")
	optJSONOutput = flag.Bool("json", false, "")
	optTimeout = flag.Int("timeout", defTimeout, "")
	optFind = flag.String("find", "", "")

	flag.IntVar(optTimeout, "t", defTimeout, "")
	flag.Var(&optHeaders, "header", "")

	flag.Parse()

	argURL = flag.Arg(0)

	return &CLIApplication{
		Out: os.Stdout,
	}
}

// Run executes main application
func (c *CLIApplication) Run() error {
	if *optVersionInformation {
		fmt.Fprintln(c.Out, version)
		return nil
	}

	if err := c.Validate(); err != nil {
		return err
	}

	return nil
}

// Validate runs validations for flags
func (c *CLIApplication) Validate() error {
	_, err := url.ParseRequestURI(argURL)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	if *optTimeout > 100 || *optTimeout < 1 {
		return fmt.Errorf("invalid timeout value: %d", *optTimeout)
	}
	return c.GetResult()
}

// GetResult fetches the status information of given URL
func (c *CLIApplication) GetResult() error {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
	}

	timeout := time.Duration(*optTimeout) * time.Second
	client := &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "GET", argURL, nil)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}

	if len(optHeaders) > 0 {
		for _, headerValue := range optHeaders {
			vals := strings.Split(headerValue, ":")
			req.Header.Set(vals[0], vals[1])
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if *optJSONOutput {
		js := &JSONResponse{
			URL:       argURL,
			Status:    resp.StatusCode,
			CheckedAt: time.Now().UTC(),
			Find:      nil,
			Found:     nil,
		}

		if *optFind != "" {
			body, err := io.ReadAll(resp.Body)
			if err == nil {
				var boolFound bool
				boolFound = strings.Contains(string(body), *optFind)
				js.Find = optFind
				js.Found = &boolFound
			}
		}

		j, err := json.Marshal(js)
		if err != nil {
			return fmt.Errorf("error: %v", err)
		}
		c.Out.Write(j)
		return nil
	}
	if *optVerboseOutput {
		fmt.Fprintf(c.Out, "%s -> ", argURL)
	}
	fmt.Fprintf(c.Out, "%d\n", resp.StatusCode)
	return nil
}

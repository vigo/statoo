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
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/vigo/statoo/app/version"
)

const defTimeout = 10

var _ flag.Value = (*HeadersFlag)(nil)
var _ flag.Value = (*BasicAuthFlag)(nil)

// HeadersFlag ...
type HeadersFlag []string

func (f *HeadersFlag) String() string {
	return fmt.Sprintf("%s", *f)
}

// Set ...
func (f *HeadersFlag) Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("header should not be empty")
	}
	if strings.Count(value, ":") != 1 {
		return fmt.Errorf("invalind header data: %s", value)
	}
	if len(strings.FieldsFunc(value, func(c rune) bool { return c == ':' })) != 2 {
		return fmt.Errorf("invalind header data: %s", value)
	}
	*f = append(*f, value)
	return nil
}

// BasicAuthFlag ...
type BasicAuthFlag string

// Set ...
func (f *BasicAuthFlag) Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("auth flag should not be empty")
	}
	if strings.Count(value, ":") != 1 {
		return fmt.Errorf("invalind auth data: %s", value)
	}
	if len(strings.FieldsFunc(value, func(c rune) bool { return c == ':' })) != 2 {
		return fmt.Errorf("invalind auth data: %s", value)
	}

	*f = BasicAuthFlag(value)
	return nil
}

func (f *BasicAuthFlag) String() string {
	return string(*f)
}

var (
	// ArgURL ...
	ArgURL string

	// OptVersionInformation ...
	OptVersionInformation *bool

	// OptTimeout ...
	OptTimeout *int

	// OptVerboseOutput ...
	OptVerboseOutput *bool

	// OptJSONOutput ...
	OptJSONOutput *bool

	// OptHeaders ...
	OptHeaders HeadersFlag
	// OptFind ...
	OptFind *string

	// OptBasicAuth ...
	OptBasicAuth BasicAuthFlag

	usage = `
usage: %[1]s [-flags] URL

  flags:

  -version        display version information (%s)
  -verbose        verbose output              (default: false)
  -header         request header, multiple allowed, "Key: Value"
  -t, -timeout    default timeout in seconds  (default: %d)
  -h, -help       display help
  -j, -json       provides json output
  -f, -find       find text in response body if -json is set
  -a, -auth       basic auth "username:password"

  examples:
  
  $ %[1]s "https://ugur.ozyilmazel.com"
  $ %[1]s -timeout 30 "https://ugur.ozyilmazel.com"
  $ %[1]s -verbose "https://ugur.ozyilmazel.com"
  $ %[1]s -json https://vigo.io
  $ %[1]s -json -find "python" https://vigo.io
  $ %[1]s -header "Authorization: Bearer TOKEN" https://vigo.io
  $ %[1]s -header "Authorization: Bearer TOKEN" -header "X-Api-Key: APIKEY" https://vigo.io
  $ %[1]s -json -find "Meetup organization" https://vigo.io
  $ %[1]s -auth "user:secret" https://vigo.io

`
	bashCompletion = `__statoo_comp()
{
    local cur next
    cur="${COMP_WORDS[COMP_CWORD]}"
    opts="-a -auth -f -find -header -h -help -j -json -t -timeout -verbose -version"
    COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
}
complete -F __statoo_comp statoo`
)

// CLIApplication represents app structure
type CLIApplication struct {
	Out io.Writer
}

// JSONResponse represents data structure of json response
type JSONResponse struct {
	URL       string    `json:"url"`
	Status    int       `json:"status"`
	CheckedAt time.Time `json:"checked_at"`
	Elapsed   float64   `json:"elapsed,omitempty"`
	Length    int       `json:"length,omitempty"`
	Find      *string   `json:"find,omitempty"`
	Found     *bool     `json:"found,omitempty"`
}

func trimSpaces(s []string) {
	for i, v := range s {
		s[i] = strings.TrimSpace(v)
	}
}

// NewCLIApplication creates new CLIApplication instance
func NewCLIApplication() *CLIApplication {
	flag.Usage = func() {
		// w/o os.Stdout, you need to pipe out via
		// cmd &> /path/to/file
		fmt.Fprintf(os.Stdout, usage, os.Args[0], version.Version, defTimeout)
		os.Exit(0)
	}

	OptVersionInformation = flag.Bool("version", false, fmt.Sprintf("display version information (%s)", version.Version))
	OptVerboseOutput = flag.Bool("verbose", false, "verbose output")

	helpJSON := "provides json output"
	OptJSONOutput = flag.Bool("json", false, helpJSON)
	flag.BoolVar(OptJSONOutput, "j", false, helpJSON+" (short)")

	helpTimeout := "default timeout in seconds"
	OptTimeout = flag.Int("timeout", defTimeout, helpTimeout)
	flag.IntVar(OptTimeout, "t", defTimeout, helpTimeout+" (short)")

	helpFind := "find text in response body if -json is set"
	OptFind = flag.String("find", "", helpFind)
	flag.StringVar(OptFind, "f", "", helpFind+" (short)")

	flag.Var(&OptHeaders, "header", "")

	helpBasicAuth := "basic auth \"username:password\""
	flag.Var(&OptBasicAuth, "auth", helpBasicAuth)
	flag.Var(&OptBasicAuth, "a", helpBasicAuth+" (short)")

	flag.Parse()

	ArgURL = flag.Arg(0)

	return &CLIApplication{
		Out: os.Stdout,
	}
}

// Run executes main application
func (c *CLIApplication) Run() error {
	if *OptVersionInformation {
		fmt.Fprintln(c.Out, version.Version)
		return nil
	}

	if ArgURL == "bash-completion" {
		fmt.Fprintln(c.Out, bashCompletion)
		return nil
	}

	if err := c.Validate(); err != nil {
		return err
	}

	return nil
}

// Validate runs validations for flags
func (c *CLIApplication) Validate() error {
	_, err := url.ParseRequestURI(ArgURL)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	if *OptTimeout > 100 || *OptTimeout < 1 {
		return fmt.Errorf("invalid timeout value: %d", *OptTimeout)
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

	timeout := time.Duration(*OptTimeout) * time.Second
	client := &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", ArgURL, nil)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}

	req.Header.Set("Accept-Encoding", "gzip")

	if len(OptHeaders) > 0 {
		for _, headerValue := range OptHeaders {
			vals := strings.Split(headerValue, ":")
			trimSpaces(vals)
			req.Header.Set(vals[0], vals[1])
		}
	}

	if OptBasicAuth.String() != "" {
		words := strings.Split(OptBasicAuth.String(), ":")
		trimSpaces(words) // remove spaces, foo      :  bar => foo:bar
		req.SetBasicAuth(words[0], words[1])
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}
	elapsed := time.Since(start)

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if *OptJSONOutput {
		js := &JSONResponse{
			URL:       ArgURL,
			Status:    resp.StatusCode,
			CheckedAt: time.Now().UTC(),
			Elapsed:   float64(elapsed) / float64(time.Millisecond),
			Find:      nil,
			Found:     nil,
		}

		if *OptFind != "" {
			var bodyReader io.ReadCloser

			switch resp.Header.Get("Content-Encoding") {
			case "gzip":
				bodyReader, err = gzip.NewReader(resp.Body)
				if err != nil {
					return fmt.Errorf("body read (gzip) error: %v", err)
				}
				defer bodyReader.Close()
			default:
				bodyReader = resp.Body
			}

			body, err := io.ReadAll(bodyReader)
			if err == nil {
				boolFound := strings.Contains(string(body), *OptFind)
				js.Find = OptFind
				js.Found = &boolFound
			}
			js.Length = len(body)
		}

		j, err := json.Marshal(js)
		if err != nil {
			return fmt.Errorf("error: %v", err)
		}

		_, err = c.Out.Write(j)
		if err != nil {
			return fmt.Errorf("error: %v", err)
		}
		return nil
	}
	if *OptVerboseOutput {
		fmt.Fprintf(c.Out, "%s -> ", ArgURL)
	}
	fmt.Fprintf(c.Out, "%d\n", resp.StatusCode)
	return nil
}

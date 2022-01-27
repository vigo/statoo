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

const (
	defTimeout    = 10
	defTimeoutMin = 1
	defTimeoutMax = 100
)

var (
	errEmptyHeader    = errors.New("header should not be empty")
	errInvalidHeader  = errors.New("invalid header value")
	errInvalidTimeout = errors.New("invalid timeout")
)

// HeadersFlag holds header information for http request.
type HeadersFlag []string

func (f *HeadersFlag) String() string {
	return fmt.Sprintf("%s", *f)
}

// Set appends valid header values to HeadersFlag.
func (f *HeadersFlag) Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errEmptyHeader
	}
	if strings.Count(value, ":") != 1 {
		return fmt.Errorf("%w: %s", errInvalidHeader, value)
	}
	if len(strings.FieldsFunc(value, func(c rune) bool { return c == ':' })) != 2 {
		return fmt.Errorf("%w: %s", errInvalidHeader, value)
	}
	*f = append(*f, value)
	return nil
}

var (
	// ArgURL holds URL input from command-line.
	ArgURL string

	// OptVersionInformation holds boolean for displaying version information.
	OptVersionInformation *bool

	// OptTimeout holds default timeout for network transport operations.
	OptTimeout *int

	// OptVerboseOutput holds boolean for displaying detailed output.
	OptVerboseOutput *bool

	// OptJSONOutput holds boolean for json response instead of text.
	OptJSONOutput *bool

	// OptHeaders holds custom request header key:value.
	OptHeaders HeadersFlag
	// OptFind holds lookup string in the body of the response.
	OptFind *string

	// OptBasicAuth holds basic auth key:value credentials.
	OptBasicAuth *string

	// OptInsecureSkipVerify holds certificate check option.
	OptInsecureSkipVerify *bool

	usage = `
usage: %[1]s [-flags] URL

  flags:

  -version        display version information (%s)
  -verbose        verbose output (default: false)
  -header         request header, multiple allowed, "Key: Value"
  -t, -timeout    default timeout in seconds (default: %d, min: %d, max: %d)
  -h, -help       display help
  -j, -json       provides json output
  -f, -find       find text in response body if -json is set
  -a, -auth       basic auth "username:password"
  -s, -skip       skip certificate check and hostname in that certificate (default: false)

  examples:
  
  $ %[1]s "https://ugur.ozyilmazel.com"
  $ %[1]s -timeout 30 "https://ugur.ozyilmazel.com"
  $ %[1]s -verbose "https://ugur.ozyilmazel.com"
  $ %[1]s -json https://vigo.io
  $ %[1]s -json -find "python" https://vigo.io
  $ %[1]s -header "Authorization: Bearer TOKEN" https://vigo.io
  $ %[1]s -header "Authorization: Bearer TOKEN" -header "X-Api-Key: APIKEY" https://vigo.io
  $ %[1]s -json -find "Golang" https://vigo.io
  $ %[1]s -auth "user:secret" https://vigo.io

`
	bashCompletion = `__statoo_comp()
{
    local cur next
    cur="${COMP_WORDS[COMP_CWORD]}"
    opts="-a -auth -f -find -header -h -help -j -json -t -timeout -s -skip -verbose -version"
    COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
}
complete -F __statoo_comp statoo`
)

// CLIApplication represents app structure.
type CLIApplication struct {
	Out io.Writer
}

// JSONResponse represents data structure of json response.
type JSONResponse struct {
	URL                  string    `json:"url"`
	Status               int       `json:"status"`
	CheckedAt            time.Time `json:"checked_at"`
	Elapsed              float64   `json:"elapsed,omitempty"`
	Length               int       `json:"length,omitempty"`
	Find                 *string   `json:"find,omitempty"`
	Found                *bool     `json:"found,omitempty"`
	SkipCertificateCheck *bool     `json:"skipcc,omitempty"`
}

func trimSpaces(s []string) {
	for i, v := range s {
		s[i] = strings.TrimSpace(v)
	}
}

func flagUsage(code int) func() {
	return func() {
		fmt.Fprintf(
			os.Stdout,
			usage,
			os.Args[0],
			version.Version,
			defTimeout,
			defTimeoutMin,
			defTimeoutMax,
		)
		if code > 0 {
			os.Exit(code)
		}
	}
}

// NewCLIApplication creates new CLIApplication instance.
func NewCLIApplication() *CLIApplication {
	flag.Usage = flagUsage(0)

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
	OptBasicAuth = flag.String("auth", "", helpBasicAuth)
	flag.StringVar(OptBasicAuth, "a", "", helpBasicAuth+" (short)")

	helpOptInsecureSkipVerify := "skip certificate check and hostname in that certificate"
	OptInsecureSkipVerify = flag.Bool("skip", false, helpOptInsecureSkipVerify)
	flag.BoolVar(OptInsecureSkipVerify, "s", false, helpOptInsecureSkipVerify+" (short)")

	flag.Parse()

	ArgURL = flag.Arg(0)

	return &CLIApplication{
		Out: os.Stdout,
	}
}

// Run executes main application.
func (c *CLIApplication) Run() error {
	if *OptVersionInformation {
		fmt.Fprintln(c.Out, version.Version)
		return nil
	}

	if ArgURL == "bash-completion" {
		fmt.Fprintln(c.Out, bashCompletion)
		return nil
	}
	return c.Validate()
}

// Validate runs validations for flags.
func (c *CLIApplication) Validate() error {
	if len(ArgURL) == 0 {
		flagUsage(-1)()
		return nil
	}

	_, err := url.ParseRequestURI(ArgURL)
	if err != nil {
		return fmt.Errorf("url parse error: %w", err)
	}

	if *OptTimeout > 100 || *OptTimeout < 1 {
		return fmt.Errorf("%w: %d", errInvalidTimeout, *OptTimeout)
	}
	return c.GetResult()
}

// GetResult fetches the status information of given URL.
func (c *CLIApplication) GetResult() error {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.MaxIdleConns = 10
	tr.IdleConnTimeout = 30 * time.Second
	tr.DisableCompression = true
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: *OptInsecureSkipVerify} //nolint

	timeout := time.Duration(*OptTimeout) * time.Second
	client := &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "GET", ArgURL, nil)
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}

	req.Header.Set("Accept-Encoding", "gzip")

	if len(OptHeaders) > 0 {
		for _, headerValue := range OptHeaders {
			vals := strings.Split(headerValue, ":")
			trimSpaces(vals)
			req.Header.Set(vals[0], vals[1])
		}
	}

	if *OptBasicAuth != "" {
		words := strings.Split(*OptBasicAuth, ":")
		trimSpaces(words)
		fmt.Println("words", words)
		req.SetBasicAuth(words[0], words[1])
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("response error: %w", err)
	}
	elapsed := time.Since(start)

	if resp.Body != nil {
		defer func() {
			if errClose := resp.Body.Close(); err != nil {
				fmt.Fprintln(os.Stderr, errClose.Error())
			}
		}()
	}

	if *OptJSONOutput {
		js := &JSONResponse{
			URL:                  ArgURL,
			Status:               resp.StatusCode,
			CheckedAt:            time.Now().UTC(),
			Elapsed:              float64(elapsed) / float64(time.Millisecond),
			Find:                 nil,
			Found:                nil,
			SkipCertificateCheck: OptInsecureSkipVerify,
		}

		if *OptFind != "" {
			var bodyReader io.ReadCloser

			switch resp.Header.Get("Content-Encoding") {
			case "gzip":
				bodyReader, err = gzip.NewReader(resp.Body)
				if err != nil {
					return fmt.Errorf("body read (gzip) error: %w", err)
				}
				defer func() {
					if err := bodyReader.Close(); err != nil {
						fmt.Fprintln(os.Stderr, err.Error())
					}
				}()
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
			return fmt.Errorf("json marshal error: %w", err)
		}

		_, err = c.Out.Write(j)
		if err != nil {
			return fmt.Errorf("write error: %w", err)
		}
		return nil
	}
	if *OptVerboseOutput {
		fmt.Fprintf(c.Out, "%s -> ", ArgURL)
	}
	fmt.Fprintf(c.Out, "%d\n", resp.StatusCode)
	return nil
}

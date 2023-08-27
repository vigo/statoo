/*
Package app is the core library of statoo command-line app

Usage

	cmd := NewCLIApplication()
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
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

	"github.com/vigo/statoo/v2/app/flags"
	"github.com/vigo/statoo/v2/app/version"
)

const (
	defTimeout    = 10
	defTimeoutMin = 1
	defTimeoutMax = 100
)

var errInvalidTimeout = errors.New("invalid timeout value")

// variable declarations.
var (
	ArgURL                   string
	OptVersionInformation    *bool
	OptCommitHashInformation *bool
	OptTimeout               *int
	OptVerboseOutput         *bool
	OptJSONOutput            *bool
	OptRequestHeaders        flags.RequestHeadersFlag
	OptResponseHeaders       flags.ResponseHeadersFlag
	OptFind                  *string
	OptBasicAuth             *string
	OptInsecureSkipVerify    *bool
)

// CLIApplication represents app structure.
type CLIApplication struct {
	Out io.Writer
}

// JSONResponse represents data structure of json response.
type JSONResponse struct {
	URL                  string           `json:"url"`
	Status               int              `json:"status"`
	CheckedAt            time.Time        `json:"checked_at"`
	Elapsed              float64          `json:"elapsed,omitempty"`
	Length               int              `json:"length,omitempty"`
	Find                 *string          `json:"find,omitempty"`
	Found                *bool            `json:"found,omitempty"`
	SkipCertificateCheck *bool            `json:"skipcc,omitempty"`
	ResponseHeaders      *map[string]bool `json:"response_headers,omitempty"`
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
			version.CommitHash,
		)
		if code > 0 {
			os.Exit(code)
		}
	}
}

// NewCLIApplication creates new CLIApplication instance.
func NewCLIApplication() *CLIApplication {
	flag.Usage = flagUsage(0)

	OptVersionInformation = flag.Bool(
		"version",
		false,
		fmt.Sprintf("display version information (%s)", version.Version),
	)

	OptCommitHashInformation = flag.Bool(
		"commithash",
		false,
		fmt.Sprintf("display build information (%s)", version.CommitHash),
	)

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

	helpRequestHeaders := "add http headers to your request. can be multiple"
	flag.Var(&OptRequestHeaders, "request-header", helpRequestHeaders)

	helpResponseHeaders := "query response headers, \"Server:GitHub.com\". can be multiple"
	flag.Var(&OptResponseHeaders, "response-header", helpResponseHeaders)

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

	if *OptCommitHashInformation {
		fmt.Fprintln(c.Out, version.CommitHash)
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ArgURL, nil)
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}

	req.Header.Set("Accept-Encoding", "gzip")

	if len(OptRequestHeaders) > 0 {
		for _, headerValue := range OptRequestHeaders {
			vals := strings.Split(headerValue, ":")
			trimSpaces(vals)
			req.Header.Set(vals[0], vals[1])
		}
	}

	if *OptBasicAuth != "" {
		words := strings.Split(*OptBasicAuth, ":")
		trimSpaces(words)
		req.SetBasicAuth(words[0], words[1])
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("response error: %w", err)
	}
	elapsed := time.Since(start)

	// The http Client and Transport guarantee that Body is always
	// non-nil, even on responses without a body or responses with
	// a zero-length body.
	defer func() {
		_ = resp.Body.Close()
	}()

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

		if len(OptResponseHeaders) > 0 {
			foundResponseHeaders := make(map[string]bool)
			for _, headerValue := range OptResponseHeaders {
				vals := strings.Split(headerValue, ":")
				trimSpaces(vals)
				mapKey := vals[0] + "=" + vals[1]

				hvals, ok := resp.Header[vals[0]]
				if ok {
					hval := hvals[0]
					if hval == vals[1] {
						foundResponseHeaders[mapKey] = true
					}
				} else {
					foundResponseHeaders[mapKey] = false
				}
			}
			js.ResponseHeaders = &foundResponseHeaders
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
						fmt.Fprintln(os.Stderr, "gzip body reader close error: %w", err)
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

	_, _ = io.Copy(io.Discard, resp.Body)

	if *OptVerboseOutput {
		fmt.Fprintf(c.Out, "%s -> ", ArgURL)
	}
	fmt.Fprintf(c.Out, "%d\n", resp.StatusCode)
	return nil
}

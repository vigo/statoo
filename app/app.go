package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const version = "0.0.0"
const defTimeout = 10

var (
	argURL                string
	optVersionInformation *bool
	optTimeout            *int

	usage = `
usage: %[1]s [-flags] URL

  flags:

  -version        display version information (%s)
  -t, -timeout    default timeout in seconds  (default: %d)
  -h, -help       display help
  -verbose        verbose output              (default: false)

  examples:
  
  $ %[1]s "https://ugur.ozyilmazel.com"
  $ %[1]s -timeout 30 "https://ugur.ozyilmazel.com"

`
)

// CLIApplication represents app structure
type CLIApplication struct {
	Out io.Writer
}

// NewCLIApplication creates new CLIApplication instance
func NewCLIApplication() *CLIApplication {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage, os.Args[0], version, defTimeout)
	}

	optVersionInformation = flag.Bool("version", false, "")
	optTimeout = flag.Int("timeout", defTimeout, "")
	flag.IntVar(optTimeout, "t", defTimeout, "")
	flag.Parse()

	argURL = flag.Arg(0)

	return &CLIApplication{
		Out: os.Stdout,
	}
}

// Run executes main application
func (c *CLIApplication) Run() error {
	if *optVersionInformation {
		c.Version()
	}

	if err := c.Validate(); err != nil {
		return err
	}

	return nil
}

// Validate runs validations for flags
func (c *CLIApplication) Validate() error {
	if argURL == "" {
		return errors.New("please provide URL")
	}

	if argURL[:4] != "http" {
		return errors.New("URL should start with http:// or https://")
	}

	if *optTimeout > 100 || *optTimeout < 1 {
		return fmt.Errorf("invalid timeout value: %v", optTimeout)
	}
	return c.GetGivenURL()
}

// GetGivenURL fetches the given URL
func (c *CLIApplication) GetGivenURL() error {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
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

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}

	fmt.Fprintf(c.Out, "%d\n", resp.StatusCode)
	return nil
}

// Version returns the current version of CLIApplication
func (c *CLIApplication) Version() {
	fmt.Fprintf(os.Stderr, "%s\n", version)
	os.Exit(0)
}

package app_test

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vigo/statoo/app"
	"github.com/vigo/statoo/app/version"
)

func TestCustomHeadersFlag(t *testing.T) {
	var flags flag.FlagSet
	var h app.HeadersFlag

	flags.Init("test", flag.ContinueOnError)
	flags.Var(&h, "header", "usage")
	if err := flags.Parse([]string{"-header=foobar"}); err == nil {
		t.Error(err)
	}
	if err := flags.Parse([]string{"-header=foo.bar"}); err == nil {
		t.Error(err)
	}
	if err := flags.Parse([]string{"-header=foo.bar", "-header=foobar"}); err == nil {
		t.Error(err)
	}
	if err := flags.Parse([]string{"-header=foo:bar"}); err != nil {
		t.Error(err)
	}
}

func TestCustomAuthFlag(t *testing.T) {
	var flags flag.FlagSet
	var a app.BasicAuthFlag

	flags.Init("test", flag.ContinueOnError)
	flags.Var(&a, "auth", "usage")
	flags.Var(&a, "a", "usage")
	if err := flags.Parse([]string{"-a=foobar"}); err == nil {
		t.Error(err)
	}
	if err := flags.Parse([]string{"-auth=foo-bar"}); err == nil {
		t.Error(err)
	}
	if err := flags.Parse([]string{"-auth=foo:bar"}); err != nil {
		t.Error(err)
	}
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func gzipWrapper(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer func() { _ = gz.Close() }()
		gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		handler.ServeHTTP(gzw, r)
	})
}

func TestResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello world\n"))
	})

	handlerWithGZ := gzipWrapper(handler)

	cmd := app.NewCLIApplication()

	t.Run("test fake 200 reponse", func(t *testing.T) {
		cmd.Out = new(bytes.Buffer)
		ts := httptest.NewServer(handler)

		app.ArgURL = ts.URL
		if err := cmd.Run(); err != nil {
			t.Error(err)
		}
		*app.OptJSONOutput = false
	})

	t.Run("json reponse", func(t *testing.T) {
		buff := new(bytes.Buffer)
		cmd.Out = buff

		ts := httptest.NewServer(handler)
		app.ArgURL = ts.URL
		*app.OptJSONOutput = true

		if err := cmd.Run(); err != nil {
			t.Error(err)
		}

		body, _ := ioutil.ReadAll(buff)
		jr := new(app.JSONResponse)
		_ = json.Unmarshal(body, jr)
		if got := jr.Status; got != 200 {
			t.Errorf("want 200, got: %v", got)
		}
		*app.OptJSONOutput = false
	})

	t.Run("find text", func(t *testing.T) {
		buff := new(bytes.Buffer)
		cmd.Out = buff

		ts := httptest.NewServer(handler)
		app.ArgURL = ts.URL
		*app.OptJSONOutput = true
		*app.OptFind = "hello"

		if err := cmd.Run(); err != nil {
			t.Error(err)
		}

		body, _ := ioutil.ReadAll(buff)

		jr := new(app.JSONResponse)
		_ = json.Unmarshal(body, jr)
		if got := jr.Status; got != 200 {
			t.Errorf("want 200, got: %v", got)
		}
		if got := jr.Length; got != 12 {
			t.Errorf("want 12, got: %v", got)
		}
		if got := *jr.Found; got != true {
			t.Errorf("want true, got: %v", got)
		}
		if got := *jr.Find; got != "hello" {
			t.Errorf("want true, got: %v", got)
		}
		*app.OptJSONOutput = false
		*app.OptFind = ""
	})

	t.Run("gzip handler and find text", func(t *testing.T) {
		buff := new(bytes.Buffer)
		cmd.Out = buff

		ts := httptest.NewServer(handlerWithGZ)
		app.ArgURL = ts.URL

		*app.OptJSONOutput = true
		*app.OptFind = "hello"

		if err := cmd.Run(); err != nil {
			t.Error(err)
		}

		body, _ := ioutil.ReadAll(buff)
		jr := new(app.JSONResponse)
		_ = json.Unmarshal(body, jr)

		*app.OptJSONOutput = false
		app.OptFind = nil
	})

	t.Run("test empty URL arg", func(t *testing.T) {
		app.ArgURL = ""
		cmd.Out = new(bytes.Buffer)
		if got := cmd.Run(); got.Error() != "parse \"\": empty url" {
			t.Errorf("got: %v", got)
		}
	})

	t.Run("test URL w/o prefix", func(t *testing.T) {
		app.ArgURL = "vigo.io"
		cmd.Out = new(bytes.Buffer)
		if got := cmd.Run(); got.Error() != "parse \"vigo.io\": invalid URI for request" {
			t.Errorf("got: %v", got)
		}
	})

	t.Run("set errorious timeout", func(t *testing.T) {
		*app.OptTimeout = 200
		app.ArgURL = "https://vigo.io"
		cmd.Out = new(bytes.Buffer)
		if got := cmd.Run(); got.Error() != "invalid timeout value: 200" {
			t.Errorf("want nil, got: %v", got)
		}
	})

	t.Run("get version", func(t *testing.T) {
		*app.OptVersionInformation = true
		app.ArgURL = ""

		buff := new(bytes.Buffer)
		cmd.Out = buff
		if got := cmd.Run(); got != nil {
			t.Errorf("want nil, got: %v", got)
		}

		got := string(bytes.TrimSpace(buff.Bytes()))
		if got != version.Version {
			t.Errorf("want %v, got: %v", version.Version, got)
		}
		*app.OptVersionInformation = false
	})

	t.Run("bash completion", func(t *testing.T) {
		app.ArgURL = "bash-completion"

		buff := new(bytes.Buffer)
		cmd.Out = buff
		if got := cmd.Run(); got != nil {
			t.Errorf("want nil, got: %v", got)
		}
		got := string(bytes.TrimSpace(buff.Bytes()))
		if !strings.Contains(got, "__statoo_comp()") {
			t.Errorf("result should contain __statoo_comp() got: %v", got)
		}
		app.ArgURL = ""
	})
}

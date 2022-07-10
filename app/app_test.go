package app_test

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/vigo/statoo/app"
	"github.com/vigo/statoo/app/flags"
	"github.com/vigo/statoo/app/version"
)

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	n, err := w.Writer.Write(b)
	return n, fmt.Errorf("gzip error: %w", err)
}

func gzipWrapper(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer func() {
			if err := gz.Close(); err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
			}
		}()
		gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		handler.ServeHTTP(gzw, r)
	})
}

func TestResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Server", "FakeServer")
		_, _ = w.Write([]byte("hello world\n"))
	})

	handlerBasicAuth := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {
			_, _ = w.Write([]byte("basic auth, " + username + "," + password + "\n"))
		}
	})

	handlerWithGZ := gzipWrapper(handler)

	cmd := app.NewCLIApplication()

	t.Run("test empty URL arg", func(t *testing.T) {
		app.ArgURL = ""
		cmd.Out = new(bytes.Buffer)
		if err := cmd.Run(); err != nil {
			t.Errorf("want: nil, got: %v", err)
		}
	})

	t.Run("test fake 200 response", func(t *testing.T) {
		cmd.Out = new(bytes.Buffer)
		ts := httptest.NewServer(handler)

		app.ArgURL = ts.URL
		if err := cmd.Run(); err != nil {
			t.Error(err)
		}
		*app.OptJSONOutput = false
	})

	t.Run("basic auth", func(t *testing.T) {
		buff := new(bytes.Buffer)
		cmd.Out = buff

		ts := httptest.NewServer(handlerBasicAuth)
		app.ArgURL = ts.URL

		*app.OptJSONOutput = true
		*app.OptBasicAuth = "foo:bar"

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

	t.Run("json response", func(t *testing.T) {
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

	t.Run("find headers", func(t *testing.T) {
		buff := new(bytes.Buffer)
		cmd.Out = buff

		ts := httptest.NewServer(handler)
		app.ArgURL = ts.URL
		*app.OptJSONOutput = true
		responseHeaders := flags.ResponseHeadersFlag{
			"Content-Type: text/plain; charset=utf-8",
			"Server: FakeServer",
		}
		app.OptResponseHeaders = responseHeaders

		if err := cmd.Run(); err != nil {
			t.Error(err)
		}

		body, _ := ioutil.ReadAll(buff)

		jr := new(app.JSONResponse)
		_ = json.Unmarshal(body, jr)

		if jr.ResponseHeaders == nil {
			t.Error("must have response headers")
		}
		rh := *jr.ResponseHeaders
		val, ok := rh["Content-Type=text/plain; charset=utf-8"]
		if !ok {
			t.Error("Content-Type=text/plain; charset=utf-8 not found")
		}

		if !val {
			t.Error("value must be true")
		}

		val, ok = rh["Server=FakeServer"]
		if !ok {
			t.Error("Server not found")
		}
		if !val {
			t.Error("value must be true")
		}
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

	t.Run("test URL w/o prefix", func(t *testing.T) {
		app.ArgURL = "vigo.io"
		cmd.Out = new(bytes.Buffer)

		want := "url parse error: parse \"vigo.io\": invalid URI for request"
		if got := cmd.Run(); got.Error() != want {
			t.Errorf("want: %v, got: %v", want, got)
		}
	})

	t.Run("set errorious timeout max", func(t *testing.T) {
		*app.OptTimeout = 200
		app.ArgURL = "https://vigo.io"
		cmd.Out = new(bytes.Buffer)

		want := "invalid timeout value: 200"
		if got := cmd.Run(); got.Error() != want {
			t.Errorf("want: %v, got: %v", want, got)
		}
		*app.OptTimeout = 10
	})

	t.Run("set errorious timeout min", func(t *testing.T) {
		*app.OptTimeout = 0
		app.ArgURL = "https://vigo.io"
		cmd.Out = new(bytes.Buffer)

		want := "invalid timeout value: 0"
		if got := cmd.Run(); got.Error() != want {
			t.Errorf("want: %v, got: %v", want, got)
		}
		*app.OptTimeout = 10
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

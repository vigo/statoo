package app

import (
	"bytes"
	"testing"
)

func TestCLIApplication(t *testing.T) {
	cmd := NewCLIApplication()
	buff := new(bytes.Buffer)
	cmd.Out = buff

	t.Run("call w/o URL", func(t *testing.T) {
		if got := cmd.Run(); got.Error() != "please provide URL" {
			t.Errorf("got: %v", got)
		}
	})

	t.Run("URL w/o prefix", func(t *testing.T) {
		argURL = "vigo.io"
		if got := cmd.Run(); got.Error() != "URL should start with http:// or https://" {
			t.Errorf("got: %v", got)
		}
	})

	t.Run("set errorious timeout", func(t *testing.T) {
		*optTimeout = 200
		argURL = "https://vigo.io"

		if got := cmd.Run(); got.Error() != "invalid timeout value: 200" {
			t.Errorf("want nil, got: %v", got)
		}
	})

	t.Run("get version", func(t *testing.T) {
		*optVersionInformation = true

		if got := cmd.Run(); got != nil {
			t.Errorf("want nil, got: %v", got)
		}

		got := string(bytes.TrimSpace(buff.Bytes()))
		if got != version {
			t.Errorf("want %v, got: %v", version, got)
		}
	})
}

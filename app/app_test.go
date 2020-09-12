package app

import (
	"bytes"
	"os"
	"testing"
)

var cmd *CLIApplication

func TestMain(m *testing.M) {
	cmd = NewCLIApplication()
	os.Exit(m.Run())
}

func TestAppVersion(t *testing.T) {
	t.Run("app should have a version information", func(t *testing.T) {

		buff := new(bytes.Buffer)

		*optVersionInformation = true
		cmd.Out = buff
		cmd.Run()

		curVersion := string(bytes.TrimSpace(buff.Bytes()))
		if curVersion != version {
			t.Errorf("want: %s, got: %s", version, curVersion)
		}
	})
}

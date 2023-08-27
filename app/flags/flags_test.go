package flags_test

import (
	"flag"
	"testing"

	"github.com/vigo/statoo/v2/app/flags"
)

func TestCustomRequestHeadersFlag(t *testing.T) {
	var tflags flag.FlagSet
	var h flags.RequestHeadersFlag

	tflags.Init("test", flag.ContinueOnError)
	tflags.Var(&h, "request-header", "usage")

	if err := tflags.Parse([]string{"-request-header="}); err == nil {
		t.Error(err)
	}
	if err := tflags.Parse([]string{"-request-header=foobar"}); err == nil {
		t.Error(err)
	}
	if err := tflags.Parse([]string{"-request-header=foo.bar"}); err == nil {
		t.Error(err)
	}
	if err := tflags.Parse([]string{"-request-header=foo.bar", "-header=foobar"}); err == nil {
		t.Error(err)
	}
	if err := tflags.Parse([]string{"-request-header=foo;bar"}); err == nil {
		t.Error(err)
	}
	if err := tflags.Parse([]string{"-request-header=foo:bar:baz"}); err == nil {
		t.Error(err)
	}
	if err := tflags.Parse([]string{"-request-header=foo:bar"}); err != nil {
		t.Error(err)
	}
}

func TestCustomResponseHeadersFlag(t *testing.T) {
	var tflags flag.FlagSet
	var h flags.ResponseHeadersFlag

	tflags.Init("test", flag.ContinueOnError)
	tflags.Var(&h, "response-header", "usage")

	if err := tflags.Parse([]string{"-response-header="}); err == nil {
		t.Error(err)
	}
	if err := tflags.Parse([]string{"-response-header=foobar"}); err == nil {
		t.Error(err)
	}
	if err := tflags.Parse([]string{"-response-header=foo.bar"}); err == nil {
		t.Error(err)
	}
	if err := tflags.Parse([]string{"-response-header=foo.bar", "-header=foobar"}); err == nil {
		t.Error(err)
	}
	if err := tflags.Parse([]string{"-response-header=foo;bar"}); err == nil {
		t.Error(err)
	}
	if err := tflags.Parse([]string{"-response-header=foo:bar:baz"}); err == nil {
		t.Error(err)
	}
	if err := tflags.Parse([]string{"-response-header=foo:bar"}); err != nil {
		t.Error(err)
	}
}

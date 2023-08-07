![Version](https://img.shields.io/badge/version-2.0.2-orange.svg)
![Go](https://img.shields.io/github/go-mod/go-version/vigo/statoo)
[![Documentation](https://godoc.org/github.com/vigo/statoo?status.svg)](https://pkg.go.dev/github.com/vigo/statoo)
[![Go Report Card](https://goreportcard.com/badge/github.com/vigo/statoo)](https://goreportcard.com/report/github.com/vigo/statoo)
[![Build Status](https://travis-ci.org/vigo/statoo.svg?branch=main)](https://travis-ci.org/vigo/statoo)
![Go Build Status](https://github.com/vigo/statoo/actions/workflows/go.yml/badge.svg)
![GolangCI-Lint Status](https://github.com/vigo/statoo/actions/workflows/golang-lint.yml/badge.svg)
![Docker Lint Status](https://github.com/vigo/statoo/actions/workflows/docker-lint.yml/badge.svg)
[![codecov](https://codecov.io/gh/vigo/statoo/branch/main/graph/badge.svg?token=BTVK8VKVZM)](https://codecov.io/gh/vigo/statoo)
![Docker Pulls](https://img.shields.io/docker/pulls/vigo/statoo)
![Docker Size](https://img.shields.io/docker/image-size/vigo/statoo)
![Docker Build Status](https://github.com/vigo/statoo/actions/workflows/dockerhub.yml/badge.svg)
![Powered by Rake](https://img.shields.io/badge/powered_by-rake-blue?logo=ruby)


# Statoo

A super basic http tool that makes only `GET` request to given URL and returns
status code of the response. Well, if you are `curl` or `http` (*httpie*) user,
you can make the same kind of request and get a kind-of same response since
`statoo` is way better simple :)

`statoo` injects `Accept-Encoding: gzip` request header to every http request!

## Installation

You can install from the source;

```bash
go install github.com/vigo/statoo@latest
```

or, you can install from `brew`:

```bash
brew tap vigo/statoo
brew install statoo
```

## Usage:

```bash
statoo -h
```

```bash
usage: ./statoo [-flags] URL

  flags:

  -version           display version information (%s)
  -verbose           verbose output (default: false)
  -request-header    request header, multiple allowed, "Key: Value", case sensitive
  -response-header   response header for lookup -json is set, multiple allowed, "Key: Value"
  -t, -timeout       default timeout in seconds (default: %d, min: %d, max: %d)
  -h, -help          display help
  -j, -json          provides json output
  -f, -find          find text in response body if -json is set, case sensitive
  -a, -auth          basic auth "username:password"
  -s, -skip          skip certificate check and hostname in that certificate (default: false)
  -commithash        displays current build/commit hash (%s)

  examples:
  
  $ ./statoo "https://ugur.ozyilmazel.com"
  $ ./statoo -timeout 30 "https://ugur.ozyilmazel.com"
  $ ./statoo -verbose "https://ugur.ozyilmazel.com"
  $ ./statoo -json https://vigo.io
  $ ./statoo -json -find "python" https://vigo.io
  $ ./statoo -json -find "Python" https://vigo.io
  $ ./statoo -json -find "Golang" https://vigo.io
  $ ./statoo -request-header "Authorization: Bearer TOKEN" https://vigo.io
  $ ./statoo -request-header "Authorization: Bearer TOKEN" -header "X-Api-Key: APIKEY" https://vigo.io
  $ ./statoo -auth "user:secret" https://vigo.io
  $ ./statoo -json -response-header "Server: GitHub.com" https://vigo.io
  $ ./statoo -json -response-header "Server: GitHub.com" -response-header "Foo: bar" https://vigo.io
```

Let’s try:

```bash
statoo "https://ugur.ozyilmazel.com"
# 200
```

```bash
statoo -verbose "https://ugur.ozyilmazel.com"
# https://ugur.ozyilmazel.com -> 200
```

or;

```bash
statoo -json https://vigo.io
```

response;

```json
{
    "url": "https://vigo.io",
    "status": 200,
    "checked_at": "2021-05-13T18:09:26.342012Z",
    "elapsed": 210.587871,
    "skipcc": false
}
```

`elapsed` represents response is in milliseconds.

Let’s find text inside of the response body. This feature is only available if
the `-json` flag is set! `length` represents response size in bytes
(*gzipped*) when you search something in body!

```bash
statoo -json -find "Golang" https://vigo.io
```

```json
{
    "url": "https://vigo.io",
    "status": 200,
    "checked_at": "2022-01-26T20:08:33.735768Z",
    "elapsed": 242.93925,
    "length": 7827,
    "find": "Golang",
    "found": true,
    "skipcc": false
}
```

```bash
statoo -json -find "golang" https://vigo.io # case sensitive
```

```json
{
    "url": "https://vigo.io",
    "status": 200,
    "checked_at": "2022-01-26T20:14:03.487002Z",
    "elapsed": 253.665083,
    "length": 7827,
    "find": "golang",
    "found": false,
    "skipcc": false
}
```

You can add basic authentication via `-auth` flag

```bash
statoo -auth "username:password" https://your.basic.auth.url
```

Now you can pass multiple `-request-header` flags:

```bash
statoo -request-header "Key1: Value1" -request-header "Key2: Value2" "https://ugur.ozyilmazel.com"
```

You can query/search for response headers. You can pass multiple values, all
**case sensitive**!. Let’s lookup for `Server` and `Foo` response header values.
`Server` value should be `GitHub.com` and `Foo` value should be `bar`:

```bash
statoo -json -response-header "Server: GitHub.com" -response-header "Foo: bar" https://vigo.io
```

Response:

```json
{
    "url": "https://vigo.io",
    "status": 200,
    "checked_at": "2022-07-09T17:51:14.792987Z",
    "elapsed": 305.502833,
    "skipcc": false,
    "response_headers": {
        "Foo=bar": false,
        "Server=GitHub.com": true
    }
}
```

`Server` response header matches exactly!

It’s better to pipe `-json` output to `jq` or `python -m json.tool` for pretty
print :)

That’s it!

Bash completions is available via;

```bash
eval "$(statoo bash-completion)"
```

**New**

You can check current build/commit hash via;

```bash
statoo -commithash
```

---

## Rake Tasks

```bash
$ rake -T

rake default               # show avaliable tasks (default task)
rake docker:lint           # lint Dockerfile
rake release[revision]     # release new version major,minor,patch, default: patch
rake test:run[verbose]     # run tests, generate coverage
rake test:show_coverage    # show coverage after running tests
rake test:update_coverage  # update coverage value in README
```

---

## Docker

https://hub.docker.com/r/vigo/statoo/

```bash
# latest
docker run vigo/statoo -h
docker run vigo/statoo -json -find "Meetup organization" https://vigo.io
```

---

## Contributor(s)

* [Uğur "vigo" Özyılmazel](https://github.com/vigo) - Creator, maintainer
* [Erman İmer](https://github.com/ermanimer) - Contributor
* [Rishi Kumar Ray](https://github.com/RishiKumarRay) - Contributor

---

## Contribute

All PR’s are welcome!

1. `fork` (https://github.com/vigo/statoo/fork)
1. Create your `branch` (`git checkout -b my-feature`)
1. `commit` yours (`git commit -am 'add some functionality'`)
1. `push` your `branch` (`git push origin my-feature`)
1. Than create a new **Pull Request**!

This project is intended to be a safe, welcoming space for collaboration, and
contributors are expected to adhere to the [code of conduct][coc].

---

## License

This project is licensed under MIT

[coc]: https://github.com/vigo/statoo/blob/main/CODE_OF_CONDUCT.md
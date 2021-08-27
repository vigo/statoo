![Version](https://img.shields.io/badge/version-1.2.3-orange.svg)
![Go](https://img.shields.io/badge/go-1.16-black.svg)
[![Documentation](https://godoc.org/github.com/vigo/statoo?status.svg)](https://pkg.go.dev/github.com/vigo/statoo)
[![Go Report Card](https://goreportcard.com/badge/github.com/vigo/statoo)](https://goreportcard.com/report/github.com/vigo/statoo)
[![Build Status](https://travis-ci.org/vigo/statoo.svg?branch=main)](https://travis-ci.org/vigo/statoo)
![Go Build Status](https://github.com/vigo/statoo/actions/workflows/go.yml/badge.svg)
![Test Coverage](https://img.shields.io/badge/coverage-80.2%25-orange.svg)

# Statoo

A super basic http tool that makes only `GET` request to given URL and returns
status code of the response. Well, if you are `curl` or `http` (*httpie*) user,
you can make the same kind of request and get a kind-of same response since
`statoo` is way better simple :)

`statoo` injects `Accept-Encoding: gzip` request header to every http request!

## Installation

You can install from the source;

```bash
$ go get github.com/vigo/statoo
```

or, you can install from `brew`:

```bash
$ brew tap vigo/statoo
$ brew install statoo
```

## Usage:

```bash
$ statoo -h

usage: ./statoo [-flags] URL

  flags:

  -version        display version information (X.X.X)
  -verbose        verbose output              (default: false)
  -header         request header, multiple allowed
  -t, -timeout    default timeout in seconds  (default: 10)
  -h, -help       display help
  -j, -json       provides json output
  -f, -find       find text in response body if -json is set
  -a, -auth       basic auth "username:password"

  examples:
  
  $ ./statoo "https://ugur.ozyilmazel.com"
  $ ./statoo -timeout 30 "https://ugur.ozyilmazel.com"
  $ ./statoo -verbose "https://ugur.ozyilmazel.com"
  $ ./statoo -json https://vigo.io
  $ ./statoo -json -find "python" https://vigo.io
  $ ./statoo -header "Authorization: Bearer TOKEN" https://vigo.io
  $ ./statoo -header "Authorization: Bearer TOKEN" -header "X-Api-Key: APIKEY" https://vigo.io
  $ ./statoo -json -find "Meetup organization" https://vigo.io
  $ ./statoo -auth "user:secret" https://vigo.io
```

Let’s try:

```bash
$ statoo "https://ugur.ozyilmazel.com"
200

$ statoo -verbose "https://ugur.ozyilmazel.com"
https://ugur.ozyilmazel.com -> 200
```

or;

```bash
$ statoo -json https://vigo.io
```

response;

```json
{
    "url": "https://vigo.io",
    "status": 200,
    "checked_at": "2021-05-13T18:09:26.342012Z",
    "elapsed": 210.587871,
    "length": 1453
}
```

- `elapsed` represents response is in milliseconds.
- `length` represents response size in bytes (*gzipped*)

Let’s find text inside of the response body. This feature is only available
if the `-json` flag is set!

```bash
$ statoo -json -find "Meetup organization" https://vigo.io
{
    "url": "https://vigo.io",
    "status": 200,
    "checked_at": "2021-05-13T18:10:38.196705Z",
    "elapsed": 183.128016,
    "length": 1453,
    "find": "Meetup organization",
    "found": true
}

$ statoo -json -find "meetup organization" https://vigo.io # case sensitive
{
    "url": "https://vigo.io",
    "status": 200,
    "checked_at": "2021-05-13T18:10:58.100932Z",
    "elapsed": 189.403753,
    "length": 1453,
    "find": "meetup organization",
    "found": false
}
```

You can add basic authentication via `-auth` flag

```bash
$ statoo -auth "username:password" https://your.basic.auth.url
```

Now you can pass multiple `-header` flags:

```bash
$ statoo -header "Key1: Value1" -header "Key2: Value2" "https://ugur.ozyilmazel.com"
```

It’s better to pipe `-json` output to `jq` or `python -m json.tool` for pretty print :)

That’s it!

Bash completions is available via;

```bash
$ eval "$(statoo bash-completion)"
```

---

## Rake Tasks

```bash
$ rake -T

rake default               # show avaliable tasks (default task)
rake docker:build          # Build
rake docker:rmi            # Delete image
rake docker:run            # Run
rake release[revision]     # Release new version major,minor,patch, default: patch
rake test:run[verbose]     # run tests, generate coverage
rake test:show_coverage    # show coverage after running tests
rake test:update_coverage  # update coverage value in README
```

---

## Docker

build:

```bash
$ docker build . -t statoo
```

run:

```bash
$ docker run -i -t statoo:latest statoo -h
$ docker run -i -t statoo:latest statoo -json -find "Meetup organization" https://vigo.io
```

---

## Contributor(s)

* [Uğur "vigo" Özyılmazel](https://github.com/vigo) - Creator, maintainer
* [Erman İmer](https://github.com/ermanimer) - Contributor

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
![Version](https://img.shields.io/badge/version-0.2.2-orange.svg)
![Go](https://img.shields.io/badge/go-1.16-black.svg)
[![Documentation](https://godoc.org/github.com/vigo/statoo?status.svg)](https://pkg.go.dev/github.com/vigo/statoo)
[![Go Report Card](https://goreportcard.com/badge/github.com/vigo/statoo)](https://goreportcard.com/report/github.com/vigo/statoo)
[![Build Status](https://travis-ci.org/vigo/statoo.svg?branch=main)](https://travis-ci.org/vigo/statoo)
![Go Build Status](https://github.com/vigo/statoo/actions/workflows/go.yml/badge.svg)

# Statoo

A super basic http tool that makes only `GET` request to given URL and returns
status code of the response. Well, if you are `curl` or `http` (*httpie*) user,
you can make the same kind of request and get a kind-of same response since
`statoo` is way better simple :)

## Install

```bash
$ go get -u github.com/vigo/statoo
```

## Usage:

```bash
$ statoo -h


usage: statoo [-flags] URL

  flags:

  -version        display version information (X.X.X)
  -t, -timeout    default timeout in seconds  (default: 10)
  -h, -help       display help
  -json           provides json output
  -verbose        verbose output              (default: false)
  -header         request header, multiple allowed
  -find           find text in repsonse body if -json is set

  examples:
  
  $ statoo "https://ugur.ozyilmazel.com"
  $ statoo -timeout 30 "https://ugur.ozyilmazel.com"
  $ statoo -verbose "https://ugur.ozyilmazel.com"
  $ statoo -json https://vigo.io
  $ statoo -header "Authorization: Bearer TOKEN" https://vigo.io
  $ statoo -header "Authorization: Bearer TOKEN" -header "X-Api-Key: APIKEY" https://vigo.io
  $ statoo -json -find "Meetup organization" https://vigo.io

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
    "checked_at": "2021-05-10T10:26:04.972779Z",
    "response_duration": 196.038446
}
```

`response_duration` is in milliseconds.

Let’s find text inside of the response body. This feature is only available
if the `-json` flag is set!

```bash
$ statoo -json -find "Meetup organization" https://vigo.io
{
    "url": "https://vigo.io",
    "status": 200,
    "checked_at": "2021-05-10T10:26:38.481739Z",
    "find": "Meetup organization",
    "found": true,
    "response_duration": 1119.754662
}

$ statoo -json -find "meetup organization" https://vigo.io # case sensitive
{
    "url": "https://vigo.io",
    "status": 200,
    "checked_at": "2021-05-10T10:27:05.754963Z",
    "find": "meetup organization",
    "found": false,
    "response_duration": 98.336486
}
```

Now you can pass multiple `-header` flags:

```bash
$ statoo -header "Key1: Value1" -header "Key2: Value2" "https://ugur.ozyilmazel.com"
```

It’s better to pipe `-json` output to `jq` for pretty print :)

That’s it!

---

## Rake Tasks

```bash
$ rake -T

rake default            # show avaliable tasks (default task)
rake docker:build       # Build
rake docker:rmi         # Delete image
rake docker:run         # Run
rake release[revision]  # Release new version major,minor,patch, default: patch
rake test[verbose]      # run tests
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

## Contributer(s)

* [Uğur "vigo" Özyılmazel](https://github.com/vigo) - Creator, maintainer

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
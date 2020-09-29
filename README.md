![Version](https://img.shields.io/badge/version-0.1.2-orange.svg)
![Go](https://img.shields.io/badge/go-1.15.1-black.svg)
[![Documentation](https://godoc.org/github.com/vigo/statoo?status.svg)](https://pkg.go.dev/github.com/vigo/statoo)
[![Go Report Card](https://goreportcard.com/badge/github.com/vigo/statoo)](https://goreportcard.com/report/github.com/vigo/statoo)
[![Build Status](https://travis-ci.org/vigo/statoo.svg?branch=main)](https://travis-ci.org/vigo/statoo)

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

  examples:
  
  $ statoo "https://ugur.ozyilmazel.com"
  $ statoo -timeout 30 "https://ugur.ozyilmazel.com"
  $ statoo -verbose "https://ugur.ozyilmazel.com"
  $ statoo -json http://vigo.io
  $ statoo -header "Authorization: Bearer TOKEN" http://vigo.io
  $ statoo -header "Authorization: Bearer TOKEN" -header "X-Api-Key: APIKEY" http://vigo.io

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
$ statoo -json http://vigo.io
```

response;

```json
{
  "url": "http://vigo.io",
  "status": 200,
  "checked_at": "2020-09-18T04:56:14.664255Z"
}
```

Now you can pass multiple `-header` flags:

```bash
$ status -header "Key1: Value1" -header "Key2: Value2" "https://ugur.ozyilmazel.com"
```

It’s better to pipe `-json` output to `jq` for pretty print :)

That’s it!

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
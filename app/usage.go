package app

var usage = `
usage: %[1]s [-flags] URL

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
  
  $ %[1]s "https://ugur.ozyilmazel.com"
  $ %[1]s -timeout 30 "https://ugur.ozyilmazel.com"
  $ %[1]s -verbose "https://ugur.ozyilmazel.com"
  $ %[1]s -json https://vigo.io
  $ %[1]s -json -find "python" https://vigo.io
  $ %[1]s -json -find "Python" https://vigo.io
  $ %[1]s -json -find "Golang" https://vigo.io
  $ %[1]s -request-header "Authorization: Bearer TOKEN" https://vigo.io
  $ %[1]s -request-header "Authorization: Bearer TOKEN" -header "X-Api-Key: APIKEY" https://vigo.io
  $ %[1]s -auth "user:secret" https://vigo.io
  $ %[1]s -json -response-header "Server: GitHub.com" https://vigo.io
  $ %[1]s -json -response-header "Server: GitHub.com" -response-header "Foo: bar" https://vigo.io

`

FROM golang:1.19-alpine AS builder
WORKDIR /go/src/github.com/vigo/statoo
COPY . .
RUN apk add --no-cache git=2.36.3-r0 \
    ca-certificates=20220614-r0 \
    && CGO_ENABLED=0 \
    GOOS=linux \
    go build -ldflags="-X 'github.com/vigo/statoo/app/version.CommitHash=$(git rev-parse HEAD)'" -a -installsuffix cgo -o statoo .

FROM alpine:3.15
RUN apk --no-cache add 
COPY --from=builder /go/src/github.com/vigo/statoo/statoo /bin/statoo
ENTRYPOINT ["/bin/statoo"]
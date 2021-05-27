FROM golang:1.16-alpine3.13 AS builder

WORKDIR /go/src/app
COPY go.mod go.sum /go/src/app/
ADD . /go/src/app

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -tags netgo \
    -mod vendor \
    -o /go/bin/ingest \
    -ldflags="-s -w" \
    .

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /go/bin/ingest .
ENTRYPOINT ["./ingest"]


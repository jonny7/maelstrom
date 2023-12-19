FROM golang:1.21.5-alpine as builder

ENV CGO_ENABLED=0

RUN apk --no-cache add ca-certificates

WORKDIR /go/src/github.com/jonny7/maelstrom

COPY . .

RUN go mod download
RUN go build -o bin/maelstrom ./cmd/maelstrom

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/jonny7/maelstrom/bin/maelstrom ./maelstrom

CMD ["./maelstrom"]
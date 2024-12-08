FROM golang:alpine as builder

WORKDIR /app
RUN apk add --no-cache make
COPY go.mod go.sum ./
RUN go mod download

COPY ./ /app
RUN make build

FROM alpine:latest

COPY --from=builder /app/bin/main /usr/local/bin/maddr-proxy

ENTRYPOINT [ "/usr/local/bin/maddr-proxy" ]

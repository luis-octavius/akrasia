FROM golang:1.25.1-alpine AS builder 

RUN apk add --no-cache git sqlite 

WORKDIR /app 

COPY go.mod go.sum ./ 
RUN go mod download 

COPY . . 

RUN go build -o akrasia . 

FROM alpine:latest 

RUN apk add --no-cache sqlite 

WORKDIR /root/

COPY --from=builder /app/akrasia . 

VOLUME ["/data"]

ENTRYPOINT ["./akrasia"]

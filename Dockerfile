FROM golang:1.25.1-alpine AS builder 

RUN apk add --no-cache git sqlite 

WORKDIR /app 

COPY go.mod go.sum ./ 
RUN go mod download 

COPY . . 

COPY .env .env 

COPY akrasia.db /data/akrasia.db 

RUN go build -o akrasia . 

FROM alpine:latest 

RUN apk add --no-cache sqlite 

WORKDIR /root/

COPY --from=builder /app/akrasia . 
COPY --from=builder /app/.env .env   
COPY --from=builder /app/akrasia.db /data/akrasia.db  

VOLUME ["/data"]

ENV DB_PATH="/data/akrasia.db"

ENTRYPOINT ["./akrasia"]

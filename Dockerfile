# --- build stage ---
FROM golang:1.27.1-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o akrasia .

# --- runtime stage ---
FROM alpine:latest

# tzdata is necessary, SQLITe (date('now','localtime')) and time.Now() from Go 
# depends on the container fuse. Without that, alpine runs in UTC and the day of the task 
# turns out, in practice, dealocated 3h from the real hour. 
RUN apk add --no-cache tzdata
ENV TZ=America/Sao_Paulo

# non-root user, with a owned directory for work
RUN adduser -D -h /home/akrasia akrasia

WORKDIR /home/akrasia

COPY --from=builder /app/akrasia .

# points explicitely to the path that the volume exposes - without that, 
# GetDbPath() goes in os.UserConfigDir() (~/.config/akrasia/akrasia.db),
# it turns out that will be out of the volume and it is lost in every recreation
ENV DB_PATH=/data/akrasia.db

RUN mkdir -p /data && chown -R akrasia:akrasia /data /home/akrasia

USER akrasia

VOLUME ["/data"]

ENTRYPOINT ["./akrasia"]

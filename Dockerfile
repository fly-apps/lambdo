FROM golang:1.21 as builder

WORKDIR /app
COPY . /app
RUN GOOS=linux GOARCH=amd64 go build -o /app/lambdo

FROM debian:bookworm-slim

COPY --from=builder /app/lambdo /usr/local/bin/lambdo

CMD ["lambdo"]

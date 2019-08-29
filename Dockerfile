FROM golang:1.12.9-alpine3.10 AS builder

WORKDIR /root/app
COPY . .
RUN GOOS=linux go build -mod vendor -o app main.go

FROM alpine:latest
RUN apk add --no-cache \
		ca-certificates
WORKDIR /bin/
COPY --from=builder /root/app/app .
ENTRYPOINT [ "app","-stdout" ]


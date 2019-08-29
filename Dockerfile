FROM golang:1.12.9-alpine3.10 AS builder

WORKDIR /root/app
COPY . .
RUN apk add git
RUN go mod vendor
RUN GOOS=linux go build -o app main.go

FROM alpine:latest
RUN apk add --no-cache \
		ca-certificates
WORKDIR /bin/
COPY --from=builder /root/app/app .
CMD [ "app","-stdout" ]
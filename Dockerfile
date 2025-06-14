FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED=0
ENV GOPROXY=https://goproxy.cn,direct
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

RUN apk update --no-cache && apk add --no-cache tzdata

WORKDIR /build

# dependency correlation
ADD vendor ./vendor
ADD go.mod .
ADD go.sum .
RUN go mod verify

COPY . .
RUN go build -ldflags="-s -w" -o /app/actbot main.go


FROM alpine:latest

# certificates and time zones
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai
ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /app/actbot /app/actbot

RUN chmod +x /app/actbot

ENTRYPOINT ["/app/actbot"]
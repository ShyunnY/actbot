FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED=0

WORKDIR /build

# dependency correlation
ADD vendor ./vendor
ADD go.mod .
ADD go.sum .
RUN go mod verify

COPY . .
RUN go build -ldflags="-s -w" -o /app/actbot .


FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/actbot /app/actbot

RUN chmod +x /app/actbot

ENTRYPOINT ["/app/actbot"]
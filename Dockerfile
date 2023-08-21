FROM golang:1.21-alpine AS builder
WORKDIR /build

COPY ./go.mod .
COPY ./go.sum .

RUN go mod download

COPY . .

RUN go build -o app cmd/main/main.go

FROM alpine

RUN apk update && apk upgrade

# Reduce image size
RUN rm -rf /var/cache/apk/* && \
    rm -rf /tmp/*

# Avoid running code as a root user
RUN adduser -D appuser
USER appuser

WORKDIR /app

COPY --from=builder /build/app /app/app

ENV CONFIG_PATH ""
ENV HTTP_ADDR ":8888"
EXPOSE 8888

CMD ["./app"]

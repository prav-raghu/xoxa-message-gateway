# syntax=docker/dockerfile:1
FROM golang:1.25-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG CMD_PATH=./cmd/api
RUN CGO_ENABLED=0 go build -o /out/app ${CMD_PATH}

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app

COPY --from=build /out/app ./app
COPY docs ./docs
COPY migrations ./migrations

EXPOSE 8080 9090
ENTRYPOINT ["./app"]

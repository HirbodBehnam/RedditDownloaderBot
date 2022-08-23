# syntax=docker/dockerfile:1

## Build
FROM golang:1.18-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -o /reddit-downloader-bot ./cmd/RedditDownloaderBot

## Deploy
FROM alpine:latest

WORKDIR /

# Add ffmpeg
RUN apk update
RUN apk add ffmpeg

COPY --from=build /reddit-downloader-bot /reddit-downloader-bot

USER nobody

ENTRYPOINT ["/reddit-downloader-bot"]

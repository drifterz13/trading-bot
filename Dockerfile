FROM golang:1.16-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -o /bot

FROM alpine:edge

WORKDIR /

RUN mkdir -p data
COPY --from=build /bot /bot

ENTRYPOINT ["/bot"]
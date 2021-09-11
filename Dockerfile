FROM golang:1.16-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /robo

FROM alpine:edge

WORKDIR /

RUN mkdir -p data
COPY --from=build /robo /robo

ENTRYPOINT ["/robo"]
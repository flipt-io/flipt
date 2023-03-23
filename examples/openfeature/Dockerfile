ARG GO_VERSION=1.20

FROM golang:${GO_VERSION}-alpine

RUN apk update && apk add --no-cache git

WORKDIR /example
COPY . .

RUN go mod tidy

ENTRYPOINT ["go", "run", "-tags", "example", "main.go"]

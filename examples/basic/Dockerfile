ARG GO_VERSION=1.23

FROM golang:${GO_VERSION}-alpine

RUN apk add --no-cache git

WORKDIR /example
COPY . .

RUN go mod tidy

CMD ["go", "run", "-tags", "example", "main.go"]

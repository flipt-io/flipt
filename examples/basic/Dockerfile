FROM golang:1.11-alpine

RUN apk update && apk add --no-cache git

WORKDIR /example
COPY . .

RUN go get "github.com/markphelps/flipt-grpc-go" \
    "google.golang.org/grpc"

ENTRYPOINT ["go", "run", "-tags", "example", "main.go"]

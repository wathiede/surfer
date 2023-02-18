FROM golang:alpine

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
COPY vendor/github.com/prometheus/client_golang/go.mod  ./vendor/github.com/prometheus/client_golang/
#COPY  vendor/github.com/prometheus/client_golang/go.sum ./vendor/github.com/prometheus/client_golang/
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/ ./...

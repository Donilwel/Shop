FROM golang:1.23.0

WORKDIR /usr/src/app

COPY . .

RUN go mod download

RUN go test -v ./...

CMD ["go", "run", "cmd/main.go"]
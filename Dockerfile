FROM golang:1.23.0

WORKDIR /usr/src/app

COPY . .

CMD ["go", "run", "cmd/main.go"]
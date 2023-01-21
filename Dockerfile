FROM golang:1.16-alpine

WORKDIR $GOPATH/src/app/

COPY . .

RUN go mod download
RUN go mod verify

RUN go build cmd/web_server/main.go

CMD ["./main"]

FROM golang:latest

ENV GOPATH=/
COPY ./ ./

RUN go mod download
RUN go build -o main .

CMD ["./main"]
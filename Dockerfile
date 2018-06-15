FROM golang:1.10.3-alpine3.7

ADD . /go/src/go-nsq-test
RUN apk update && apk add git
RUN go get github.com/nsqio/go-nsq
WORKDIR /go/src/go-nsq-test

CMD ["go", "run", "main.go"]

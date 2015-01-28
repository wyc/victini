FROM golang
MAINTAINER Aditya Mukerjee <dev@chimeracoder.net>


ADD . /go/src/github.com/wyc/victini


RUN go get github.com/wyc/victini
RUN go install github.com/wyc/victini


WORKDIR /go/src/github.com/wyc/victini

ENTRYPOINT /go/bin/victini

EXPOSE 8000






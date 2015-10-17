FROM golang:1.5
MAINTAINER xtaci <daniel820313@gmail.com>
ENV GOBIN /go/bin
COPY . /go
WORKDIR /go
ENV GOPATH /go:/go/.godeps
RUN go install chat
RUN rm -rf pkg src .godeps
RUN mkdir /data
ENTRYPOINT /go/bin/chat
EXPOSE 50008

FROM golang:latest
MAINTAINER xtaci <daniel820313@gmail.com>
COPY . /go/src/chat
RUN go install chat
ENTRYPOINT /go/bin/chat
EXPOSE 50008
RUN mkdir /data
VOLUME /data

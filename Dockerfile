FROM golang:latest
MAINTAINER xtaci <daniel820313@gmail.com>
ENV GOBIN /go/bin
COPY . /go
WORKDIR /go
RUN go install chat
RUN rm -rf pkg src
ENTRYPOINT /go/bin/chat
EXPOSE 50008
RUN mkdir /data
VOLUME /data

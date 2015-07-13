package main

import (
	"net"
	"os"

	log "github.com/gonet2/libs/nsq-logger"
	_ "github.com/gonet2/libs/statsd-pprof"
	"google.golang.org/grpc"
)

import (
	pb "proto"
)

const (
	_port = ":50008"
)

func main() {
	log.SetPrefix(SERVICE)
	// 监听
	lis, err := net.Listen("tcp", _port)
	if err != nil {
		log.Critical(err)
		os.Exit(-1)
	}
	log.Info("listening on ", lis.Addr())

	// 注册服务
	s := grpc.NewServer()
	ins := &server{}
	ins.init()
	pb.RegisterChatServiceServer(s, ins)
	// 开始服务
	s.Serve(lis)
}

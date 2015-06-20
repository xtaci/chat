package main

import (
	"golang.org/x/net/context"
)

import (
	pb "proto"
)

const (
	SERVICE = "[CHAT]"
)

const (
	BOLTDB_FILE   = "/data/CHAT.DAT"
	BOLTDB_BUCKET = "CHAT"
)

type server struct {
}

func (s *server) init() {
}

func (s *server) Receive(p *pb.Chat_Nil, stream pb.ChatService_ReceiveServer) error {
	return nil
}

func (s *server) Send(ctx context.Context, msg *pb.Chat_Message) (*pb.Chat_Nil, error) {
	return nil, nil
}

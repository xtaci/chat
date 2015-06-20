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
	BOLTDB_FILE          = "/data/CHAT.DAT"
	BOLTDB_P2P_BUCKET    = "P2P"
	BOLTDB_GROUP_BUCKET  = "GROUP"
	BOLTDB_GLOBAL_BUCKET = "GLOBAL"
)

type msg_queue struct {
	msgs []pb.Chat_Message
}

type server struct {
	p2p    map[int32]*msg_queue
	group  map[int32]*msg_queue
	global msg_queue
}

func (s *server) init() {
}

func (s *server) Receive(p *pb.Chat_Nil, stream pb.ChatService_ReceiveServer) error {
	return nil
}

func (s *server) Send(ctx context.Context, msg *pb.Chat_Message) (*pb.Chat_Nil, error) {
	return nil, nil
}

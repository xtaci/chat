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
	MAX_QUEUE_SIZE       = 128 // num of message kept
)

// messageq queue
type MsgQueue struct {
	Messages []pb.Chat_Message
}

type server struct {
	Users  map[int32]*MsgQueue
	Groups map[int32]*MsgQueue
	Global MsgQueue
}

func (s *server) init() {
	s.Users = make(map[int32]*MsgQueue)
	s.Groups = make(map[int32]*MsgQueue)
}

func (s *server) Receive(p *pb.Chat_Nil, stream pb.ChatService_ReceiveServer) error {
	return nil
}

func (s *server) Send(ctx context.Context, msg *pb.Chat_Message) (*pb.Chat_Nil, error) {
	return nil, nil
}

func (s *server) Inbox(context.Context, *pb.Chat_Id) (*pb.Chat_MessageList, error) {
	return nil, nil
}
func (s *server) GroupInbox(context.Context, *pb.Chat_Id) (*pb.Chat_MessageList, error) {
	return nil, nil
}
func (s *server) GlobalInbox(context.Context, *pb.Chat_Id) (*pb.Chat_MessageList, error) {
	return nil, nil
}
func (s *server) CreateUser(context.Context, *pb.Chat_Id) (*pb.Chat_Nil, error) {
	return nil, nil
}

func (s *server) CreateGroup(context.Context, *pb.Chat_Id) (*pb.Chat_Nil, error) {
	return nil, nil
}

func (s *server) JoinGroup(context.Context, *pb.Chat_JoinGroup) (*pb.Chat_Nil, error) {
	return nil, nil
}
func (s *server) LeaveGroup(context.Context, *pb.Chat_LeaveGroup) (*pb.Chat_Nil, error) {
	return nil, nil
}

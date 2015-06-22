package main

import (
	"golang.org/x/net/context"
	"sync"
)

import (
	. "proto"
)

const (
	SERVICE = "[CHAT]"
)

const (
	BOLTDB_FILE        = "/data/CHAT.DAT"
	BOLTDB_USER_BUCKET = "USER"
	BOLTDB_MUC_BUCKET  = "MUC"
	MAX_QUEUE_SIZE     = 128 // num of message kept
)

type Message struct {
	From     int32
	To       int32
	FromName string
	Body     string
}

type User struct {
	Name  string
	Inbox []Message
}

type Muc struct {
	UserIds []int32
	Inbox   []Message
}

type server struct {
	Users map[int32]*User
	Mucs  map[int32]*Muc
	sync.RWMutex
}

func (s *server) init() {
	s.Users = make(map[int32]*User)
	s.Mucs = make(map[int32]*Muc)
}

func (s *server) Packet(stream ChatService_PacketServer) error {
	// TODO: for bidirectional stream
	return nil
}

func (s *server) Reg(ctx context.Context, req *Chat_Id) (*Chat_Nil, error) {
	s.Lock()
	defer s.Unlock()
	// TODO: register a user
	return nil, nil
}

func (s *server) UpdateInfo(ctx context.Context, req *Chat_Id) (*Chat_Nil, error) {
	// TODO: update a user
	return nil, nil
}

func (s *server) RegMuc(context.Context, *Chat_MucReq) (*Chat_Nil, error) {
	// TODO: register a muc
	return nil, nil
}

func (s *server) JoinMuc(context.Context, *Chat_MucReq) (*Chat_Nil, error) {
	// TODO: join a muc
	return nil, nil
}

func (s *server) LeaveMuc(context.Context, *Chat_MucReq) (*Chat_Nil, error) {
	// TODO: leave a muc
	return nil, nil
}

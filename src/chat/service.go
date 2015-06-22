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

func (s *server) Subscribe(ChatService_SubscribeServer) error {
	return nil
}

func (s *server) MucSubscribe(ChatService_MucSubscribeServer) error {
	return nil
}

func (s *server) Send(context.Context, *Chat_Message) (*Chat_Nil, error) {
	return nil, nil
}

func (s *server) Reg(ctx context.Context, req *Chat_Id) (*Chat_Nil, error) {
	s.Lock()
	defer s.Unlock()
	return nil, nil
}

func (s *server) RegMuc(context.Context, *Chat_MucReq) (*Chat_Nil, error) {
	return nil, nil
}

func (s *server) JoinMuc(context.Context, *Chat_MucReq) (*Chat_Nil, error) {
	return nil, nil
}

func (s *server) LeaveMuc(context.Context, *Chat_MucReq) (*Chat_Nil, error) {
	return nil, nil
}

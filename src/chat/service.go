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

type User struct {
	Inbox []Chat_Message
}

type Muc struct {
	UserIds []int32
	Inbox   []Chat_Message
}

type server struct {
	users          map[int32]*User
	mucs           map[int32]*Muc
	chat_chan      chan Chat_Message
	chat_chan_ctrl chan Chat_Param
	muc_chan       chan Chat_Message
	muc_chan_ctrl  chan Chat_Param
	sync.RWMutex
}

func (s *server) init() {
	s.users = make(map[int32]*User)
	s.mucs = make(map[int32]*Muc)
}

func (s *server) Subscribe(stream ChatService_SubscribeServer) error {
	die := make(chan bool)
	defer close(die)

	go func() {
		for {
			select {
			case msg <- s.chat_chan:
			case ctrl <- s.chat_chan_ctrl:
			case <-die:
				return
			}
		}
	}()

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		s.chat_chan_ctrl <- in
	}
	return nil
}

func (s *server) MucSubscribe(stream ChatService_MucSubscribeServer) error {
	return nil
}

func (s *server) Send(ctx context.Context, msg *Chat_Message) (*Chat_Nil, error) {
	switch msg.Type {
	case Chat_CHAT:
		if u := s.users[msg.ToId]; u != nil {
			u.Inbox = append(u.Inbox, *msg)
		}
		s.chat_chan <- *msg
	case Chat_MUC:
		if u := s.mucs[msg.ToId]; u != nil {
			u.Inbox = append(u.Inbox, *msg)
		}
		s.muc_chan <- *msg
	}
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

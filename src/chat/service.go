package main

import (
	"errors"
	"golang.org/x/net/context"
	"io"
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

var (
	ERROR_BAD_MESSAGE_TYPE = errors.New("bad message type")
)

type User struct {
	Inbox []Chat_Message
	PS    *PubSub
}

type Muc struct {
	UserIds []int32
	Inbox   []Chat_Message
	PS      *PubSub
}

type server struct {
	users     map[int32]*User
	mucs      map[int32]*Muc
	user_lock sync.RWMutex
	mucs_lock sync.RWMutex
}

func (s *server) read_user(id int32) *User {
	s.user_lock.RLock()
	defer s.user_lock.RUnlock()
	return s.users[id]
}

func (s *server) read_muc(mucid int32) *Muc {
	s.mucs_lock.RLock()
	defer s.mucs_lock.RUnlock()
	return s.mucs[mucid]
}

func (s *server) init() {
	s.users = make(map[int32]*User)
	s.mucs = make(map[int32]*Muc)
}

func (s *server) Subscribe(stream ChatService_SubscribeServer) error {
	f := func(msg *Chat_Message) {
		stream.Send(msg)
	}

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		switch in.PS {
		case Chat_SUBSCRIBE:
			if user := s.read_user(in.Uid); user != nil {
				user.PS.Sub(f)
			}
		case Chat_UNSUBSCRIBE:
			if user := s.read_user(in.Uid); user != nil {
				user.PS.Leave(f)
			}
		}
	}
	return nil
}

func (s *server) MucSubscribe(stream ChatService_MucSubscribeServer) error {
	return nil
}

func (s *server) Send(ctx context.Context, msg *Chat_Message) (*Chat_Nil, error) {
	switch msg.Type {
	case Chat_CHAT:
		if user := s.read_user(msg.ToId); user != nil {
			user.PS.Pub(msg)
		}
	case Chat_MUC:
		if muc := s.read_muc(msg.ToId); muc != nil {
			muc.PS.Pub(msg)
		}
	default:
		return nil, ERROR_BAD_MESSAGE_TYPE
	}
	return nil, nil
}

func (s *server) Reg(ctx context.Context, req *Chat_Id) (*Chat_Nil, error) {
	s.user_lock.Lock()
	defer s.user_lock.Unlock()
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

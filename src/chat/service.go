package main

import (
	"errors"
	log "github.com/GameGophers/libs/nsq-logger"
	"golang.org/x/net/context"
	"io"
	"sync"
)

import (
	. "proto"
	"pubsub"
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
	OK                     = &Chat_Nil{}
	ERROR_BAD_MESSAGE_TYPE = errors.New("bad message type")
	ERROR_ALREADY_EXISTS   = errors.New("id already exists")
	ERROR_NOT_EXISTS       = errors.New("id not exists")
)

type User struct {
	Inbox []Chat_Message
	PS    *pubsub.PubSub
	sync.Mutex
}

func NewUser() *User {
	u := &User{}
	u.PS = pubsub.New()
	return u
}

type Muc struct {
	UserIds []int32
	Inbox   []Chat_Message
	PS      *pubsub.PubSub
	sync.Mutex
}

func NewMuc() *Muc {
	m := &Muc{}
	m.PS = pubsub.New()
	return m
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

	defer func() {
		// TODO : unsubscribe this line
	}()

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		switch in.PS {
		case Chat_SUBSCRIBE:
			if user := s.read_user(in.Uid); user != nil {
				user.PS.Sub(f)
			} else {
				log.Errorf("subscribe to unknown uid:%v", in.Uid)
			}
		case Chat_UNSUBSCRIBE:
			if user := s.read_user(in.Uid); user != nil {
				user.PS.Leave(f)
			} else {
				log.Errorf("un-subscribe to unknown uid:%v", in.Uid)
			}
		}
	}
	return nil
}

func (s *server) MucSubscribe(stream ChatService_MucSubscribeServer) error {
	f := func(msg *Chat_Message) {
		stream.Send(msg)
	}

	defer func() {
		// TODO : unsubscribe this line
	}()

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		switch in.PS {
		case Chat_SUBSCRIBE:
			if muc := s.read_muc(in.MucId); muc != nil {
				muc.PS.Sub(f)
			} else {
				log.Errorf("muc subscribe to unknown mucid:%v", in.MucId)
			}
		case Chat_UNSUBSCRIBE:
			if muc := s.read_muc(in.MucId); muc != nil {
				muc.PS.Leave(f)
			} else {
				log.Errorf("muc un-subscribe to unknown mucid:%v", in.MucId)
			}
		}
	}
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
	return OK, nil
}

func (s *server) Reg(ctx context.Context, req *Chat_Id) (*Chat_Nil, error) {
	s.user_lock.Lock()
	defer s.user_lock.Lock()
	user := s.users[req.Id]
	if user != nil {
		log.Errorf("id already exists:%v", req.Id)
		return nil, ERROR_ALREADY_EXISTS
	}

	s.users[req.Id] = NewUser()
	return OK, nil
}

func (s *server) RegMuc(ctx context.Context, req *Chat_MucReq) (*Chat_Nil, error) {
	s.mucs_lock.Lock()
	defer s.mucs_lock.Lock()
	m := s.mucs[req.MucId]
	if m != nil {
		log.Errorf("mucid already exists:%v", req.MucId)
		return nil, ERROR_ALREADY_EXISTS
	}

	s.mucs[req.MucId] = NewMuc()
	return OK, nil
}

func (s *server) JoinMuc(ctx context.Context, req *Chat_MucReq) (*Chat_Nil, error) {
	s.mucs_lock.Lock()
	defer s.mucs_lock.Lock()
	m := s.mucs[req.MucId]
	if m == nil {
		log.Errorf("mucid not exists:%v", req.MucId)
		return nil, ERROR_NOT_EXISTS
	}

	m.Lock()
	defer m.Unlock()
	for _, id := range m.UserIds {
		if id == req.UserId {
			return nil, ERROR_ALREADY_EXISTS
		}
	}

	m.UserIds = append(m.UserIds, req.UserId)
	return OK, nil
}

func (s *server) LeaveMuc(ctx context.Context, req *Chat_MucReq) (*Chat_Nil, error) {
	s.mucs_lock.Lock()
	defer s.mucs_lock.Lock()
	m := s.mucs[req.MucId]
	if m == nil {
		log.Errorf("mucid not exists:%v", req.MucId)
		return nil, ERROR_NOT_EXISTS
	}

	m.Lock()
	defer m.Unlock()
	for k := range m.UserIds {
		if m.UserIds[k] == req.UserId {
			m.UserIds = append(m.UserIds[:k], m.UserIds[k+1:]...)
			return OK, nil
		}
	}

	return nil, ERROR_NOT_EXISTS
}

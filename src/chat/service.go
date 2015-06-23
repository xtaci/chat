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

type EndPoint struct {
	Inbox []Chat_Message
	ps    *pubsub.PubSub
	sync.Mutex
}

func (ep *EndPoint) Push(msg *Chat_Message) {
	ep.Lock()
	defer ep.Unlock()
	if len(ep.Inbox) > MAX_QUEUE_SIZE {
		ep.Inbox = append(ep.Inbox[1:], *msg)
	} else {
		ep.Inbox = append(ep.Inbox, *msg)
	}
}

func NewEndPoint() *EndPoint {
	u := &EndPoint{}
	u.ps = pubsub.New()
	return u
}

type server struct {
	users     map[int32]*EndPoint
	mucs      map[int32]*EndPoint
	user_lock sync.RWMutex
	mucs_lock sync.RWMutex
}

func (s *server) read_user(id int32) *EndPoint {
	s.user_lock.RLock()
	defer s.user_lock.RUnlock()
	return s.users[id]
}

func (s *server) read_muc(mucid int32) *EndPoint {
	s.mucs_lock.RLock()
	defer s.mucs_lock.RUnlock()
	return s.mucs[mucid]
}

func (s *server) init() {
	s.users = make(map[int32]*EndPoint)
	s.mucs = make(map[int32]*EndPoint)
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

		var ep *EndPoint
		switch in.MT {
		case Chat_CHAT:
			ep = s.read_user(in.Uid)
		case Chat_MUC:
			ep = s.read_muc(in.MucId)
		}
		if ep == nil {
			log.Errorf("cannot find endpoint %v", in)
			continue
		}

		switch in.PS {
		case Chat_SUBSCRIBE:
			ep.ps.Sub(f)
		case Chat_UNSUBSCRIBE:
			ep.ps.Leave(f)
		}
	}
	return nil
}

func (s *server) Send(ctx context.Context, msg *Chat_Message) (*Chat_Nil, error) {
	var ep *EndPoint
	switch msg.Type {
	case Chat_CHAT:
		ep = s.read_user(msg.ToId)
	case Chat_MUC:
		ep = s.read_muc(msg.ToId)
	default:
		return nil, ERROR_BAD_MESSAGE_TYPE
	}

	if ep == nil {
		return nil, ERROR_NOT_EXISTS
	}

	ep.ps.Pub(msg)
	ep.Lock()
	defer ep.Lock()
	ep.Push(msg)
	return OK, nil
}

func (s *server) Reg(ctx context.Context, req *Chat_Id) (*Chat_Nil, error) {
	switch req.MT {
	case Chat_CHAT:
		if err := s.register_chat(req); err != nil {
			return nil, err
		}
		return OK, nil
	case Chat_MUC:
		if err := s.register_muc(req); err != nil {
			return nil, err
		}
		return OK, nil
	default:
		return nil, ERROR_BAD_MESSAGE_TYPE
	}
}

func (s *server) register_chat(req *Chat_Id) error {
	s.user_lock.Lock()
	defer s.user_lock.Lock()
	user := s.users[req.Id]
	if user != nil {
		log.Errorf("id already exists:%v", req.Id)
		return ERROR_ALREADY_EXISTS
	}

	s.users[req.Id] = NewEndPoint()
	return nil
}

func (s *server) register_muc(req *Chat_Id) error {
	s.mucs_lock.Lock()
	defer s.mucs_lock.Lock()
	m := s.mucs[req.Id]
	if m != nil {
		log.Errorf("mucid already exists:%v", req.Id)
		return ERROR_ALREADY_EXISTS
	}

	s.mucs[req.Id] = NewEndPoint()
	return nil
}

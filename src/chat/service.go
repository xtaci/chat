package main

import (
	"errors"
	log "github.com/GameGophers/libs/nsq-logger"
	"golang.org/x/net/context"
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

func (ep *EndPoint) Read() []Chat_Message {
	ep.Lock()
	defer ep.Unlock()
	ret := make([]Chat_Message, len(ep.Inbox))
	copy(ret, ep.Inbox)
	return ret
}

func NewEndPoint() *EndPoint {
	u := &EndPoint{}
	u.ps = pubsub.New()
	return u
}

type server struct {
	eps map[uint64]*EndPoint
	sync.RWMutex
}

func (s *server) read_ep(id uint64) *EndPoint {
	s.RLock()
	defer s.RUnlock()
	return s.eps[id]
}

func (s *server) init() {
	s.eps = make(map[uint64]*EndPoint)
}

func (s *server) Subscribe(p *Chat_Id, stream ChatService_SubscribeServer) error {
	die := make(chan bool)
	f := func(msg *Chat_Message) {
		if err := stream.Send(msg); err != nil {
			close(die)
		}
	}

	ep := s.read_ep(p.Id)
	if ep == nil {
		log.Errorf("cannot find endpoint %v", p)
		return ERROR_NOT_EXISTS
	}

	ep.ps.Sub(f)
	defer func() {
		ep.ps.Leave(f)
	}()

	<-die
	return nil
}

func (s *server) Read(p *Chat_Id, stream ChatService_ReadServer) error {
	ep := s.read_ep(p.Id)
	if ep == nil {
		log.Errorf("cannot find endpoint %v", p)
		return ERROR_NOT_EXISTS
	}

	msgs := ep.Read()
	for k := range msgs {
		if err := stream.Send(&msgs[k]); err != nil {
			return err
		}
	}
	return nil
}

func (s *server) Send(ctx context.Context, msg *Chat_Message) (*Chat_Nil, error) {
	ep := s.read_ep(msg.Dst)
	if ep == nil {
		return nil, ERROR_NOT_EXISTS
	}

	ep.ps.Pub(msg)
	ep.Lock()
	defer ep.Lock()
	ep.Push(msg)
	return OK, nil
}

func (s *server) Reg(ctx context.Context, p *Chat_Id) (*Chat_Nil, error) {
	s.Lock()
	defer s.Lock()
	ep := s.eps[p.Id]
	if ep != nil {
		log.Errorf("id already exists:%v", p.Id)
		return nil, ERROR_ALREADY_EXISTS
	}

	s.eps[p.Id] = NewEndPoint()
	return OK, nil
}

package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/boltdb/bolt"
	log "github.com/gonet2/libs/nsq-logger"
	"golang.org/x/net/context"
	"gopkg.in/vmihailenco/msgpack.v2"
)

import (
	. "proto"
)

const (
	SERVICE = "[CHAT]"
)

const (
	BOLTDB_FILE    = "/data/CHAT.DAT"
	BOLTDB_BUCKET  = "EPS"
	MAX_QUEUE_SIZE = 128 // num of message kept
	PENDING_SIZE   = 65536
	CHECK_INTERVAL = time.Minute
)

var (
	OK                   = &Chat_Nil{}
	ERROR_ALREADY_EXISTS = errors.New("id already exists")
	ERROR_NOT_EXISTS     = errors.New("id not exists")
)

// endpoint definition
type EndPoint struct {
	inbox []Chat_Message
	ps    *PubSub
	sync.Mutex
}

// push a message to this endpoint
func (ep *EndPoint) Push(msg *Chat_Message) {
	ep.Lock()
	defer ep.Unlock()
	if len(ep.inbox) > MAX_QUEUE_SIZE {
		ep.inbox = append(ep.inbox[1:], *msg)
	} else {
		ep.inbox = append(ep.inbox, *msg)
	}
}

// read all messages from this endpoint
func (ep *EndPoint) Read() []Chat_Message {
	ep.Lock()
	defer ep.Unlock()
	return append([]Chat_Message(nil), ep.inbox...)
}

func NewEndPoint() *EndPoint {
	u := &EndPoint{}
	u.ps = &PubSub{}
	u.ps.init()
	return u
}

// server definition
type server struct {
	eps     map[uint64]*EndPoint // end-point-s
	pending chan uint64          // dirty id pendings
	sync.RWMutex
}

func (s *server) init() {
	s.eps = make(map[uint64]*EndPoint)
	s.pending = make(chan uint64, PENDING_SIZE)
	s.restore()
	go s.persistence_task()
}

func (s *server) read_ep(id uint64) *EndPoint {
	s.RLock()
	defer s.RUnlock()
	return s.eps[id]
}

// subscribe to an endpoint & receive server streams
func (s *server) Subscribe(p *Chat_Id, stream ChatService_SubscribeServer) error {
	// read endpoint
	ep := s.read_ep(p.Id)
	if ep == nil {
		log.Errorf("cannot find endpoint %v", p)
		return ERROR_NOT_EXISTS
	}

	// send history chat messages
	msgs := ep.Read()
	for k := range msgs {
		if err := stream.Send(&msgs[k]); err != nil {
			return nil
		}
	}

	// create wrapper
	die := make(chan bool)
	f := NewSubscriber(func(msg *Chat_Message) {
		if err := stream.Send(msg); err != nil {
			close(die)
		}
	})
	log.Tracef("new subscriber: %p", f)

	// subscribe to the endpoint
	ep.ps.Sub(f)
	defer func() {
		ep.ps.Leave(f)
	}()

	<-die
	return nil
}

func (s *server) Send(ctx context.Context, msg *Chat_Message) (*Chat_Nil, error) {
	ep := s.read_ep(msg.Id)
	if ep == nil {
		return nil, ERROR_NOT_EXISTS
	}

	ep.ps.Pub(msg)
	ep.Push(msg)
	s.pending <- msg.Id
	return OK, nil
}

func (s *server) Reg(ctx context.Context, p *Chat_Id) (*Chat_Nil, error) {
	s.Lock()
	defer s.Unlock()
	ep := s.eps[p.Id]
	if ep != nil {
		log.Errorf("id already exists:%v", p.Id)
		return nil, ERROR_ALREADY_EXISTS
	}

	s.eps[p.Id] = NewEndPoint()
	s.pending <- p.Id
	return OK, nil
}

// persistence endpoints into db
func (s *server) persistence_task() {
	timer := time.After(CHECK_INTERVAL)
	db := s.open_db()
	changes := make(map[uint64]bool)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	for {
		select {
		case key := <-s.pending:
			changes[key] = true
		case <-timer:
			s.dump(db, changes)
			log.Infof("perisisted %v endpoints:", len(changes))
			changes = make(map[uint64]bool)
			timer = time.After(CHECK_INTERVAL)
		case nr := <-sig:
			s.dump(db, changes)
			db.Close()
			log.Info(nr)
			os.Exit(0)
		}
	}
}

func (s *server) open_db() *bolt.DB {
	db, err := bolt.Open(BOLTDB_FILE, 0600, nil)
	if err != nil {
		log.Critical(err)
		os.Exit(-1)
	}
	// create bulket
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(BOLTDB_BUCKET))
		if err != nil {
			log.Criticalf("create bucket: %s", err)
			os.Exit(-1)
		}
		return nil
	})
	return db
}

func (s *server) dump(db *bolt.DB, changes map[uint64]bool) {
	for k := range changes {
		ep := s.read_ep(k)
		if ep == nil {
			log.Errorf("cannot find endpoint %v", k)
			continue
		}

		// serialization and save
		bin, err := msgpack.Marshal(ep.Read())
		if err != nil {
			log.Critical("cannot marshal:", err)
			continue
		}

		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(BOLTDB_BUCKET))
			err := b.Put([]byte(fmt.Sprint(k)), bin)
			return err
		})
	}
}

func (s *server) restore() {
	// restore data from db file
	db := s.open_db()
	defer db.Close()
	count := 0
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BOLTDB_BUCKET))
		b.ForEach(func(k, v []byte) error {
			var msg []Chat_Message
			err := msgpack.Unmarshal(v, &msg)
			if err != nil {
				log.Critical("chat data corrupted:", err)
				os.Exit(-1)
			}
			id, err := strconv.ParseUint(string(k), 0, 64)
			if err != nil {
				log.Critical("chat data corrupted:", err)
				os.Exit(-1)
			}
			ep := NewEndPoint()
			ep.inbox = msg
			s.eps[id] = ep
			count++
			return nil
		})
		return nil
	})

	log.Infof("restored %v chats", count)
}

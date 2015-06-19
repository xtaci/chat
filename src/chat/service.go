package main

import (
	"errors"
	"strings"

	"golang.org/x/net/context"
)

import (
	pb "proto"
)

const (
	SERVICE = "[CHAT]"
)

const (
	BOLTDB_FILE   = "/data/CHAT-DUMP.DAT"
	BOLTDB_BUCKET = "CHAT"
	QUEUE_SIZE    = 1024
)

var (
	ERROR_CHANNEL_NOT_EXISTS = errors.New("channel not exists")
	_channel_map             = map[string]string{
		"global": "global",
		"room":   "room",
	}
)

type server struct {
	queue chan *pb.Chat_Msg
}

func (s *server) init() {
	s.queue = make(chan *pb.Chat_Msg, QUEUE_SIZE)
}

//---------------------------------------------- listen the chat room
func (s *server) Query(in *pb.Chat_Channel, stream pb.ChatService_QueryServer) error {

	for {
		//chat_msg := <-s.queue
		//stream.Send(chat_msg)
	}
}

//----------------------------------------------- send a message
func (s *server) Send(ctx context.Context, in *pb.Chat_Msg) (*pb.Chat_NullResult, error) {
	c := strings.Split(in.Channel, ":")
	ch, _ := c[0], c[1]
	if _, ok := _channel_map[ch]; !ok {
		return nil, ERROR_CHANNEL_NOT_EXISTS
	}
	s.queue <- in
	return &pb.Chat_NullResult{}, nil
}

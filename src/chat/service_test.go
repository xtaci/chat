package main

import (
	"log"
	. "proto"
	"testing"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	address = "localhost:50008"
)

func TestChat(t *testing.T) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address)
	if err != nil {
		t.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := NewChatServiceClient(conn)

	// Contact the server and print out its response.
	_, err = c.Reg(context.Background(), &Chat_Id{Id: 1})
	if err != nil {
		log.Printf("could not query: %v", err)
	}

	go send(&Chat_Message{Id: 1, Body: []byte("Hello")})
	go recv(&Chat_Id{1})
}

func send(m *Chat_Message) {
	conn, err := grpc.Dial(address)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := NewChatServiceClient(conn)
	for {
		_, err := c.Send(context.Background(), m)
		if err != nil {
			log.Printf("send err : %v", err)
		}
		log.Printf("send msg: %v", m)
		time.Sleep(3 * time.Second)
	}
}

func recv(chat_id *Chat_Id) {
	conn, err := grpc.Dial(address)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := NewChatServiceClient(conn)
	stream, err := c.Subscribe(context.Background(), chat_id)
	if err != nil {
		log.Fatal(err)
	}
	for {
		message, err := stream.Recv()
		if err != nil {
			log.Printf("recv err :%v", err)
		}
		log.Print(message)
	}
	log.Printf("recv msg end")
}

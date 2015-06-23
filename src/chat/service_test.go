package main

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	. "proto"
	"testing"
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
		t.Logf("could not query: %v", err)
	}

	stream, err := c.Subscribe(context.Background(), &Chat_Id{Id: 1})
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		_, err = c.Send(context.Background(), &Chat_Message{Id: 1, Body: []byte("hello world")})
		t.Log("sent")
		if err != nil {
			t.Fatal(err)
		}
	}()

	message, err := stream.Recv()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(message)
}

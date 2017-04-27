package main

import (
	"chat/kafka"
	pb "chat/proto"
	"fmt"
	"testing"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	address = "127.0.0.1:10000"
)

var (
	conn *grpc.ClientConn
	err  error
)

func init() {
	addrs := []string{"127.0.0.1:9092"}
	ChatTopic := "chat_updates"
	kafka.InitTest(addrs, ChatTopic)
}

func TestChat(t *testing.T) {
	// Set up a connection to the server.
	conn, err = grpc.Dial(address, grpc.WithInsecure(), grpc.WithTimeout(5*time.Second), grpc.WithBlock())
	if err != nil {
		t.Fatalf("did not connect: %v", err)
	}
	c := pb.NewChatServiceClient(conn)

	// Contact the server and print out its response.
	_, err = c.Reg(context.Background(), &pb.Chat_Id{Id: 1})
	if err != nil {
		t.Logf("could not query: %v", err)
	}
	done := make(chan bool, 2)
	const COUNT = 10
	go recv(&pb.Chat_Consumer{Id: 1, From: -1}, COUNT, done, t)
	go recv(&pb.Chat_Consumer{Id: 1, From: -1}, COUNT, done, t)
	time.Sleep(3 * time.Second)
	send(1, []byte("Hello"), COUNT, t)
	<-done
	<-done
	fmt.Println("finish")
}

func TestUnReg(t *testing.T) {
	c := pb.NewChatServiceClient(conn)
	_, err := c.UnReg(context.Background(), &pb.Chat_Id{Id: 1})
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan bool, 1)
	recv(&pb.Chat_Consumer{Id: 1, From: -1}, 1, done, t)
	<-done
}

//send message to kafka topic.
func send(ep uint64, msg []byte, count int, t *testing.T) {
	for i := 0; i < count; i++ {
		kafka.SendChat(ep, msg)
		fmt.Printf("send: ep:%v, msg:%v\n", ep, string(msg))
	}
}

func recv(chat_id *pb.Chat_Consumer, count int, done chan bool, t *testing.T) {
	c := pb.NewChatServiceClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, err := c.Subscribe(ctx, chat_id)
	if err != nil {
		t.Fatal(err)
		return
	}

	for i := 0; i < count; i++ {
		message, err := stream.Recv()
		if err != nil {
			t.Fatal(err)
			break
		}
		fmt.Println("recv:", count, message.Id, string(message.Body), message.Offset)
	}
	done <- true
}

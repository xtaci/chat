package main

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "proto"
	"testing"
)

const (
	address = "localhost:50001"
)

func TestRankChange(t *testing.T) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address)
	if err != nil {
		t.Fatal("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewRankingServiceClient(conn)

	// Contact the server and print out its response.
	_, err = c.RankChange(context.Background(), &pb.Ranking_Change{1, 100, "testkey"})
	if err != nil {
		t.Fatalf("could not query: %v", err)
	}
}

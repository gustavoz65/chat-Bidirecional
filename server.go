package main

import (
	"bidirecional/chat"
	"fmt"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
)

type chatServer struct {
	chat.UnimplementedChatServiceServer
	mu       sync.Mutex
	clients  map[chat.ChatService_JoinServer]bool
	messages chan *chat.Message
}

func newChatServer() *chatServer {
	return &chatServer{
		clients:  make(map[chat.ChatService_JoinServer]bool),
		messages: make(chan *chat.Message),
	}
}

func (s *chatServer) Join(stream chat.ChatService_JoinServer) error {

	s.mu.Lock()
	s.clients[stream] = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, stream)
		s.mu.Unlock()
	}()

	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				return
			}
			s.messages <- msg
		}
	}()

	for messages := range s.messages {
		s.mu.Lock()
		for client := range s.clients {
			if err := client.Send(messages); err != nil {
				fmt.Printf("Error sending message to client: %v\n", err)
			}
		}
		s.mu.Unlock()
	}

	return nil
}

func main() {
	listerner, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	server := grpc.NewServer()
	chat.RegisterChatServiceServer(server, newChatServer())

	if err := server.Serve(listerner); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}

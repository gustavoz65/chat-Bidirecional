package main

import (
	"bidirecional/chat"
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
)

func main() {
	connection, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	defer connection.Close()

	chatClient := chat.NewChatServiceClient(connection)

	stream, err := chatClient.Join(context.Background())

	if err != nil {
		log.Fatalf("Error joining chat: %v", err)
	}

	fmt.Print("Digite seu nome: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	user := scanner.Text()

	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				log.Fatalf("Error receiving message: %v", err)
			}

			fmt.Println(fmt.Sprintf("[%s] %s: %s", time.Unix(msg.Timestamp, 0).Format("15:04"), msg.User, msg.Text))
		}
	}()

	for scanner.Scan() {
		msg := &chat.Message{
			User:      user,
			Text:      scanner.Text(),
			Timestamp: time.Now().Unix(),
		}
		if err := stream.Send(msg); err != nil {
			log.Fatalf("Error sending message: %v", err)
		}
	}
}

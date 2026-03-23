package main

import (
	"context"
	"fmt"
	"log"
	pb "tasksmgr/gen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient(":9091", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	client := pb.NewNotesServiceClient(conn)

	resp, err := client.CreateNote(context.Background(), &pb.CreateNoteRequest{Title: "Тестовая заметка"})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Create note with ID: %d", resp.Id)

	note, err := client.GetNoteById(context.Background(), &pb.GetNoteByIdRequest{Id: 1})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Get note with ID: %d", note.Id)

}

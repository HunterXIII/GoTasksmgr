package main

import (
	"context"
	"fmt"
	"io"
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

	_, err = client.CreateNote(context.Background(), &pb.CreateNoteRequest{Title: "Тестовая заметка"})
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.CreateNote(context.Background(), &pb.CreateNoteRequest{Title: "Тестовая заметка №2"})
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println("Notes were created")

	note, err := client.GetNoteById(context.Background(), &pb.GetNoteByIdRequest{Id: 1})
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Printf("Get note: %d - %s\n", note.Id, note.Title)

	stream, err := client.GetListNotes(context.Background(), &pb.ListNoteRequest{})
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println("START OF LIST")
	for {
		note, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("END OF LIST")
			break
		}
		fmt.Printf("%d - %s\n", note.Id, note.Title)
	}

}

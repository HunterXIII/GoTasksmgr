package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	pb "tasksmgr/gen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type AuthResponse struct {
	Token string `json:"token"`
}

func GetToken() string {
	url := "http://localhost:8080/auth/login"

	// тело запроса (если нужен POST)
	payload := map[string]string{
		"username": "admin",
		"password": "12345678",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return ""
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return ""
	}

	var authResponse AuthResponse

	err = json.NewDecoder(resp.Body).Decode(&authResponse)
	if err != nil {
		return ""
	}

	return authResponse.Token

}

func main() {

	conn, err := grpc.NewClient(":9091", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	client := pb.NewNotesServiceClient(conn)

	token := GetToken()
	ctx := context.Background()
	ctx = metadata.AppendToOutgoingContext(
		ctx,
		"authorization",
		"Bearer "+token,
	)

	_, err = client.CreateNote(ctx, &pb.CreateNoteRequest{Title: "Тестовая заметка"})
	if err != nil {
		log.Fatal(err)
		return
	}

	_, err = client.CreateNote(ctx, &pb.CreateNoteRequest{Title: "Тестовая заметка №2"})
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println("Notes were created")

	note, err := client.GetNoteById(ctx, &pb.GetNoteByIdRequest{Id: 1})
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Printf("Get note: %d - %s\n", note.Id, note.Title)

	stream, err := client.GetListNotes(ctx, &pb.ListNoteRequest{})
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
		fmt.Printf("%d - %s (User: %d)\n", note.Id, note.Title, note.UserId)
	}

}

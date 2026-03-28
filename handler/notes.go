package handler

import (
	"context"
	"fmt"
	notesmgrv1 "tasksmgr/gen"
	"tasksmgr/repo"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NotesHandler struct {
	notesmgrv1.UnimplementedNotesServiceServer
	repo *repo.NotesRepository
}

func NewNotesHandler(r *repo.NotesRepository) *NotesHandler {
	return &NotesHandler{repo: r}
}

func (h *NotesHandler) CreateNote(ctx context.Context, note *notesmgrv1.CreateNoteRequest) (*notesmgrv1.CreateNoteResponse, error) {
	id, err := h.repo.CreateNote(ctx, note.Title)
	if err != nil {
		fmt.Printf("Note didn't create: %s\n", err)
		return nil, status.Error(codes.Internal, "Note didn't create")
	}

	return &notesmgrv1.CreateNoteResponse{Id: int64(id)}, nil
}

func (h *NotesHandler) GetNoteById(ctx context.Context, getNote *notesmgrv1.GetNoteByIdRequest) (*notesmgrv1.Note, error) {
	note, err := h.repo.GetNoteById(ctx, int(getNote.Id))
	if err != nil {
		fmt.Printf("Note not found: %s\n", err)
		return nil, status.Error(codes.NotFound, "Note not found")
	}

	return &notesmgrv1.Note{
		Id:     int64(note.Id),
		Title:  note.Title,
		UserId: int64(note.UserID),
	}, nil
}

func (h *NotesHandler) GetListNotes(r *notesmgrv1.ListNoteRequest, stream grpc.ServerStreamingServer[notesmgrv1.Note]) error {
	notes, err := h.repo.GetNotes(context.Background())
	if err != nil {
		fmt.Printf("Note didn't get: %s\n", err)
		return status.Error(codes.Internal, "Notes didn't get")
	}

	for _, note := range notes {
		if err := stream.Send(&notesmgrv1.Note{Id: int64(note.Id), Title: note.Title, UserId: int64(note.UserID)}); err != nil {
			return err
		}
		time.Sleep(2 * time.Second)
	}
	return nil
}

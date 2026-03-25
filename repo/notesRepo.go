package repo

import (
	"context"
	"fmt"
)

type Note struct {
	Id     int    `json:"id"`
	Title  string `json:"title"`
	UserID int    `json:"user_id"`
}

type NotesRepository struct {
	notes  map[int]Note
	nextID int
}

func NewNotesRepository() *NotesRepository {
	return &NotesRepository{
		notes:  make(map[int]Note),
		nextID: 1,
	}
}

func (r *NotesRepository) CreateNote(ctx context.Context, title string) (int, error) {
	r.notes[r.nextID] = Note{Id: r.nextID, Title: title}
	r.nextID++
	return r.nextID - 1, nil
}

func (r *NotesRepository) GetNoteById(ctx context.Context, id int) (*Note, error) {
	note, ok := r.notes[id]
	if !ok {
		return nil, fmt.Errorf("note with id=%d not found", id)
	}

	return &note, nil
}

func (r *NotesRepository) GetNotes(ctx context.Context) ([]Note, error) {
	result := []Note{}
	for _, note := range r.notes {
		result = append(result, note)
	}
	return result, nil
}

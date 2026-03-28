package repo

import (
	"context"
	"database/sql"
	"fmt"
	"tasksmgr/contextx"
)

type Note struct {
	Id     int    `json:"id"`
	Title  string `json:"title"`
	UserID int    `json:"user_id"`
}

type NotesRepository struct {
	db *sql.DB
}

func NewNotesRepository(db *sql.DB) *NotesRepository {
	return &NotesRepository{
		db: db,
	}
}

func (r *NotesRepository) CreateNote(ctx context.Context, title string) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	defer tx.Rollback()
	var id int
	err = tx.QueryRowContext(ctx, "INSERT INTO notes (title, user_id) VALUES ($1, $2) RETURNING id", title, ctx.Value(contextx.UserIDKey{}).(int)).Scan(&id)
	if err != nil {
		return 0, err
	}
	tx.Commit()
	return id, nil
}

func (r *NotesRepository) GetNoteById(ctx context.Context, id int) (*Note, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	defer tx.Rollback()
	var note Note
	err = tx.QueryRowContext(ctx, "SELECT id, title, user_id FROM notes WHERE id = $1", id).Scan(&note.Id, &note.Title, &note.UserID)
	if err != nil {
		return nil, err
	}
	tx.Commit()
	return &note, nil
}

func (r *NotesRepository) GetNotes(ctx context.Context) ([]Note, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, "SELECT id, title, user_id FROM notes")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer rows.Close()

	notes := []Note{}

	for rows.Next() {
		var note Note
		if err := rows.Scan(&note.Id, &note.Title, &note.UserID); err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}

	tx.Commit()
	return notes, nil

}

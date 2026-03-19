package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

type Task struct {
	Id     int    `json:"id"`
	Title  string `json:"title"`
	Done   bool   `json:"done"`
	UserID int    `json:"user_id"`
}

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, newTask Task) (int, error) {

	tx, err := r.db.BeginTx(ctx, nil)
	defer tx.Rollback()
	var id int
	err = tx.QueryRowContext(ctx, "INSERT INTO tasks (title, done, user_id) VALUES ($1, $2, $3) RETURNING id", newTask.Title, newTask.Done, newTask.UserID).Scan(&id)
	if err != nil {
		return 0, err
	}

	var lastId int
	tx.QueryRowContext(ctx, "SELECT max(id) FROM tasks").Scan(&lastId)
	newTask.Id = lastId

	jsonValue, err := json.Marshal(newTask)
	if err != nil {
		return 0, err
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO logs (author_id, action, old_value, new_value) VALUES ($1, 'tasks:insert', '{}', $2)", newTask.UserID, jsonValue)
	if err != nil {

		return 0, err
	}

	tx.Commit()
	return id, nil
}

func (r *TaskRepository) GetById(ctx context.Context, id int) (*Task, error) {
	var task Task
	err := r.db.QueryRowContext(ctx, "SELECT id, title, done, user_id FROM tasks WHERE id = $1", id).Scan(&task.Id, &task.Title, &task.Done, &task.UserID)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

/*
func (r *TaskRepository) GoGetById(result chan *Task, errs chan error, ctx context.Context, id int) (*Task, error) {
	var task Task
	err := r.db.QueryRowContext(ctx, "SELECT id, title, done, user_id FROM tasks WHERE id = $1", id).Scan(&task.Id, &task.Title, &task.Done, &task.UserID)
	if err != nil {
		errs <- err
	}
	result <- &task
}
*/

func (r *TaskRepository) List(ctx context.Context, limit *int, offset *int) ([]Task, error) {
	query := "SELECT id, title, done, user_id FROM tasks"
	args := []any{}
	argID := 1

	if limit != nil {
		query += fmt.Sprintf(" LIMIT $%d", argID)
		args = append(args, *limit)
		argID++
	}

	if offset != nil {
		query += fmt.Sprintf(" OFFSET $%d", argID)
		args = append(args, *offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []Task{}

	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.Id, &task.Title, &task.Done, &task.UserID); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *TaskRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = $1", id)
	return err
}

func (r *TaskRepository) Update(ctx context.Context, updateTask Task) error {

	tx, err := r.db.BeginTx(ctx, nil)

	_, err = tx.ExecContext(ctx, "UPDATE tasks SET title = $1, done = $2 WHERE id = $3", updateTask.Title, updateTask.Done, updateTask.Id)
	if err != nil {
		tx.Rollback()
		return err
	}

	var oldTask Task
	err = tx.QueryRowContext(ctx, "SELECT id, title, done, user_id FROM tasks WHERE id = $1", updateTask.Id).Scan(&oldTask.Id, &oldTask.Title, &oldTask.Done, &oldTask.UserID)
	if err != nil {
		tx.Rollback()
		return err
	}

	jsonOldValue, err := json.Marshal(oldTask)
	if err != nil {
		tx.Rollback()
		return err
	}

	jsonNewValue, err := json.Marshal(updateTask)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO logs (author_id, action, old_value, new_value) VALUES ($1, 'tasks:update', $2, $3)", oldTask.UserID, jsonOldValue, jsonNewValue)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil

}

func (r *TaskRepository) UpdateWordCount(ctx context.Context, id int, count int) error {

	_, err := r.db.ExecContext(ctx, "UPDATE tasks SET word_count=$1 WHERE id=$2", count, id)

	return err
}

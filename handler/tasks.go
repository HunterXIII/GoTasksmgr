package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"tasksmgr/repo"
)

type TaskHandler struct {
	repo  *repo.TaskRepository
	queue chan int
}

func NewTaskHandler(repo *repo.TaskRepository, queue chan int) *TaskHandler {
	return &TaskHandler{repo: repo, queue: queue}
}

func (h *TaskHandler) CreateTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var task repo.Task
		err := json.NewDecoder(r.Body).Decode(&task)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		id, err := h.repo.Create(ctx, task)

		select {
		case h.queue <- id:
		default:
			h.repo.Delete(ctx, id)
			w.WriteHeader(http.StatusServiceUnavailable)
			res, _ := json.Marshal("Index queue full")
			w.Write(res)
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Printf("Failed to create a task: %s", err)
			res, _ := json.Marshal("Failed to create a task: ")
			w.Write(res)
		}
	}
}

func (h *TaskHandler) GetById() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Printf("Uncorrect Id: %s\n", err)
			res, _ := json.Marshal("Uncorrect ID")
			w.Write(res)
			return
		}

		// task, err := h.repo.GetById(r.Context(), id)
		// db_ch := make(chan *repo.Task)
		// db_err_ch := make(chan error)

		// select {
		//	case
		//	case
		// }

		task, err := h.repo.GetById(r.Context(), id)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Printf("Task is not found: %s\n", err)
			res, _ := json.Marshal("Task is not found")
			w.Write(res)
			return
		}

		/*
			something := make(chan string)
			go func() {
				fmt.Println("Start goroutine..")
				time.Sleep(3 * time.Second)
				something <- "something from goroutine"
			}()

			resJson := struct {
				Id        int    `json:"id"`
				Title     string `json:"title"`
				Done      bool   `json:"done"`
				UserID    int    `json:"user_id"`
				Something string `json:"somethind"`
			}{
				Id:        task.Id,
				Title:     task.Title,
				Done:      task.Done,
				UserID:    task.UserID,
				Something: <-something,
			}
		*/
		res, _ := json.Marshal(task)
		w.Write(res)
	}
}

func (h *TaskHandler) GetList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// time.Sleep(3 * time.Second)
		tasks, err := h.repo.List(r.Context(), nil, nil)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Printf("Error to get a list of tasks: %s\n", err)
			return
		}
		res, _ := json.Marshal(tasks)
		w.Write(res)
	}
}

func (h *TaskHandler) DeleteTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Printf("Uncorrect Id: %s\n", err)
			res, _ := json.Marshal("Uncorrect ID")
			w.Write(res)
			return
		}

		userID, ok := r.Context().Value("UserID").(int)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			res, _ := json.Marshal("No authorize")
			w.Write(res)
			return
		}

		task, err := h.repo.GetById(r.Context(), id)

		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			res, _ := json.Marshal("Task is not found")
			w.Write(res)
			return
		}

		if userID != task.UserID {
			w.WriteHeader(http.StatusForbidden)
			res, _ := json.Marshal("Access denied")
			w.Write(res)
			return
		}

		err = h.repo.Delete(r.Context(), id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Printf("Failed to delete a task: %s", err)
			res, _ := json.Marshal("Failed to delete a task")
			w.Write(res)
		}

	}
}

func (h *TaskHandler) UpdateTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var task repo.Task

		id, err := strconv.Atoi(r.PathValue("id"))

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Printf("Uncorrect Id: %s\n", err)
			res, _ := json.Marshal("Uncorrect ID")
			w.Write(res)
			return
		}

		task.Id = id

		err = json.NewDecoder(r.Body).Decode(&task)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		userID, ok := r.Context().Value("UserID").(int)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			res, _ := json.Marshal("No authorize")
			w.Write(res)
			return
		}

		if userID != task.UserID {
			w.WriteHeader(http.StatusForbidden)
			res, _ := json.Marshal("Access denied")
			w.Write(res)
			return
		}

		ctx := r.Context()

		err = h.repo.Update(ctx, task)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Printf("Failed to update a task: %s", err)
			res, _ := json.Marshal("Failed to update a task")
			w.Write(res)
		}
	}
}

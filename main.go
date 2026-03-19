package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"tasksmgr/handler"
	"tasksmgr/indexer"
	"tasksmgr/middleware"
	"tasksmgr/repo"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type User struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// func validateLenOfName(name string) bool {
// 	if len([]rune(name)) < 3 {
// 		return false
// 	}
// 	return true
// }

func main() {

	db, err := sql.Open("pgx", "postgres://postgres:postgres@localhost:5433/go")
	if err != nil {
		fmt.Printf("Failed to connect: %s\n", err)
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Printf("Failed to ping: %s\n", err)
		return
	}

	fmt.Printf("Succesfuly connected to database\n")

	mux := http.NewServeMux()

	// AUTH
	usersRepo := repo.NewUserRepository()
	authHandler := handler.NewAuthHandler(usersRepo)
	mux.HandleFunc("POST /auth/login", authHandler.Login())
	mux.Handle("GET /auth/whoami", authHandler.AuthMiddleware(http.HandlerFunc(authHandler.WhoAmI)))
	mux.Handle("POST /auth/refresh", authHandler.AuthMiddleware(http.HandlerFunc(authHandler.Refresh)))

	// TASKS
	queue := make(chan int, 3)
	taskRepo := repo.NewTaskRepository(db)
	taskHandler := handler.NewTaskHandler(taskRepo, queue)

	worker := indexer.NewWorker(queue, taskRepo)
	ctx := context.Background()
	go worker.Start(ctx)

	mux.HandleFunc("POST /tasks", taskHandler.CreateTask())
	mux.Handle("GET /tasks", middleware.TimeoutMiddleware((taskHandler.GetList())))
	mux.HandleFunc("GET /tasks/{id}", taskHandler.GetById())
	mux.Handle("DELETE /tasks/{id}", authHandler.AuthMiddleware(taskHandler.DeleteTask()))
	mux.Handle("PUT /tasks/{id}", authHandler.AuthMiddleware(taskHandler.UpdateTask()))

	// USERS
	// users := make(map[int]User)
	// users[1] = User{Id: 1, Name: "Roman", Age: 19}
	// nextId := 2

	// mux.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Content-Type", "application/json")
	// 	var newUser User

	// 	json.NewDecoder(r.Body).Decode(&newUser)

	// 	if !validateLenOfName(newUser.Name) {
	// 		w.WriteHeader(http.StatusBadRequest)
	// 		res, _ := json.Marshal("Len of name must be more than 2 symbols")
	// 		w.Write(res)
	// 		return
	// 	}

	// 	users[nextId] = User{Id: nextId, Name: newUser.Name, Age: newUser.Age}
	// 	res, _ := json.Marshal(users[nextId])
	// 	nextId++
	// 	w.WriteHeader(http.StatusCreated)
	// 	w.Write(res)
	// })

	// mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Content-Type", "application/json")
	// 	idStr := r.PathValue("id")
	// 	id, err := strconv.Atoi(idStr)
	// 	if err != nil {
	// 		w.WriteHeader(http.StatusBadRequest)
	// 		res, _ := json.Marshal("Uncorrect Id")
	// 		w.Write(res)
	// 		return
	// 	}

	// 	user, ok := users[id]
	// 	if !ok {
	// 		w.WriteHeader(http.StatusNotFound)
	// 		res, _ := json.Marshal("User is not found")
	// 		w.Write(res)
	// 		return
	// 	}
	// 	res, _ := json.Marshal(user)
	// 	w.Write(res)
	// })

	// mux.HandleFunc("GET /users", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Content-Type", "application/json")
	// 	userList := []User{}
	// 	for _, user := range users {
	// 		userList = append(userList, user)
	// 	}
	// 	res, _ := json.Marshal(userList)
	// 	w.Write(res)
	// })

	// mux.HandleFunc("PUT /users/{id}", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Content-Type", "application/json")
	// 	idStr := r.PathValue("id")
	// 	id, err := strconv.Atoi(idStr)

	// 	if err != nil {
	// 		w.WriteHeader(http.StatusBadRequest)
	// 		res, _ := json.Marshal("Uncorrect Id")
	// 		w.Write(res)
	// 		return
	// 	}

	// 	_, ok := users[id]
	// 	if !ok {
	// 		w.WriteHeader(http.StatusNotFound)
	// 		res, _ := json.Marshal("User is not found")
	// 		w.Write(res)
	// 		return
	// 	}

	// 	var updateUser User
	// 	json.NewDecoder(r.Body).Decode(&updateUser)
	// 	updateUser.Id = id
	// 	if !validateLenOfName(updateUser.Name) {
	// 		w.WriteHeader(http.StatusBadRequest)
	// 		res, _ := json.Marshal("Len of name must be more than 2 symbols")
	// 		w.Write(res)
	// 		return
	// 	}
	// 	users[id] = updateUser
	// 	res, _ := json.Marshal(users[id])
	// 	w.Write(res)

	// })

	// mux.HandleFunc("PATCH /users/{id}", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Content-Type", "application/json")
	// 	idStr := r.PathValue("id")
	// 	id, err := strconv.Atoi(idStr)

	// 	if err != nil {
	// 		w.WriteHeader(http.StatusBadRequest)
	// 		res, _ := json.Marshal("Uncorrect Id")
	// 		w.Write(res)
	// 		return
	// 	}

	// 	_, ok := users[id]
	// 	if !ok {
	// 		w.WriteHeader(http.StatusNotFound)
	// 		res, _ := json.Marshal("User is not found")
	// 		w.Write(res)
	// 		return
	// 	}

	// 	var updateUser User
	// 	json.NewDecoder(r.Body).Decode(&updateUser)
	// 	updateUser.Id = id
	// 	user := users[id]
	// 	if updateUser.Name != "" {
	// 		if !validateLenOfName(updateUser.Name) {
	// 			w.WriteHeader(http.StatusBadRequest)
	// 			res, _ := json.Marshal("Len of name must be more than 2 symbols")
	// 			w.Write(res)
	// 			return
	// 		}
	// 		user.Name = updateUser.Name
	// 	}

	// 	if updateUser.Age > 0 {
	// 		user.Age = updateUser.Age
	// 	}

	// 	users[id] = user
	// 	res, _ := json.Marshal(users[id])
	// 	w.Write(res)
	// })

	srv := &http.Server{
		Addr:        ":8080",
		Handler:     mux,
		ReadTimeout: 5 * time.Second,
	}

	go srv.ListenAndServe()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	srv.Shutdown(shutdownCtx)
}

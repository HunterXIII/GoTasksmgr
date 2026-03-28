package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"tasksmgr/functions"
	"tasksmgr/handler"
	"tasksmgr/indexer"
	"tasksmgr/interceptor"
	"tasksmgr/middleware"
	"tasksmgr/repo"
	"time"

	pb "tasksmgr/gen"

	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
)

type User struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {

	db, err := sql.Open("pgx", "postgres://postgres:postgres@localhost:5435/mydb")
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go functions.StepFunction(ctx)

	mux.HandleFunc("GET /reports/slow", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Start slow endpoint")

		select {
		case <-time.After(5 * time.Second):
			fmt.Println("Finish slow endpoint")
			res, _ := json.Marshal("Finished")
			w.Write(res)
		case <-r.Context().Done():
			fmt.Println("Cancel slow endpoint")
			return
		}
	})

	lis, err := net.Listen("tcp", ":9091")
	if err != nil {
		log.Fatal(err)
	}

	notesRepo := repo.NewNotesRepository(db)
	notesHandler := handler.NewNotesHandler(notesRepo)

	srv_gRPC := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.RequestIDInterceptor,
			authHandler.JWTAuthInterceptor,
		),
	)
	pb.RegisterNotesServiceServer(srv_gRPC, notesHandler)
	go srv_gRPC.Serve(lis)

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

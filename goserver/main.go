package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func formHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "parse form error: "+err.Error(), http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	message := r.FormValue("message")

	fmt.Printf("name=%s email=%s message=%s\n", name, email, message)
	fmt.Fprintln(w, "Post successfully")
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not supported", http.StatusMethodNotAllowed)
		return
	}

	fmt.Fprintln(w, "Hello world")
}

func main() {
	fileServer := http.FileServer(http.Dir("./static"))

	mux := http.NewServeMux()

	mux.Handle("/", fileServer)
	mux.HandleFunc("/hello", helloHandler)
	mux.HandleFunc("/form", formHandler)
	mux.HandleFunc("/submit", formHandler)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	//  1 listen for os signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	//  2 run and catch the error in channel
	errChan := make(chan error, 1)

	go func() {
		fmt.Println("Starting the server in 8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
			log.Fatal(err)
		}

	}()

	select {
	case sig := <-signals:
		fmt.Println("recive %s , shutting down...\n", sig)
	case err := <-errChan:
		log.Fatalf("server error: %v", err)
	}

	//  making a context giving 10 sec to server to complet the task for just close the all the pending task [human written comment]

	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancle()

	if err := srv.Shutdown(ctx); err != nil {
		log.Println("graceful shutdown failed %v", err)
		_ = srv.Close()
	}

	// 5. signal completion through runchan
	runchan := make(chan struct{}, 1)
	close(runchan)
	fmt.Println("server stopped cleanly")

}

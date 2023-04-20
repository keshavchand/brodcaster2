package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type ServerInfo struct {
	sync.Mutex
	Ctx context.Context
	// this map is from the notification channel
	// to the queue where one should write the info
	Nmap map[string]*Queue[[]byte]
}

func main() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Starting")

	info := ServerInfo{}
	ctx, stop := context.WithCancel(context.Background())
	info.Ctx = ctx
	info.Nmap = make(map[string]*Queue[[]byte])

	mux := http.NewServeMux()
	mux.HandleFunc("/n/", info.BrodcastNotifications)
	mux.HandleFunc("/s/", info.HandleNotifications)
	mux.Handle("/", http.FileServer(http.Dir("./static")))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	server.RegisterOnShutdown(stop)
	go func() {
		err := server.ListenAndServe()
		log.Printf("server shutdown: %v\n", err)
	}()

	<-done
	close(done)
	duration := 5 * time.Second
	log.Println("Shutting Down in", duration, "seconds")
	c, _ := context.WithTimeout(context.Background(), duration)
	server.Shutdown(c)
}

var (
	RoomNotPresentError = errors.New("room not present")
	WrongRoomError      = errors.New("wrong room name")
)

func (server *ServerInfo) Reader(uri string) (Reader[[]byte], error) {
	path := strings.Split(uri, "/")[1:]
	if len(path) != 2 {
		return Reader[[]byte]{}, WrongRoomError
	}

	log.Printf(path[1])

	server.Lock()
	defer server.Unlock()
	q, ok := server.Nmap[path[1]]
	if !ok {
		q = NewQueue[[]byte]()
		server.Nmap[path[1]] = q
	}

	return q.NewReader(), nil
}

func (server *ServerInfo) HandleNotifications(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		server.HandleNotificationsPost(w, r)
	case "Get":
		server.HandleNotificationsGet(w, r)
	default:
		return
	}
}

func (server *ServerInfo) HandleNotificationsGet(w http.ResponseWriter, r *http.Request) {
	log.Println("NOT IMPLEMENTED")
	w.WriteHeader(500)
}

func (server *ServerInfo) HandleNotificationsPost(w http.ResponseWriter, r *http.Request) {
	log.Println("Recieving Request")

	type Req struct {
		Room string
		Msg  string
	}

	var req Req

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		log.Println(err)
	}

	if req.Room == "" {
		w.WriteHeader(400)
		io.WriteString(w, "Bad Request Error: No Room")
		return
	}
	if len(req.Msg) == 0 {
		w.WriteHeader(400)
		io.WriteString(w, "Bad Request Error: No Msg")
		return
	}

	server.Lock()
	defer server.Unlock()
	q, ok := server.Nmap[req.Room]
	if !ok {
		q = NewQueue[[]byte]()
		server.Nmap[req.Room] = q
	}

	q.Insert([]byte(req.Msg))
}

func (server *ServerInfo) BrodcastNotifications(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("Content-Type", "text/event-stream")
	w.Header().Add("Cache-Control", "no-cache")

	reader, err := server.Reader(r.URL.Path)
	if err != nil {
		switch {
		case errors.Is(err, WrongRoomError):
		case errors.Is(err, RoomNotPresentError):
			w.WriteHeader(400)
			io.WriteString(w, fmt.Sprintf("Internal Server Error: %s", err.Error()))
		default:
			w.WriteHeader(500)
			io.WriteString(w, "Unknown internal error")
		}
		return
	}

	for {
		select {
		case <-r.Context().Done():
			return
		case <-server.Ctx.Done():
			return
		case <-reader.Wait():
			result, stop := reader.Read()
			if stop {
				return
			}
			w.Write([]byte("data: "))
			w.Write(result)
			w.Write([]byte("\n\n"))
			flusher, ok := w.(http.Flusher)
			if !ok {
				return
			}
			flusher.Flush()
		}
	}
}

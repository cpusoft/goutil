package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"

	websub "github.com/cpusoft/goutil/websubutil"
	"github.com/cpusoft/goutil/websubutil/store/bolt"
)

func main() {
	store, err := bolt.New("hub.db")

	if err != nil {
		log.Fatal(err)
	}

	h := websub.New(store)

	r := http.NewServeMux()

	r.HandleFunc("/", h.ServeHTTP)

	log.Println("Starting server on :8080")

	go http.ListenAndServe(":8080", r)

	interrupt := make(chan os.Signal, 1)

	signal.Notify(interrupt, os.Interrupt)

	<-interrupt
}

package main

import (
	"context"
	"log"
	"net/http"

	"github.com/devdevaraj/conductor/dns_manager"
	"github.com/devdevaraj/conductor/examiner"
	"github.com/devdevaraj/conductor/handle_proxy"
	"github.com/devdevaraj/conductor/handlers"
	"github.com/devdevaraj/conductor/init_redis"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

var (
	ctx = context.Background()
	rdb *redis.Client
)

func main() {
	// global_init.Init()
	rdb = init_redis.InitRedis(ctx)
	router := mux.NewRouter()

	router.HandleFunc("/remove-container", func(w http.ResponseWriter, r *http.Request) {
		dns_manager.RemoveContainer(w, r, rdb, ctx)
	})
	router.HandleFunc("/close-port/{short_id}", func(w http.ResponseWriter, r *http.Request) {
		dns_manager.ClosePort(w, r, rdb, ctx)
	})
	router.HandleFunc("/open-port", func(w http.ResponseWriter, r *http.Request) {
		dns_manager.OpenPort(w, r, rdb, ctx)
	})
	router.HandleFunc("/wait-for-vms", handlers.WaitForVMs)

	router.HandleFunc("/examiner", examiner.Examiner)

	router.HandleFunc("/{id}/{vm}", func(w http.ResponseWriter, r *http.Request) {
		handle_proxy.HandleProxy(w, r, rdb, ctx)
	})

	log.Println("Starting Conductor API server on :8082...")
	log.Fatal(http.ListenAndServe(":8082", router))
}

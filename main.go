package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func serveDefault(w http.ResponseWriter, r *http.Request) {
	// /response ok
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func serveSignin(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/signin" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "signin.html")
}

func main() {
	hub := H

	router := mux.NewRouter()

	go hub.Run()
	router.Methods("GET").Path("/").HandlerFunc(serveDefault)
	router.Methods("POST").Path("/signin").HandlerFunc(handleSignin)
	router.Methods("POST").Path("/signup").HandlerFunc(handleSignup)
	router.HandleFunc("/ws", ServeWs)

	router.Methods("GET").Path("/user").HandlerFunc(handleListUsers)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)

	srv := &http.Server{
		Handler: handler,
		Addr:    "0.0.0.0:8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

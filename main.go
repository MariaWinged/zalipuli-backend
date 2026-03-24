package main

import (
	"log"
	"os"

	"zalipuli/internal/cmd/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := server.New(":" + port)
	log.Println("starting server on", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

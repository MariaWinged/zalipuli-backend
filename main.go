package main

import (
	"log"

	"zalipuli/internal/cmd/server"
)

func main() {
	srv := server.New(":8080")
	log.Println("starting server on", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

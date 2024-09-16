package main

import (
	"fmt"
	"log"

	"github.com/Garetonchick/personal-blog/internal/server"
)

func main() {
	host := "0.0.0.0"
	port := "4444"
	addr := host + ":" + port

	srv, err := server.New()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Server is listening on %s address\n", addr)

	if err := srv.ListenAndServe(addr); err != nil {
		log.Fatal(err)
	}
}

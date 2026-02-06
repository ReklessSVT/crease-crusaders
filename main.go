package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// 'handleHome' is automatically found in handlers.go
	http.HandleFunc("/", handleHome)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("🏒 Crease Crusaders Hub is live on port %s\n", port)

	// Start the server
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

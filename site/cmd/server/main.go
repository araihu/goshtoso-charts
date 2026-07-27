package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/araihu/goshtoso-charts/site/internal/server"
)

func main() {
	port := flag.Int("port", 8091, "HTTP port")
	flag.Parse()
	address := fmt.Sprintf(":%d", *port)
	log.Printf("Goshtoso Charts demo: http://localhost%s", address)
	log.Fatal(http.ListenAndServe(address, server.New()))
}

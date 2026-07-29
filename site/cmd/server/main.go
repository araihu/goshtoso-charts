package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/araihu/goshtoso-charts/site/internal/buildinfo"
	"github.com/araihu/goshtoso-charts/site/internal/server"
)

func main() {
	port := flag.Int("port", 8091, "HTTP port")
	showVersion := flag.Bool("version", false, "print site version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println(buildinfo.SiteVersion())
		return
	}
	address := fmt.Sprintf(":%d", *port)
	log.Printf("Goshtoso Charts demo %s: http://localhost%s", buildinfo.SiteVersion(), address)
	log.Fatal(http.ListenAndServe(address, server.New()))
}

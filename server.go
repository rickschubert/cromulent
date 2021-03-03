package main

import (
	"log"
	"net/http"
)

// App represents the server's internal state.
// It holds configuration about providers and content
type App struct {
	ContentClients map[Provider]Client
	Config         ContentMix
}

func (App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s", req.Method, req.URL.String())
	w.WriteHeader(http.StatusNotImplemented)
}

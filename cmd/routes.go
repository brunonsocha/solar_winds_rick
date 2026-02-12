package main

import (
	"net/http"
	"fmt"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /search", app.searchCharacter)
	mux.HandleFunc("GET /top-pairs", app.topPairs)
	return mux
}



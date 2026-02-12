package main

import (
	"encoding/json"
	"net/http"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /search", app.search)
	// mux.HandleFunc("GET /top-pairs", app.topPairs)
	return mux
}

// will put handlers here instead of making a seperate handlers.go file

func (app *application) search(w http.ResponseWriter, r *http.Request) {
	term := r.URL.Query().Get("term")
	payload, err := app.characterService.GetPayload(term)
	if err != nil {
		app.errorLog.Println(err)
		http.Error(w, "error processing payload", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(payload); err != nil {
		app.errorLog.Println(err)
		http.Error(w, "error providing a response", http.StatusInternalServerError)
	}
}

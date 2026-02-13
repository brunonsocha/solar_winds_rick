package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /search", app.search)
	mux.HandleFunc("GET /top-pairs", app.topPairs)
	return mux
}

func (app *application) search(w http.ResponseWriter, r *http.Request) {
	term := r.URL.Query().Get("term")
	limitStr := r.URL.Query().Get("limit")
	var limit int
	if limitStr == "" {
		limit = 0
	} else {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			app.errorLog.Println(err)
			http.Error(w, "incorrect limit", http.StatusBadRequest)
			return
		}
	}
	payload, err := app.characterService.GetSearchPayload(term, limit)
	if err != nil {
		app.errorLog.Println(err)
		http.Error(w, "error processing payload", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(payload); err != nil {
		app.errorLog.Println(err)
		http.Error(w, "error providing a response", http.StatusInternalServerError)
	}
}

func (app *application) topPairs(w http.ResponseWriter, r *http.Request) {
	minStr := r.URL.Query().Get("min")
	maxStr := r.URL.Query().Get("max")
	limitStr := r.URL.Query().Get("limit")
	var limit int
	// the takehome specifies that limit is optional, but doesn't say that min or max are optional
	if minStr == "" || maxStr == "" {
		app.errorLog.Println(fmt.Errorf("incorrect min/max parameters"))
		http.Error(w, "incorrect min/max parameters", http.StatusBadRequest)
		return
	}
	minVal, err := strconv.Atoi(minStr)
	if err != nil {
		app.errorLog.Println(fmt.Errorf("incorrect min parameter"))
		http.Error(w, "incorrect min parameter", http.StatusBadRequest)
		return
	}
	maxVal, err := strconv.Atoi(maxStr)
	if err != nil {
		app.errorLog.Println(fmt.Errorf("incorrect max parameter"))
		http.Error(w, "incorrect max parameter", http.StatusBadRequest)
		return
	}
	if minVal > maxVal {
		app.errorLog.Println(fmt.Errorf("min parameter can't be higher than the max parameter"))
		http.Error(w, "min can't be higher than max", http.StatusBadRequest)
		return
	}
	limit = 20
	if limitStr != "" {
		var err error
		val, err := strconv.Atoi(limitStr)
		if err != nil {
			app.errorLog.Println(err)
			http.Error(w, "incorrect limit", http.StatusBadRequest)
			return
		}
		if val > 0 {
			limit = val
		}
	}
	payload, err := app.characterService.GetPairsPayload(minVal, maxVal, limit)
	if err != nil {
		app.errorLog.Println(err)
		http.Error(w, "error processing payload", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(payload); err != nil {
		app.errorLog.Println(err)
		http.Error(w, "error providing a response", http.StatusInternalServerError)
	}
}

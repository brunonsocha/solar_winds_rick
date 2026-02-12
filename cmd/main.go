package main

import (
	"log"
	"net/http"
	"staz_morty/internal"
	"time"
	"os"
)

type application struct {
	infoLog *log.Logger
	errorLog *log.Logger
	characterService *internal.SearchService
}


func main() {
	infoLog := log.New(os.Stdout, "[INFO]\t", log.Ltime)
	errorLog := log.New(os.Stderr, "[BŁĄD]\t", log.Ltime)
	httpClient := &http.Client{Timeout: 10*time.Second}
	service := &internal.SearchService{
		Client: httpClient,
		Url: "https://rickandmortyapi.com/api",
	}
	app := &application{
		infoLog: infoLog,
		errorLog: errorLog,
		characterService: service,
	}
	srv := &http.Server{
		Addr: ":8080",
		Handler: app.routes(),
	}
	app.infoLog.Printf("Starting server on %s", srv.Addr)
	srv.ListenAndServe()
}

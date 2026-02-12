package internal

import (
	"net/http"
	"time"
)

type CharacterService struct {
	Client *http.Client
	Url string
}

type ApiRes struct {
	Info InfoRes `json:"info"`
	Results []struct{} `json:"results"` 
}

type InfoRes struct {
	Count int `json:"count"`
	Pages int `json:"pages"`
	Next *string `json:"next"`
	Prev *string `json:"prev"`
}

type CharacterRes struct {
	Id int `json:"id"`
	Name string `json:"name"`
	Status string `json:"status"`
	Species string `json:"species"`
	Type string `json:"type"`
	Gender string `json:"gender"`
	Origin struct {
		Name string `json:"name"`
		Url string `json:"url"`
	} `json:"origin"`
	Location struct {
		Name string `json:"name"`
		Url string `json:"url"`
	} `json:"location"`
	Image string `json:"image"`
	Episode []string `json:"episode"`
	Url string `json:"url"`
	Created time.Time `json:"created"`
}

type LocationRes struct {
	Id int `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Dimension string `json:"dimension"`
	Residents []string `json:"residents"`
	Url string `json:"url"`
	Created time.Time `json:"created"`
}

type EpisodeRes struct {
	Id int `json:"id"`
	Name string `json:"name"`
	AirDate string `json:"air_date"`
	Episode string `json:"episode"`
	Characters []string `json:"characters"`
	Url string `json:"url"`
	Created time.Time `json:"created"`
}

type SearchResult struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Url string `json:"url"`
}

type SearchPayload struct {
	Info []SearchResult 
}

func (c *CharacterService) GetPayload(term string, limit int) (*SearchPayload, error) {
	// spawn 3 goroutines
}

func getCharacterData(term string) ([]CharacterRes, error){
	
}

func getLocationData(term string) ([]LocationRes, error) {

}

func getEpisodeData(term string) ([]EpisodeRes, error) {

}

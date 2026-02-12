package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type CharacterService struct {
	Client *http.Client
	Url string
}

type ApiCharRes struct {
	Info InfoRes `json:"info"`
	Results []CharacterRes `json:"results"` 
}

type ApiLocRes struct {
	Info InfoRes `json:"info"`
	Results []LocationRes `json:"results"`
}

type ApiEpRes struct {
	Info InfoRes `json:"info"`
	Results []EpisodeRes `json:"results"`
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

func (c *CharacterService) GetPayload(term string) (*SearchPayload, error) {
	characters, err := c.getCharacterData(term)
	if err != nil {
		return nil, err
	}
	locations, err := c.getLocationData(term)
	if err != nil {
		return nil, err
	}
	episodes, err := c.getEpisodeData(term)
	if err != nil {
		return nil, err
	}
	var results []SearchResult
	for _, i := range characters {
		results = append(results, SearchResult{
			Name: i.Name,
			Type: "character",
			Url: i.Url,
		})
	}
	for _, i := range locations {
		results = append(results, SearchResult{
			Name: i.Name,
			Type: "location",
			Url: i.Url,
		})
	}
	for _, i := range episodes {
		results = append(results, SearchResult{
			Name: i.Name,
			Type: "episode",
			Url: i.Url,
		})
	}
	return &SearchPayload{Info: results}, nil
}

func (c * CharacterService) getCharacterData(term string) ([]CharacterRes, error){
	fullUrl := fmt.Sprintf("%s/character/?name=%s", c.Url, term)
	response, err := c.Client.Get(fullUrl)
	if err != nil {
		return nil, err
	}
	if response.StatusCode == http.StatusNotFound {
		return []CharacterRes{}, nil
	}
	// could check for status code as well, but the unmarshal will throw an error anyway
	defer response.Body.Close()
	var apiRes ApiCharRes
	jsonBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(jsonBody, &apiRes); err != nil {
		return nil, err
	}
	return apiRes.Results, nil
}

func (c *CharacterService) getLocationData(term string) ([]LocationRes, error) {
	fullUrl := fmt.Sprintf("%s/location/?name=%s", c.Url, term)
	response, err := c.Client.Get(fullUrl)
	if err != nil {
		return nil, err
	}
	if response.StatusCode == http.StatusNotFound {
		return []LocationRes{}, nil
	}
	defer response.Body.Close()
	var apiRes ApiLocRes
	jsonBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(jsonBody, &apiRes); err != nil {
		return nil, err
	}
	return apiRes.Results, nil
}

func (c *CharacterService) getEpisodeData(term string) ([]EpisodeRes, error) {
	fullUrl := fmt.Sprintf("%s/episode/?name=%s", c.Url, term)
	response, err := c.Client.Get(fullUrl)
	if err != nil {
		return nil, err
	}
	if response.StatusCode == http.StatusNotFound {
		return []EpisodeRes{}, nil
	}
	defer response.Body.Close()
	var apiRes ApiEpRes
	jsonBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(jsonBody, &apiRes); err != nil {
		return nil, err
	}
	return apiRes.Results, nil
}

package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
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

func (c *CharacterService) GetPayload(term string, limit int) ([]SearchResult, error) {
	resultChan := make(chan SearchResult, limit)
	var results []SearchResult
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		chars, err := c.getCharacterData(term)
		if err != nil {
			return
		}
		for _, i := range chars {
			resultChan <- SearchResult{Name: i.Name, Type: "character", Url: i.Url}
		}
	}()	
	wg.Add(1)
	go func() {
		locs, err := c.getLocationData(term)
		if err != nil {
			return
		}
		for _, i := range locs {
			resultChan <- SearchResult{Name: i.Name, Type: "location", Url: i.Url}
		}
	}()	
	wg.Add(1)
	go func() {
		eps, err := c.getEpisodeData(term)
		if err != nil {
			return
		}
		for _, i := range eps {
			resultChan <- SearchResult{Name: i.Name, Type: "episode", Url: i.Url}
		}
	}()	
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	for i := range resultChan{
		results = append(results, i)
		if len(results) == limit {
			break
		}
	}
	return results, nil
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

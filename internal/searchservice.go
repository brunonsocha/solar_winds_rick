package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type SearchService struct {
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

type IdPair struct {
	Character1 int
	Character2 int
}

type PairsResult struct {
	Character1 struct {
		Name string `json:"name"`
		Url string `json:"url"`
	} `json:"character2"`
	Character2 struct {
		Name string `json:"name"`
		Url string `json:"url"`
	} `json:"character2"`
	Episodes int `json:"episodes"`
}

func (s *SearchService) GetSearchPayload(term string, limit int) ([]SearchResult, error) {
	resultChan := make(chan SearchResult, limit)
	done := make(chan struct{})
	defer close(done)
	var results []SearchResult
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		chars, err := s.getCharacterData(term)
		if err != nil {
			return
		}
		for _, i := range chars {
			select {
			case resultChan <- SearchResult{Name: i.Name, Type: "character", Url: i.Url}:
			case <- done:
			return
			}
		}
	}()	
	wg.Add(1)
	go func() {
		defer wg.Done()
		locs, err := s.getLocationData(term)
		if err != nil {
			return
		}
		for _, i := range locs {
			select {
			case resultChan <- SearchResult{Name: i.Name, Type: "location", Url: i.Url}:
			case <- done:
				return
			}
		}
	}()	
	wg.Add(1)
	go func() {
		defer wg.Done()
		eps, err := s.getEpisodeData(term)
		if err != nil {
			return
		}
		for _, i := range eps {
			select {
			case resultChan <- SearchResult{Name: i.Name, Type: "episode", Url: i.Url}:
			case <- done:
				return
			}
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

func (s * SearchService) getCharacterData(term string) ([]CharacterRes, error){
	fullUrl := fmt.Sprintf("%s/character/?name=%s", s.Url, term)
	response, err := s.Client.Get(fullUrl)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		return []CharacterRes{}, nil
	}
	// could check for status code as well, but the unmarshal will throw an error anyway
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

func (s *SearchService) getLocationData(term string) ([]LocationRes, error) {
	fullUrl := fmt.Sprintf("%s/location/?name=%s", s.Url, term)
	response, err := s.Client.Get(fullUrl)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		return []LocationRes{}, nil
	}
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

func (s *SearchService) getEpisodeData(term string) ([]EpisodeRes, error) {
	fullUrl := fmt.Sprintf("%s/episode/?name=%s", s.Url, term)
	response, err := s.Client.Get(fullUrl)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		return []EpisodeRes{}, nil
	}
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

func (s *SearchService) GetPairsPayload(minVal, maxVal, limit int) ([]PairsResult, error) {
	episodes, err := s.getEpisodeData("")
	var characterPairs map[]
	if err != nil {
		return nil, err
	}
	for _, i := range episodes {
		for _, j := range i.Characters {
			
		}
	}
}

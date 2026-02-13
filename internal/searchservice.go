package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strconv"
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

type PairStats struct {
	Pair IdPair
	Count int
}

type PairsResult struct {
	Character1 struct {
		Name string `json:"name"`
		Url string `json:"url"`
	} `json:"character1"`
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
		s.getCharacterData(term, resultChan, done)
	}()	
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.getLocationData(term, resultChan, done)
	}()	
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.getEpisodeData(term, resultChan, done)
	}()	
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	for i := range resultChan{
		results = append(results, i)
		if len(results) == limit {
			return results, nil
		}
	}
	return results, nil
}

func (s * SearchService) getCharacterData(term string, results chan<- SearchResult, done <-chan struct{}) {
	fullUrl := fmt.Sprintf("%s/character/?name=%s", s.Url, url.QueryEscape(term))
	for fullUrl != ""{
		response, err := s.Client.Get(fullUrl)
		if err != nil {
			return 
		}
		if response.StatusCode != http.StatusOK {
			response.Body.Close()
			return
		}
		var apiRes ApiCharRes
		jsonBody, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			return
		}
		if err := json.Unmarshal(jsonBody, &apiRes); err != nil {
			return
		}
		for _, i := range apiRes.Results {
			select {
			case results <- SearchResult{Name: i.Name, Type: "character", Url: i.Url}:
			case <- done:
				return
			}
		}
		if apiRes.Info.Next != nil {
			fullUrl = *apiRes.Info.Next
		} else {
			fullUrl = ""
		}
		
	}
}

func (s *SearchService) getLocationData(term string, results chan<- SearchResult, done <-chan struct{}) {
	fullUrl := fmt.Sprintf("%s/location/?name=%s", s.Url, url.QueryEscape(term))
	for fullUrl != ""{
		response, err := s.Client.Get(fullUrl)
		if err != nil {
			return 
		}
		if response.StatusCode != http.StatusOK {
			response.Body.Close()
			return
		}
		var apiRes ApiLocRes
		jsonBody, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			return
		}
		if err := json.Unmarshal(jsonBody, &apiRes); err != nil {
			return
		}
		for _, i := range apiRes.Results {
			select {
			case results <- SearchResult{Name: i.Name, Type: "location", Url: i.Url}:
			case <- done:
				return
			}
		}
		if apiRes.Info.Next != nil {
			fullUrl = *apiRes.Info.Next
		} else {
			fullUrl = ""
		}
	}
}

func (s *SearchService) getEpisodeData(term string, results chan<- SearchResult, done <-chan struct{}) {
	fullUrl := fmt.Sprintf("%s/episode/?name=%s", s.Url, url.QueryEscape(term))
	for fullUrl != ""{
		response, err := s.Client.Get(fullUrl)
		if err != nil {
			return
		}
		if response.StatusCode != http.StatusOK {
			response.Body.Close()
			return
		}
		var apiRes ApiEpRes
		jsonBody, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			return
		}
		if err := json.Unmarshal(jsonBody, &apiRes); err != nil {
			return
		}
		for _, i := range apiRes.Results {
			select {
			case results <- SearchResult{Name: i.Name, Type: "episode", Url: i.Url}:
			case <- done:
				return
			}
		}
		if apiRes.Info.Next != nil {
			fullUrl = *apiRes.Info.Next
		} else {
			fullUrl = ""
		}
		
	}
}

func (s *SearchService) GetPairsPayload(minVal, maxVal, limit int) ([]PairsResult, error) {
	episodes, err := s.getAllEpisodes("")
	characterPairs := make(map[IdPair]int)
	charactersToRetrieve := make(map[int]struct{})
	var result []PairsResult
	var idsToRetrieve []int
	var pairStats []PairStats
	if err != nil {
		return nil, err
	}
	for _, ep := range episodes {
		var charIDs []int
		for _, charUrl := range ep.Characters {
			// no regex needed
			id, err := strconv.Atoi(path.Base(charUrl))
			if err != nil {
				return nil, err
			}
			charIDs = append(charIDs, id)
		}
		for i := 0; i < len(charIDs); i++ {
			for j := i+1; j < len(charIDs); j++ {
				id1 := charIDs[i]
				id2 := charIDs[j]
				pair := IdPair{
					Character1: min(id1, id2),
					Character2: max(id1, id2),
				}
				characterPairs[pair]++
			}
		}
	}
	for k, v := range characterPairs {
		if v >= minVal && v <= maxVal {
			pairStats = append(pairStats, PairStats{Pair: k, Count: v})
		}
	}
	sort.Slice(pairStats, func(i, j int) bool{
		return pairStats[i].Count > pairStats[j].Count
	})
	if limit > 0 && len(pairStats) > limit {
		pairStats = pairStats[:limit]
	}
	for _, i := range pairStats {
		charactersToRetrieve[i.Pair.Character1] = struct{}{}
		charactersToRetrieve[i.Pair.Character2] = struct{}{}
	}
	for id, _ := range charactersToRetrieve {
		idsToRetrieve = append(idsToRetrieve, id)
	}
	names, err := s.getCharacters(idsToRetrieve)
	if err != nil {
		return nil, err
	}
	for _, stat := range pairStats {
		p := PairsResult{}
		p.Character1.Name = names[stat.Pair.Character1]
		p.Character1.Url = fmt.Sprintf("%s/character/%d", s.Url, stat.Pair.Character1)
		p.Character2.Name = names[stat.Pair.Character2]
		p.Character2.Url = fmt.Sprintf("%s/character/%d", s.Url, stat.Pair.Character2)
		p.Episodes = stat.Count
		result = append(result, p)
	}
	return result, nil
}

func (s *SearchService) getAllEpisodes(term string) ([]EpisodeRes, error){
	fullUrl := fmt.Sprintf("%s/episode/?name=%s", s.Url, url.QueryEscape(term))
	var results []EpisodeRes
	for fullUrl != ""{
		response, err := s.Client.Get(fullUrl)
		if err != nil {
			return nil, err
		}
		if response.StatusCode == http.StatusNotFound {
			response.Body.Close()
			return results, nil
		}
		var apiRes ApiEpRes
		jsonBody, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			return results, err
		}
		if err := json.Unmarshal(jsonBody, &apiRes); err != nil {
			return results, err
		}
		results = append(results, apiRes.Results...)
		if apiRes.Info.Next != nil {
			fullUrl = *apiRes.Info.Next
		} else {
			fullUrl = ""
		}
	}
	return results, nil
}
func (s *SearchService) getCharacters(characterIds []int) (map[int]string, error) {
	chars := ""
	result := make(map[int]string)
	if len(characterIds) == 0 {
		return result, nil
	}
	for _, i := range characterIds {
		chars += fmt.Sprintf("%d,", i)
	}
	chars = chars[:(len(chars)-1)]
	fullUrl := fmt.Sprintf("%s/character/%s", s.Url, chars)
	response, err := s.Client.Get(fullUrl)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	jsonBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if len(characterIds) == 1 {
		var char CharacterRes
		if err := json.Unmarshal(jsonBody, &char); err != nil {
			return nil, err
		}
		result[char.Id] = char.Name
	} else {
		var chars []CharacterRes
		if err := json.Unmarshal(jsonBody, &chars); err != nil {
			return nil, err
		}
		for _, char := range chars {
			result[char.Id] = char.Name
		}
	}
	return result, nil
}

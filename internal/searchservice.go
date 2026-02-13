package internal

import (
	"encoding/json"
	"fmt"
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
	// pointers to handle nulls
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
	// create a channel for structs and for singalling that we're done
	resultChan := make(chan SearchResult, limit)
	done := make(chan struct{})
	// close done on function return to end goroutines
	defer close(done)
	var results []SearchResult
	var wg sync.WaitGroup
	// spawn goroutines to search for characters, eps and locations
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
	// wait for goroutines to finish and close the result chan
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

// comments in this function apply to location and episode functions as well, they're virtually identical
// maybe there's room to use generics here?
func (s * SearchService) getCharacterData(term string, results chan<- SearchResult, done <-chan struct{}) {
	fullUrl := fmt.Sprintf("%s/character/?name=%s", s.Url, url.QueryEscape(term))
	// loop for pagination
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
		if err := json.NewDecoder(response.Body).Decode(&apiRes); err != nil {
			response.Body.Close()
			return
		}
		response.Body.Close()
		// iterate over results, send them to the results chan, check if we're done
		for _, i := range apiRes.Results {
			select {
			case results <- SearchResult{Name: i.Name, Type: "character", Url: i.Url}:
			case <- done:
				return
			}
		}
		// proceed to the next url if there is one
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
		if err := json.NewDecoder(response.Body).Decode(&apiRes); err != nil {
			response.Body.Close()
			return
		}
		response.Body.Close()
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
		if err := json.NewDecoder(response.Body).Decode(&apiRes); err != nil {
			response.Body.Close()
			return
		}
		response.Body.Close()
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
	// use a helper function that gets ALL episodes
	episodes, err := s.getAllEpisodes("")
	// map of structs holding pairs + occurences
	characterPairs := make(map[IdPair]int)
	// basically a set:
	charactersToRetrieve := make(map[int]struct{})
	var result []PairsResult
	// what we'll feed to the getCharacters function
	var idsToRetrieve []int
	// slice of structs with similar structure to the characterPairs map - for sorting
	var pairStats []PairStats
	if err != nil {
		return nil, err
	}
	// create a list of character IDs, populate the map with pairs 
	for _, ep := range episodes {
		var charIDs []int
		for _, charUrl := range ep.Characters {
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
				// make sure that if the pair consists of the same characters, it's always the same
				pair := IdPair{
					Character1: min(id1, id2),
					Character2: max(id1, id2),
				}
				characterPairs[pair]++
			}
		}
	}
	// filter out pairs with occurences that don't fit within our min/max parameters
	for k, v := range characterPairs {
		if v >= minVal && v <= maxVal {
			pairStats = append(pairStats, PairStats{Pair: k, Count: v})
		}
	}
	// we can sort the slice, couldn't do so with the map
	sort.Slice(pairStats, func(i, j int) bool{
		return pairStats[i].Count > pairStats[j].Count
	})
	// limit cut off
	if limit > 0 && len(pairStats) > limit {
		pairStats = pairStats[:limit]
	}
	// use the set for checking what names we should get
	for _, i := range pairStats {
		charactersToRetrieve[i.Pair.Character1] = struct{}{}
		charactersToRetrieve[i.Pair.Character2] = struct{}{}
	}
	// get a slice out of them
	for id, _ := range charactersToRetrieve {
		idsToRetrieve = append(idsToRetrieve, id)
	}
	names, err := s.getCharacters(idsToRetrieve)
	if err != nil {
		return nil, err
	}
	// populate results
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
		if err := json.NewDecoder(response.Body).Decode(&apiRes); err != nil {
			response.Body.Close()
			return results, nil
		}
		response.Body.Close()
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
		return map[int]string{}, nil
	}

	if len(characterIds) == 1 {
		var char CharacterRes
		if err := json.NewDecoder(response.Body).Decode(&char); err != nil {
			return nil, err
		}
		result[char.Id] = char.Name
		return result, nil
	}
	var charList []CharacterRes
	if err := json.NewDecoder(response.Body).Decode(&charList); err != nil {
		return nil, err
	}
	for _, char := range charList {
		result[char.Id] = char.Name
	}
	return result, nil
}

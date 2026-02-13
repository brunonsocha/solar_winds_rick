package internal

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testServer(t *testing.T, handler http.HandlerFunc) (*SearchService, *httptest.Server) {
	srv := httptest.NewServer(handler)
	service := &SearchService{
		Client: srv.Client(),
		Url: srv.URL,
	}
	return service, srv
}

func TestGetSearchPayload(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/character/") {
			response := ApiCharRes{
				Results: []CharacterRes{
					{
						Name: "Rick Test",
						Type: "character",
						Url: "https://testing.com/api/character/1",
					},
				},
			}
			payload, _ := json.Marshal(response)
			w.Write(payload)
			return
		}
		if strings.Contains(r.URL.Path, "/location/") {
			response := ApiLocRes{
				Results: []LocationRes{
					{
						Name: "Planet Test",
						Type: "location",
						Url: "https://testing.com/api/location/1",
					},
				},
			}
			payload, _ := json.Marshal(response)
			w.Write(payload)
			return
		}
		if strings.Contains(r.URL.Path, "/episode/") {
			response := ApiEpRes{
				Results: []EpisodeRes{
					{
						Name: "Episode Test",
						Episode: "S11E11",
						Url: "https://testing.com/api/episode/1",
					},
				},
			}
			payload, _ := json.Marshal(response)
			w.Write(payload)
			return
		}
	}
	service, srv := testServer(t, handler)
	defer srv.Close()
	results, err := service.GetSearchPayload("test", 5)
	if err != nil {
		t.Fatalf("Error %v", err)
	}
	if len(results) != 3 {
		t.Errorf("Incorrect number of results - %d/3", len(results))
	}
	t.Logf("%d/3 results passed", len(results))
}

func TestGetPairsPayload(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/episode/") {
			response := ApiEpRes{
				Results: []EpisodeRes{
					{
						Name: "Episode Test",
						Episode: "S11E11",
						Characters: []string{
							"https://testing.com/api/character/1",
							"https://testing.com/api/character/2",

						},
					},
				},
			}
			payload, _ := json.Marshal(response)
			w.Write(payload)
			return
		}
		if strings.Contains(r.URL.Path, "/character/") {
			response := []CharacterRes{
					{
						Id: 1, 
						Name: "Test Rick",
					},
					{
						Id: 2, 
						Name: "Test Morty",
					},
				}
			payload, _ := json.Marshal(response)
			w.Write(payload)
			return
		}
	}
	service, srv := testServer(t, handler)
	defer srv.Close()
	pairs, err := service.GetPairsPayload(0, 10, 10)
	if err != nil {
		t.Fatalf("Error %v", err)
	}
	if len(pairs) != 1 {
		t.Errorf("Incorrect number of results - %d/1", len(pairs))
	}
	p := pairs[0]
	if p.Character1.Name != "Test Rick" {
		t.Errorf("Name should be Test Rick - %s", p.Character1.Name)
	}
	t.Logf("%d/1 results passed", len(pairs))
}

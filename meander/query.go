package meander

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"unicode"

	"github.com/web-recommendation-service/config"
)

type Place struct {
	Geometry `json:"geometry"`
	Name     string   `json:"name"`
	Icon     string   `json:"icon"`
	Photos   []*Photo `json:"photos"`
	Vicinity string   `json:"vicinity"`
}

type detailResult struct {
	response DetailResponse
	err      error
}

type DetailResponse struct {
	Result Place `json:"result"`
}

type nearbyResponse struct {
	Predictions  []Prediction `json:"predictions"`
	InfoMessages []string     `json:"info_messages"`
	ErrorMessage string       `json:"error_message"`
	Status       string       `json:"status"`
}

type Prediction struct {
	Description    string `json:"description"`
	PlaceID        string `json:"place_id"`
	DistanceMeters int    `json:"distance_meters"`
}

type Geometry struct {
	Location `json:"location"`
}

type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Photo struct {
	PhotoRef string `json:"photo_reference"`
	URL      string `json:"url"`
}

func (p *Place) Public() interface{} {
	return map[string]interface{}{
		"name":     p.Name,
		"icon":     p.Icon,
		"photos":   p.Photos,
		"lat":      p.Lat,
		"lng":      p.Lng,
		"vicinity": p.Vicinity,
	}
}

type Query struct {
	Lat     float64
	Lng     float64
	Journey string
	Radius  int
	Limit   int
}

func (q *Query) Run() []interface{} {
	var w sync.WaitGroup
	var m sync.Mutex
	places := make([]interface{}, q.Limit)
	for i := range q.Limit {
		w.Add(1)
		go func(journey string, i int) {
			defer w.Done()
			resp, err := q.FindNearyBy(strings.Join(Journeys[CapitalizeFirstLetter(journey)], ","))
			if err != nil {
				log.Println("Failed to find places:", err)
				return
			}
			if len(resp) == 0 {
				log.Println("No places found for ", Journeys[CapitalizeFirstLetter(journey)])
				return
			}
			randI := rand.IntN(len(resp))
			m.Lock()
			places[i] = resp[randI]
			m.Unlock()
		}(q.Journey, i)
	}
	w.Wait()
	return places
}

func (q *Query) FindNearyBy(types string) ([]DetailResponse, error) {
	u := config.ConfigVars.BaseURL + "/nearbysearch"
	vals := make(url.Values)
	vals.Set("location", fmt.Sprintf("%g,%g", q.Lat, q.Lng))
	vals.Set("radius", fmt.Sprintf("%d", q.Radius))
	vals.Set("types", types)
	vals.Set("api_key", config.ConfigVars.MAPApiKey)
	res, err := http.Get(u + "?" + vals.Encode())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nearby search failed with status: %d", res.StatusCode)
	}

	var response nearbyResponse
	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	ch := make(chan detailResult, len(response.Predictions))

	for _, v := range response.Predictions {
		wg.Add(1)
		go func(placeID string) {
			defer wg.Done()
			ch <- findDetail(placeID)
		}(v.PlaceID)
	}

	wg.Wait()
	close(ch)

	var finalData []DetailResponse
	for result := range ch {
		if result.err == nil {
			finalData = append(finalData, result.response)
		}
	}
	return finalData, nil
}

func findDetail(placeID string) detailResult {
	u := config.ConfigVars.BaseURL + "/details"
	vals := make(url.Values)
	vals.Set("place_id", placeID)
	vals.Set("api_key", config.ConfigVars.MAPApiKey)
	resp, err := http.Get(u + "?" + vals.Encode())
	if err != nil {
		return detailResult{
			err: err,
		}
	}
	defer resp.Body.Close()
	var dR DetailResponse
	if err := json.NewDecoder(resp.Body).Decode(&dR); err != nil {
		return detailResult{err: err}
	}
	return detailResult{
		response: dR,
		err:      nil,
	}
}

func CapitalizeFirstLetter(input string) string {
	words := strings.Fields(input)
	for i, word := range words {
		words[i] = string(unicode.ToUpper(rune(word[0]))) + word[1:]
	}
	return strings.Join(words, " ")
}

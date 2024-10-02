package services

import (
	"encoding/json"
	"fmt"
	"gltsm/models"
	"io"
	"net/http"
	"strconv"
)

type LastFmService struct {
	User   string
	ApiKey string
	Limit  int
}

func (s *LastFmService) FetchScrobbles(page *int) models.ScrobbleResponse {
	url := fmt.Sprintf("http://ws.audioscrobbler.com/2.0/?method=user.getrecenttracks&user=%s&api_key=%s&format=json", s.User, s.ApiKey)
	method := "GET"

	current_page := 1
	if page != nil {
		current_page = *page
		url += "&page=" + strconv.Itoa(current_page)
	}

	var scrobbles models.ScrobbleResponse

	client := &http.Client{}

	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return models.ScrobbleResponse{
			Err: err,
		}
	}

	response, err := client.Do(req)

	if err != nil {
		return models.ScrobbleResponse{
			Err: err,
		}
	}

	if response.StatusCode != 200 {
		return models.ScrobbleResponse{
			Err: fmt.Errorf("request was not fulfilled, received status code %d", response.StatusCode),
		}
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)

	if err != nil {
		return models.ScrobbleResponse{
			Err: err,
		}
	}

	err = json.Unmarshal([]byte(body), &scrobbles)

	if err != nil {
		return models.ScrobbleResponse{
			Err: err,
		}
	}

	return scrobbles
}

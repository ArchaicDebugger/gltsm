package services

import (
	"gltsm/models"
)

type LastFmFetcher interface {
	FetchScrobbles(page *int, results chan<- models.ScrobbleResponse)
}

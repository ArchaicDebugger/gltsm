package services

import (
	dbmodels "gltsm/models/db"
	"time"
)

type DbOperationResponse struct {
	WasAdded bool
	Id       string
}

type DbWriter interface {
	AddAlbumIfNotExists(album *dbmodels.Album) (DbOperationResponse, error)
	AddTrackIfNotExists(track *dbmodels.Track) (DbOperationResponse, error)
	AddArtistIfNotExists(artist *dbmodels.Artist) (DbOperationResponse, error)
	AddUserIfNotExists(username string) (DbOperationResponse, error)
	AddListeningHistoryIfNotExists(userId string, trackId string, date time.Time) (DbOperationResponse, error)
}

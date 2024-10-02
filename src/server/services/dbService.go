package services

import (
	dbmodels "gltsm/models/db"
	"time"
)

type DbService struct {
}

func (s *DbService) AddAlbumIfNotExists(album *dbmodels.Album) (DbOperationResponse, error) {
	conn := GetGormConnection()
	db, err := conn.DB()
	if err != nil {
		return DbOperationResponse{}, err
	}

	defer db.Close()

	result := DbOperationResponse{}

	var existing_records int64 = 0
	err = conn.Model(&dbmodels.Album{}).Where(&dbmodels.Album{Mbid: album.Mbid}).Count(&existing_records).Error
	if err != nil {
		return result, err
	}

	if existing_records > 0 {
		result.WasAdded = false
	} else {
		conn.Create(&album)

		result = DbOperationResponse{
			WasAdded: true,
			Id:       album.Mbid,
		}
	}

	return result, nil
}

func (s *DbService) AddTrackIfNotExists(track *dbmodels.Track) (DbOperationResponse, error) {
	conn := GetGormConnection()
	db, err := conn.DB()
	if err != nil {
		return DbOperationResponse{}, err
	}

	defer db.Close()

	result := DbOperationResponse{}

	var existing_records int64 = 0
	err = conn.Model(&dbmodels.Track{}).Where(&dbmodels.Track{Mbid: track.Mbid}).Count(&existing_records).Error
	if err != nil {
		return result, err
	}

	if existing_records > 0 {
		result.WasAdded = false
	} else {
		conn.Create(&track)

		result = DbOperationResponse{
			WasAdded: true,
			Id:       track.Mbid,
		}
	}

	return result, nil
}

func (s *DbService) AddArtistIfNotExists(artist *dbmodels.Artist) (DbOperationResponse, error) {
	conn := GetGormConnection()
	db, err := conn.DB()
	if err != nil {
		return DbOperationResponse{}, err
	}

	defer db.Close()

	result := DbOperationResponse{}

	var existing_records int64 = 0
	err = conn.Model(&dbmodels.Artist{}).Where(&dbmodels.Artist{Mbid: artist.Mbid}).Count(&existing_records).Error
	if err != nil {
		return result, err
	}

	if existing_records > 0 {
		result.WasAdded = false
	} else {
		conn.Create(&artist)

		result = DbOperationResponse{
			WasAdded: true,
			Id:       artist.Mbid,
		}
	}

	return result, nil
}

func (s *DbService) AddUserIfNotExists(username string) (DbOperationResponse, error) {
	conn := GetGormConnection()
	db, err := conn.DB()
	if err != nil {
		return DbOperationResponse{}, err
	}

	defer db.Close()

	result := DbOperationResponse{}

	var existing_records int64 = 0
	err = conn.Model(&dbmodels.User{}).Where(&dbmodels.User{Username: username}).Count(&existing_records).Error
	if err != nil {
		return result, err
	}

	if existing_records > 0 {
		result.WasAdded = false
	} else {
		new_record := dbmodels.User{Username: username, ID: username}
		conn.Create(&new_record)

		result = DbOperationResponse{
			WasAdded: true,
			Id:       username,
		}
	}

	return result, nil
}

func (s *DbService) AddListeningHistoryIfNotExists(userId string, trackId string, date time.Time) (DbOperationResponse, error) {
	conn := GetGormConnection()
	db, err := conn.DB()
	if err != nil {
		return DbOperationResponse{}, err
	}

	defer db.Close()

	result := DbOperationResponse{}

	var existing_records int64 = 0
	err = conn.Model(&dbmodels.ListeningHistory{}).Where(&dbmodels.ListeningHistory{UserID: userId, TrackID: trackId, Date: date}).Count(&existing_records).Error
	if err != nil {
		return result, err
	}

	if existing_records > 0 {
		result.WasAdded = false
	} else {
		new_record := dbmodels.ListeningHistory{UserID: userId, TrackID: trackId, Date: date}
		conn.Create(&new_record)

		result = DbOperationResponse{
			WasAdded: true,
		}
	}

	return result, nil
}

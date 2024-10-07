package services

import (
	"gltsm/models"
	dbmodels "gltsm/models/db"
	"sync"
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

	var existingRecords int64 = 0
	err = conn.Model(&dbmodels.Album{}).Where(&dbmodels.Album{Mbid: album.Mbid}).Count(&existingRecords).Error
	if err != nil {
		return result, err
	}

	if existingRecords > 0 {
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

	var existingRecords int64 = 0
	err = conn.Model(&dbmodels.Track{}).Where(&dbmodels.Track{Mbid: track.Mbid}).Count(&existingRecords).Error
	if err != nil {
		return result, err
	}

	if existingRecords > 0 {
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

	var existingRecords int64 = 0
	err = conn.Model(&dbmodels.Artist{}).Where(&dbmodels.Artist{Mbid: artist.Mbid}).Count(&existingRecords).Error
	if err != nil {
		return result, err
	}

	if existingRecords > 0 {
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

	var existingRecords int64 = 0
	err = conn.Model(&dbmodels.User{}).Where(&dbmodels.User{Username: username}).Count(&existingRecords).Error
	if err != nil {
		return result, err
	}

	if existingRecords > 0 {
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

	var existingRecords int64 = 0
	err = conn.Model(&dbmodels.ListeningHistory{}).Where(&dbmodels.ListeningHistory{UserID: userId, TrackID: trackId, Date: date}).Count(&existingRecords).Error
	if err != nil {
		return result, err
	}

	if existingRecords > 0 {
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

func (s *DbService) GetTracksThatFitTheMood(moodRanges []models.MoodRange) ([]dbmodels.Track, error) {
	conn := GetGormConnection()
	db, err := conn.DB()
	if err != nil {
		return nil, err
	}

	var allRanges []models.MoodRange
	for _, mr := range moodRanges {
		for i := 0; i < 10; i++ {
			allRanges = append(allRanges, models.MoodRange{
				StartTime: mr.StartTime.AddDate(-1, 0, 0),
				EndTime:   mr.EndTime.AddDate(-1, 0, 0),
			})
		}
	}

	defer db.Close()

	chunkedRanges := models.Chunk(allRanges, 50)
	var wg sync.WaitGroup
	ch := make(chan []dbmodels.Track, len(chunkedRanges))
	mapResult := make(map[string]dbmodels.Track)
	var result []dbmodels.Track

	for _, chunk := range chunkedRanges {
		wg.Add(1)
		go func(currRanges []models.MoodRange) {
			defer wg.Done()
			var entries []dbmodels.Track
			query := conn

			for _, tr := range chunk {
				query = query.Or("listening_histories.date BETWEEN ? AND ?", tr.StartTime, tr.EndTime)
			}

			err = conn.
				Preload("Album").Preload("Album.Artist").
				Table("tracks").
				Joins("JOIN listening_histories ON tracks.mbid = listening_histories.track_id").
				Where(query).
				Find(&entries).Error

			if err != nil {
				panic(err)
			}

			ch <- entries
		}(chunk)
	}

	wg.Wait()
	close(ch)

	for chunk := range ch {
		//result = append(result, chunk...)
		for _, item := range chunk {
			mapResult[item.Mbid] = item
		}
	}

	result = models.MapToArray(&mapResult)

	return result, nil
}

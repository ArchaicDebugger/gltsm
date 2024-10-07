package main

import (
	"fmt"
	"gltsm/models"
	dbmodels "gltsm/models/db"
	"gltsm/services"
	"strconv"
	"sync"
	"time"
)

type DbResultChannel chan services.DbOperationResponse

func getAllScrobbles(user string) models.ScrobbleResponse {
	var lfs services.LastFmFetcher
	//var dbs services.DbWriter
	results := make(chan models.ScrobbleResponse)

	lfs = &services.LastFmService{
		User:   user,
		ApiKey: services.EnvVariable("LASTFM_API_KEY"),
		Limit:  200,
	}

	firstCall := lfs.FetchScrobbles(nil)
	close(results)

	if firstCall.Err != nil {
		panic(firstCall.Err)
	}

	totalPages, err := strconv.Atoi(firstCall.Recenttracks.Attr.TotalPages)

	if err != nil {
		panic(err)
	}

	pagesChunk := 100

	for i := 1; i <= totalPages; i += pagesChunk {
		wg := sync.WaitGroup{}

		results = make(chan models.ScrobbleResponse, pagesChunk)
		last_page := i + pagesChunk - 1
		if last_page > totalPages {
			last_page = totalPages
		}

		for j := i; j <= last_page; j++ {
			wg.Add(1)
			go func(j int, results_ch chan<- models.ScrobbleResponse) {
				defer wg.Done()
				curr_result := lfs.FetchScrobbles(&j)
				results_ch <- curr_result
			}(j, results)
		}

		wg.Wait()
		close(results)

		percentageDone := float32(last_page) / float32(totalPages)
		percentageDone *= 100

		fmt.Printf("\rFetching scrobbles: %.2f%%", percentageDone)

		for successfulPage := range results {
			if successfulPage.Err != nil {
				fmt.Println("Error: ", successfulPage.Err)
			} else {
				firstCall.Recenttracks.Track = append(firstCall.Recenttracks.Track, successfulPage.Recenttracks.Track...)
			}
		}
		time.Sleep(CHUNK_WAIT_TIME)
	}

	fmt.Printf("\nFinished gettring scrobbles, collected %d items", len(firstCall.Recenttracks.Track))

	albumMap := make(map[string]dbmodels.Album)
	artistMap := make(map[string]dbmodels.Artist)
	trackMap := make(map[string]dbmodels.Track)

	for i := range firstCall.Recenttracks.Track {
		currentItem := firstCall.Recenttracks.Track[i]

		album_id := currentItem.Album.Mbid
		track_id := currentItem.Mbid
		artistId := currentItem.Artist.Mbid

		artistMap[artistId] = dbmodels.Artist{
			Mbid: currentItem.Artist.Mbid,
			Name: currentItem.Artist.Text,
		}

		albumMap[album_id] = dbmodels.Album{
			Mbid:     currentItem.Album.Mbid,
			ArtistID: currentItem.Artist.Mbid,
			Name:     currentItem.Album.Text,
		}

		trackMap[track_id] = dbmodels.Track{
			Mbid:     currentItem.Mbid,
			Name:     currentItem.Name,
			ArtistID: currentItem.Artist.Mbid,
			AlbumID:  currentItem.Album.Mbid,
		}
	}

	var dbs services.DbWriter = &services.DbService{}

	user_response, err := dbs.AddUserIfNotExists(user)

	if err != nil {
		panic(err)
	}

	var actionConsequence string
	if user_response.WasAdded {
		actionConsequence = "was added to"
	} else {
		actionConsequence = "was already in"
	}

	user_output_str := fmt.Sprintf("User %s %s DB", user, actionConsequence)
	fmt.Println(user_output_str)

	chunkSize := 30
	store(&artistMap, chunkSize, dbs.AddArtistIfNotExists, "artists")
	store(&albumMap, chunkSize, dbs.AddAlbumIfNotExists, "albums")
	store(&trackMap, chunkSize, dbs.AddTrackIfNotExists, "tracks")

	history_chunks := models.Chunk(firstCall.Recenttracks.Track, chunkSize)
	totalItemsAdded := 0

	for _, chunk := range history_chunks {
		wg := sync.WaitGroup{}
		dbch := make(DbResultChannel, len(chunk))
		for _, item := range chunk {
			wg.Add(1)
			go func() {
				defer wg.Done()
				uts, err := strconv.ParseInt(item.Date.Uts, 10, 64)
				if err != nil {
					panic(err)
				}
				item_time := time.Unix(uts, 0).UTC()
				response, err := dbs.AddListeningHistoryIfNotExists(user, item.Mbid, item_time)
				if err != nil {
					panic(err)
				}
				dbch <- response
			}()
		}
		wg.Wait()
		close(dbch)
		appendOperationsToTotal(&dbch, &totalItemsAdded)
	}

	fmt.Println("Finished seeding listening history, added total items:", totalItemsAdded)

	return firstCall
}

func appendOperationsToTotal(ch *DbResultChannel, totalCounts *int) {
	for response := range *ch {
		if response.WasAdded {
			*totalCounts++
		}
	}
}

func store[T any](m *map[string]T, chunkSize int, action func(*T) (services.DbOperationResponse, error), modelName string) {
	full_arr := models.MapToArray(m)
	itemChunks := models.Chunk(full_arr, chunkSize)
	total := 0

	for _, chunk := range itemChunks {
		wg := sync.WaitGroup{}
		dbch := make(DbResultChannel, len(chunk))
		for _, item := range chunk {
			wg.Add(1)
			go func(item *T) {
				defer wg.Done()
				response, err := action(item)
				if err != nil {
					panic(err)
				}
				dbch <- response
			}(&item)
		}

		wg.Wait()
		close(dbch)
		appendOperationsToTotal(&dbch, &total)
	}

	fmt.Printf("Finished seeding %s, added %d total items\n", modelName, total)
}

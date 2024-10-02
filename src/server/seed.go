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

	first_call := lfs.FetchScrobbles(nil)
	close(results)

	if first_call.Err != nil {
		panic(first_call.Err)
	}

	total_pages, err := strconv.Atoi(first_call.Recenttracks.Attr.TotalPages)

	if err != nil {
		panic(err)
	}

	pages_chunk := 100

	for i := 1; i <= total_pages; i += pages_chunk {
		wg := sync.WaitGroup{}

		results = make(chan models.ScrobbleResponse, pages_chunk)
		last_page := i + pages_chunk - 1
		if last_page > total_pages {
			last_page = total_pages
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

		percentage_done := float32(last_page) / float32(total_pages)
		percentage_done *= 100

		fmt.Printf("\rFetching scrobbles: %.2f%%", percentage_done)

		for succsessful_page := range results {
			if succsessful_page.Err != nil {
				fmt.Println("Error: ", succsessful_page.Err)
			} else {
				first_call.Recenttracks.Track = append(first_call.Recenttracks.Track, succsessful_page.Recenttracks.Track...)
			}
		}
		time.Sleep(CHUNK_WAIT_TIME)
	}

	fmt.Printf("\nFinished gettring scrobbles, collected %d items", len(first_call.Recenttracks.Track))

	album_map := make(map[string]dbmodels.Album)
	artist_map := make(map[string]dbmodels.Artist)
	track_map := make(map[string]dbmodels.Track)

	for i := range first_call.Recenttracks.Track {
		current_item := first_call.Recenttracks.Track[i]

		album_id := current_item.Album.Mbid
		track_id := current_item.Mbid
		artist_id := current_item.Artist.Mbid

		artist_map[artist_id] = dbmodels.Artist{
			Mbid: current_item.Artist.Mbid,
			Name: current_item.Artist.Text,
		}

		album_map[album_id] = dbmodels.Album{
			Mbid:     current_item.Album.Mbid,
			ArtistID: current_item.Artist.Mbid,
			Name:     current_item.Album.Text,
		}

		track_map[track_id] = dbmodels.Track{
			Mbid:     current_item.Mbid,
			Name:     current_item.Name,
			ArtistID: current_item.Artist.Mbid,
			AlbumID:  current_item.Album.Mbid,
		}
	}

	var dbs services.DbWriter = &services.DbService{}

	user_response, err := dbs.AddUserIfNotExists(user)

	if err != nil {
		panic(err)
	}

	var action_consequence string
	if user_response.WasAdded {
		action_consequence = "was added to"
	} else {
		action_consequence = "was already in"
	}

	user_output_str := fmt.Sprintf("User %s %s DB", user, action_consequence)
	fmt.Println(user_output_str)

	chunk_size := 30
	store(&artist_map, chunk_size, dbs.AddArtistIfNotExists, "artists")
	store(&album_map, chunk_size, dbs.AddAlbumIfNotExists, "albums")
	store(&track_map, chunk_size, dbs.AddTrackIfNotExists, "tracks")

	history_chunks := chunk(&first_call.Recenttracks.Track, chunk_size)
	total_items_added := 0

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
		appendOperationsToTotal(&dbch, &total_items_added)
	}

	fmt.Println("Finished seeding listening history, added total items:", total_items_added)

	return first_call
}

func appendOperationsToTotal(ch *DbResultChannel, total_counts *int) {
	for response := range *ch {
		if response.WasAdded {
			*total_counts++
		}
	}
}

func mapToArray[T any](m *map[string]T) []T {
	var result []T

	for _, value := range *m {
		result = append(result, value)
	}

	return result
}

func chunk[T any](m *[]T, chunk_size int) [][]T {
	l := len(*m)

	var result [][]T

	for i := 0; i < l; i += chunk_size {
		var slice []T
		for j := i; j < i+chunk_size && j < l; j++ {
			slice = append(slice, (*m)[j])
		}
		result = append(result, slice)
	}

	return result
}

func store[T any](m *map[string]T, chunk_size int, action func(*T) (services.DbOperationResponse, error), model_name string) {
	full_arr := mapToArray(m)
	item_chunks := chunk(&full_arr, chunk_size)
	total := 0

	for _, chunk := range item_chunks {
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

	fmt.Printf("Finished seeding %s, added %d total items\n", model_name, total)
}

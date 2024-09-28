package main

import (
	"fmt"
	"gltsm/models"
	"gltsm/services"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var CHUNK_WAIT_TIME time.Duration = 200 * time.Millisecond

func envVariable(key string) string {
	err := godotenv.Load(".env")

	if err != nil {
		panic(err)
	}

	return os.Getenv(key)
}

func main() {
	r := gin.Default()
	r.GET("/seed", func(c *gin.Context) {
		user := c.Query("user")

		if user == "" {
			c.JSON(400, gin.H{
				"error":   "Bad Request",
				"message": "User not provided",
			})
			return
		}

		scrobbles := getAllScrobbles(user)
		c.JSON(200, len(scrobbles.Recenttracks.Track))
	})
	r.Run("localhost:4500")
}

func getAllScrobbles(user string) models.ScrobbleResponse {
	var lfs services.LastFmFetcher
	results := make(chan models.ScrobbleResponse)

	lfs = &services.LastFmService{
		User:   user,
		ApiKey: envVariable("LASTFM_API_KEY"),
		Limit:  200,
	}

	go lfs.FetchScrobbles(nil, results)

	first_call := <-results
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
		results = make(chan models.ScrobbleResponse, pages_chunk)
		incoming_messages := 0
		last_page := i + pages_chunk - 1
		if last_page > total_pages {
			last_page = total_pages
		}

		for j := i; j <= last_page; j++ {
			go lfs.FetchScrobbles(&j, results)
		}

		for succsessful_page := range results {
			incoming_messages++

			if succsessful_page.Err != nil {
				fmt.Println("Error: ", succsessful_page.Err)
			} else {
				first_call.Recenttracks.Track = append(first_call.Recenttracks.Track, succsessful_page.Recenttracks.Track...)
			}

			if (i-1)+incoming_messages == last_page {
				close(results)
			}

			percentage_done := (float32(i) + float32(incoming_messages)) / float32(total_pages)
			percentage_done *= 100

			fmt.Printf("\rFetching scrobbles: %.2f%%", percentage_done)
		}
		time.Sleep(CHUNK_WAIT_TIME)
	}

	fmt.Printf("Finished gettring scrobbles, collected %d items", len(first_call.Recenttracks.Track))
	return first_call
}

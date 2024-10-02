package main

import (
	"gltsm/services"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var CHUNK_WAIT_TIME time.Duration = 200 * time.Millisecond

func main() {
	services.EnsureDbCreated()

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true, // Allow all origins
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true, // Allow credentials like cookies, authorization headers
	}))

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

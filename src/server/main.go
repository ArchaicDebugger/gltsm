package main

import (
	"gltsm/services"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var CHUNK_WAIT_TIME time.Duration = 200 * time.Millisecond

func main() {
	services.EnsureDbCreated()
	loadTimezoneCache()

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

	r.GET("/mood", func(c *gin.Context) {
		lat, lng := parseLatLng(c)
		timestr := c.Query("time")
		location := getLocation(lat, lng)
		//isUtc := false

		var currentTime time.Time

		if timestr == "" {
			currentTime = getLocalTime(lat, lng, time.Now().In(location))
		} else {
			var dateErr error
			currentTime, dateErr = time.Parse("2006-01-02 15:04", timestr)
			if dateErr != nil {
				c.JSON(400, gin.H{
					"error":   "Bad Request",
					"message": "Time should be of format yyyy-MM-dd hh:mm",
				})
			}
		}

		sun_angle := sunAltitude(lat, lng, currentTime)
		min_angle := sun_angle - (sun_angle / 4)
		max_angle := sun_angle + (sun_angle / 4)
		dayOfYear := time.Now().YearDay()
		moodRange := timeRangeForAngle(lat, lng, dayOfYear, min_angle, max_angle)
		c.JSON(200, gin.H{
			"sunAngle": sun_angle,
			"minDate":  moodRange.StartTime,
			"maxDate":  moodRange.EndTime,
		})
	})

	r.GET("/year-mood", func(c *gin.Context) {
		lat, lng := parseLatLng(c)
		location := getLocation(lat, lng)
		timestr := c.Query("time")
		var currentTime time.Time

		if timestr == "" {
			currentTime = getLocalTime(lat, lng, time.Now().In(location))
		} else {
			var dateErr error
			currentTime, dateErr = time.Parse("2006-01-02 15:04", timestr)
			if dateErr != nil {
				c.JSON(400, gin.H{
					"error":   "Bad Request",
					"message": "Time should be of format yyyy-MM-dd hh:mm",
				})
			}
		}

		sunAngle := sunAltitude(lat, lng, currentTime)
		minAngle := sunAngle - (sunAngle / 4)
		maxAngle := sunAngle + (sunAngle / 4)

		ranges := getYearTimeRangeForAngle(lat, lng, &currentTime, minAngle, maxAngle)

		c.JSON(200, ranges)
	})
	r.Run("localhost:4500")
}

func parseLatLng(c *gin.Context) (float64, float64) {
	latstr := c.Query("lat")
	lngstr := c.Query("lng")

	lat, err := strconv.ParseFloat(latstr, 64)
	if err != nil {
		c.JSON(400, gin.H{
			"error":   "Bad Request",
			"message": "Latitude is not a valid number",
		})
	}

	lng, err := strconv.ParseFloat(lngstr, 64)
	if err != nil {
		c.JSON(400, gin.H{
			"error":   "Bad Request",
			"message": "Latitude is not a valid number",
		})
	}

	return lat, lng
}

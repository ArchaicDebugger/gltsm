package main

import (
	"fmt"
	"math"
	"os"
	"sort"
	"sync"
	"time"

	timezone "github.com/evanoberholster/timezoneLookup/v2"
)

type MoodRange struct {
	Date      time.Time
	StartTime time.Time
	EndTime   time.Time
	HasValue  bool
}

// Constants
const (
	DegToRad = math.Pi / 180
	RadToDeg = 180 / math.Pi
)

// Helper functions
func sinDeg(angle float64) float64 {
	return math.Sin(angle * DegToRad)
}

func cosDeg(angle float64) float64 {
	return math.Cos(angle * DegToRad)
}

func calculateSolarDeclination(dayOfYear int) float64 {
	// Approximation for solar declination
	// δ = 23.44 * sin((360 / 365) * (dayOfYear - 81))
	return 23.44 * sinDeg((360/365.0)*(float64(dayOfYear)-81))
}

func calculateHourAngle(timeOfDay float64) float64 {
	// Local solar time angle: H = 15° per hour from solar noon
	return 15.0 * (timeOfDay - 12.0)
}

// sunAltitude calculates the Sun's altitude based on location and time
func sunAltitude(lat, lng float64, localTime time.Time) float64 {
	// Calculate the time of day in decimal hours, adjusted for longitude
	timeOffset := float64(lng) / 15.0 // Longitude to hour offset (15° = 1 hour)
	timeOfDay := float64(localTime.Hour()) + float64(localTime.Minute())/60.0 + timeOffset

	// Get the day of the year from the local time
	dayOfYear := localTime.YearDay()

	// Calculate solar declination for the given day
	solarDeclination := calculateSolarDeclination(dayOfYear)

	// Calculate hour angle
	hourAngle := calculateHourAngle(timeOfDay)

	// Calculate the Sun's altitude
	sinH := sinDeg(lat)*sinDeg(solarDeclination) + cosDeg(lat)*cosDeg(solarDeclination)*cosDeg(hourAngle)
	return math.Asin(sinH) * RadToDeg
}

func timeRangeForAngle(lat, lng float64, dayOfYear int, minAngle float64, maxAngle float64) MoodRange {
	fmt.Println("Fetching day:", dayOfYear)
	defer fmt.Println("Fetched day:", dayOfYear)
	location := getLocation(lat, lng)
	startOfDay := time.Date(time.Now().Year(), 1, dayOfYear, 0, 0, 0, 0, location)

	var firstTime, lastTime time.Time
	found := false

	var minute int

	for minute = 0; minute < 1440; minute++ {
		currentTime := startOfDay.Add(time.Duration(minute) * time.Minute)

		altitude := sunAltitude(lat, lng, currentTime)

		if altitude >= minAngle {
			firstTime = currentTime
			found = true
			break
		}
	}

	for ; minute < 1440; minute++ {
		currentTime := startOfDay.Add(time.Duration(minute) * time.Minute)

		altitude := sunAltitude(lat, lng, currentTime)

		if altitude >= maxAngle || altitude <= minAngle {
			break
		}

		lastTime = currentTime
	}

	return MoodRange{
		Date:      startOfDay,
		StartTime: firstTime,
		EndTime:   lastTime,
		HasValue:  found,
	}
}

func getYearTimeRangeForAngle(lat, lng float64, currentTime *time.Time, minAngle float64, maxAngle float64) []MoodRange {
	startOfYear := time.Date(currentTime.Year(), 1, 1, 0, 0, 0, 0, currentTime.Location())
	startOfNextYear := startOfYear.AddDate(1, 0, 0)
	endOfYear := startOfNextYear.AddDate(0, 0, -1)

	var wg = sync.WaitGroup{}
	ch := make(chan MoodRange, endOfYear.YearDay())

	for currentDate := startOfYear; currentDate.Before(endOfYear); currentDate = currentDate.AddDate(0, 0, 1) {
		wg.Add(1)
		go func(curr *time.Time) {
			defer wg.Done()
			ch <- timeRangeForAngle(lat, lng, currentDate.YearDay(), minAngle, maxAngle)
		}(&currentDate)
	}

	wg.Wait()
	close(ch)

	var ranges []MoodRange
	for day := range ch {
		if day.HasValue {
			ranges = append(ranges, day)
		}
	}

	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i].Date.Before(ranges[j].Date)
	})

	return ranges
}

var TimezoneCache timezone.Timezonecache
var tzMutex sync.Mutex

func loadTimezoneCache() {
	f, err := os.Open("timezone.data")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err = TimezoneCache.Load(f); err != nil {
		panic(err)
	}
}

func getLocation(lat float64, lng float64) *time.Location {
	tzMutex.Lock()
	defer tzMutex.Unlock()
	tzname, err := TimezoneCache.Search(lat, lng)
	if err != nil {
		panic(err)
	}

	location, err := time.LoadLocation(tzname.Name)
	if err != nil {
		panic(err)
	}

	return location
}

func getLocalTime(lat float64, lng float64, utcTime time.Time) time.Time {
	location := getLocation(lat, lng)
	return utcTime.In(location)
}

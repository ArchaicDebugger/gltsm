package services

import (
	"gltsm/models"
	dbmodels "gltsm/models/db"
)

type DbReader interface {
	GetTracksThatFitTheMood(mood []models.MoodRange) ([]dbmodels.Track, error)
}

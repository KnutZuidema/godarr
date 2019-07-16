package model

type ItemKind string

const (
	ItemKindMovie    ItemKind = "movie"
	ItemKindTVSeries          = "tv-series"
)

type ItemStatus string

const (
	ItemStatusAdded      ItemStatus = "added"
	ItemStatusMonitored             = "monitored"
	ItemStatusDownloaded            = "downloaded"
)

type Item struct {
	ID             string      `json:"id" db:"id"`
	ExternalID     string      `json:"externalId" db:"external_id"`
	Kind           ItemKind    `json:"kind" db:"kind"`
	Title          string      `json:"title" db:"title"`
	Description    string      `json:"description" db:"description"`
	ImagePath      string      `json:"imagePath" db:"image_path"`
	ReleaseYear    int         `json:"releaseYear" db:"release_year"`
	Genres         []string    `json:"genres" db:"genres"`
	Rating         float64     `json:"rating" db:"rating"`
	Status         ItemStatus  `json:"status" db:"status"`
	AdditionalData interface{} `json:"data"`
}

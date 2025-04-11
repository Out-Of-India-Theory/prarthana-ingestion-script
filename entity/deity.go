package entity

import "time"

type DeityDocument struct {
	TmpId          string              `bson:"TmpId"`
	Id             string              `bson:"_id" json:"_id"`
	Order          int                 `json:"order" bson:"order"`
	Title          map[string]string   `json:"title" bson:"title,omitempty"`
	Slug           string              `json:"slug" bson:"slug"`
	Aliases        []string            `json:"aliases" bson:"aliases,omitempty"`
	SearchKeywords []string            `json:"search_keywords" bson:"search_keywords"`
	Description    map[string]string   `json:"description" bson:"description"`
	UIInfo         DeityUIInfo         `json:"ui_info" bson:"ui_info"`
	Prarthanas     []string            `bson:"prarthanas"`
	FestivalIds    []string            `json:"festival_ids" bson:"festival_ids"`
	Region         []string            `json:"region" bson:"region"`
	AliasesV1      map[string][]string `json:"aliases_v1" bson:"aliases_v1"`
	CreatedAt      time.Time           `bson:"created_at"`
	UpdatedAt      time.Time           `bson:"updated_at"`
	Status         bool                `bson:"status"`
}

type DeityUIInfo struct {
	DefaultImage    string           `json:"default_image" bson:"default_image"`
	BackgroundImage string           `json:"background_image" bson:"background_image"`
	DeityOfTheDay   string           `json:"deity_of_the_day" bson:"deity_of_the_day,omitempty"`
	HeroImageAlbum  []HeroImageAlbum `json:"hero_image_album" bson:"hero_image_album"`
}

type HeroImageAlbum struct {
	FullImage      string `json:"full_image" bson:"full_image"`
	ThumbnailImage string `json:"thumbnail_image" bson:"thumbnail_image"`
	ShareImage     string `json:"share_image" bson:"share_image"`
}

type DeitySearchData struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Aliases  []string `json:"aliases"`
	ImageURL string   `json:"image_url"`
}

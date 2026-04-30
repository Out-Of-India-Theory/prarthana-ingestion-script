package entity

import "time"

type Chapter struct {
	Order         int               `bson:"order"`
	Timestamp     string            `bson:"timestamp"`
	Duration      string            `bson:"duration" `
	Title         map[string]string `bson:"title" `
	StotraIds     []string          `bson:"stotra_ids"`
	DurationInSec int               `bson:"-"`
}

type Variant struct {
	Duration  string    `bson:"duration" json:"duration"`
	Chapters  []Chapter `bson:"chapters" json:"chapters"`
	IsDefault bool      `bson:"is_default" json:"is_default"`
	AudioInfo AudioInfo `bson:"audio_info" json:"audio_info"`
	VariantTitle map[string]string `bson:"variant_title" json:"variant_title"`
}

type Prarthana struct {
	TmpId              string              `bson:"TmpId"`
	Id                 string              `bson:"_id"`
	Title              map[string]string   `bson:"title"`
	FestivalIds        []string            `bson:"festival_ids"`
	AudioInfo          AudioInfo           `bson:"audio_info"`
	Days               []int               `bson:"days" `
	Description        map[string]string   `bson:"description" `
	Importance         map[string]string   `bson:"importance"`
	Variants           []Variant           `bson:"variants" `
	Instruction        map[string]string   `bson:"instruction" `
	ItemsRequired      map[string][]string `bson:"items_required" `
	DeityIds           []string            `bson:"deity_ids"`
	UiInfo             PrarthanaUIInfo     `bson:"ui_info"`
	AvailableLanguages []KeyValue          `bson:"available_languages"`
	IntentBased        bool                `bson:"intent_based"`
	CreatedAt          time.Time           `bson:"created_at"`
	UpdatedAt          time.Time           `bson:"updated_at"`
}

type AudioInfo struct {
	IsAudioAvailable bool   `json:"is_audio_available" bson:"is_audio_available"`
	AudioUrl         string `json:"audio_url" bson:"audio_url"`
	IsStudioRecorded bool   `json:"is_studio_recorded" bson:"is_studio_recorded"`
}

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type PrarthanaUIInfo struct {
	AlbumArt        string `json:"album_art" bson:"album_art"`
	DefaultImageUrl string `json:"default_image_url" bson:"default_image_url"`
	TemplateNumber  string `json:"template_number" bson:"template_number"`
	BannerImageUrl  string `json:"banner_image_url" bson:"banner_image_url"`
}

type PrarthanaSearchData struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Duration         string   `json:"duration"`
	DeityNames       []string `json:"deity_names"`
	Shloks           []string `json:"shloks"`
	ImageURL         string   `json:"image_url"`
	IsAudioAvailable bool     `json:"is_audio_available"`
	CategoryNames    []string `json:"category_names"`
}

type PrarthanaSearchDoc struct {
	ID             string               `bson:"_id"`
	TmpID          string               `bson:"TmpId"`
	Title          map[string]string    `bson:"title"`
	Deity          []DeityDocument      `bson:"deity"`
	Variants       []Variant            `bson:"variants"`
	StotraDocs     []Stotra             `bson:"stotra_docs"`
	ShlokDocs      []Shlok              `bson:"shlok_docs"`
	UIDetails      PrarthanaUIInfo      `bson:"ui_info"`
	AudioInfo      AudioInfo            `json:"audio_info" bson:"audio_info"`
	CollectionName []CollectionDocument `bson:"collection_name"`
}

type CollectionDocument struct {
	ID        string            `bson:"_id"`
	Name      map[string]string `bson:"name"`
	Slug      string            `bson:"slug"`
	Key       string            `bson:"key"`
	Thumbnail string            `bson:"thumbnail"`
	SubTitle  map[string]string `bson:"sub_title"`
	Title     map[string]string `bson:"title"`
	Status    bool              `bson:"status"`
}

type PoojaMongoDocument struct {
	ID           string               `bson:"_id"`
	Title        map[string]string    `bson:"title"`
	Key          string               `bson:"key"`
	Deities      []DeityDocument      `bson:"deities"`
	Variants     []Variant            `bson:"variants"`
	ThumbnailUrl string               `bson:"thumbnail_image_url"`
	Price        string               `bson:"price"`
	Collections  []CollectionDocument `bson:"collections"`
}

type PoojaESDocument struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Key            string   `json:"key"`
	DeityNames     []string `json:"deity_names"`
	Price          int      `json:"price"`
	ThumbnailUrl   string   `json:"thumbnail_url"`
	SearchKeywords []string `json:"search_keywords"`
}

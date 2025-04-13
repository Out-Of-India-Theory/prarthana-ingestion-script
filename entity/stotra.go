package entity

import "time"

type Stotra struct {
	ID                     string            `bson:"_id"`
	IntId                  int               `bson:"int_id"`
	Title                  map[string]string `bson:"title"`
	ShlokIds               []string          `bson:"shlok_ids"`
	StotraUrl              string            `bson:"stotra_url"`
	Duration               string            `bson:"duration"`
	DurationInSeconds      int               `bson:"duration_in_seconds"`
	DurationInMilliseconds int               `bson:"duration_in_milliseconds"`
	CreatedAt              time.Time         `bson:"created_at"`
	UpdatedAt              time.Time         `bson:"updated_at"`
}

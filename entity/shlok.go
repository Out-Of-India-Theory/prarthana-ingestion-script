package entity

import (
	"time"
)

type Shlok struct {
	ID          string            `bson:"_id"`
	IntId       int               `bson:"int_id"`
	Title       map[string]string `bson:"title"`
	Explanation map[string]string `bson:"explanation"`
	Shlok       map[string]string `bson:"shlok"`
	CreatedAt   time.Time         `bson:"created_at"`
	UpdatedAt   time.Time         `bson:"updated_at"`
}

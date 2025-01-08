package entity

type DeityDocument struct {
	TmpId          string
	Id             string            `bson:"_id,omitempty" json:"_id"`
	Order          int               `json:"order" bson:"order"`
	Title          map[string]string `json:"title" bson:"title"`
	Slug           string            `json:"slug" bson:"slug"`
	Aliases        []string          `json:"aliases" bson:"aliases"`
	SearchKeywords []string          `json:"search_keywords" bson:"search_keywords"`
	Description    map[string]string `json:"description" bson:"description"`
	UIInfo         DeityUIInfo       `json:"ui_info" bson:"ui_info"`
	Prarthanas     []string          `json:"prarthanas"`
}

type DeityUIInfo struct {
	DefaultImage    string `json:"default_image" bson:"default_image"`
	BackgroundImage string `json:"background_image"`
}

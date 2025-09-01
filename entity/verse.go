package entity

type Verse struct {
	ID                int               `bson:"_id"`
	Date              string            `bson:"date"`
	ScriptureName     map[string]string `bson:"scripture_name"`
	Chapter           int64             `bson:"chapter"`
	VerseNumber       int64             `bson:"verse_number"`
	Verse             map[string]string `bson:"verse"`
	SimplifiedMeaning map[string]string `bson:"simplified_meaning"`
	Lesson            map[string]string `bson:"lesson"`
}

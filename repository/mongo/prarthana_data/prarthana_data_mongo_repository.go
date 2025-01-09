package prarthana_data

import (
	"context"
	"fmt"
	"github.com/Out-Of-India-Theory/oit-go-commons/logging"
	mongoCommons "github.com/Out-Of-India-Theory/oit-go-commons/mongo"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/configuration"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"log"
)

const (
	prarthana_collection = "prarthanas"
	deity_collection     = "deities"
	shlok_collection     = "shloks"
	stotra_collection    = "stotras"
)

type PrarthanaDataMongoRepository struct {
	logger              *zap.Logger
	prarthanaCollection *mongo.Collection
	deityCollection     *mongo.Collection
	shlokCollection     *mongo.Collection
	stotraCollection    *mongo.Collection
}

func InitPrarthanaDataMongoRepository(ctx context.Context, config configuration.Configuration) *PrarthanaDataMongoRepository {
	mongoClient := mongoCommons.InitMongoClient(ctx, config.MongoConfig)
	return &PrarthanaDataMongoRepository{
		logger:              logging.WithContext(ctx),
		prarthanaCollection: mongoClient.Database(config.MongoConfig.Database).Collection(prarthana_collection),
		deityCollection:     mongoClient.Database(config.MongoConfig.Database).Collection(deity_collection),
		shlokCollection:     mongoClient.Database(config.MongoConfig.Database).Collection(shlok_collection),
		stotraCollection:    mongoClient.Database(config.MongoConfig.Database).Collection(stotra_collection),
	}
}

func (r *PrarthanaDataMongoRepository) InsertManyShloks(ctx context.Context, shloks []entity.Shlok) error {
	var documents []interface{}
	for _, shlok := range shloks {
		documents = append(documents, shlok)
	}

	result, err := r.shlokCollection.InsertMany(ctx, documents)
	if err != nil {
		return fmt.Errorf("error inserting documents: %w", err)
	}

	log.Printf("Inserted %d documents with IDs: %v\n", len(result.InsertedIDs), result.InsertedIDs)
	return nil
}

func (r *PrarthanaDataMongoRepository) InsertManyStotras(ctx context.Context, stotras []entity.Stotra) error {
	var documents []interface{}
	for _, stotra := range stotras {
		documents = append(documents, stotra)
	}

	result, err := r.stotraCollection.InsertMany(ctx, documents)
	if err != nil {
		return fmt.Errorf("error inserting documents: %w", err)
	}

	log.Printf("Inserted %d documents with IDs: %v\n", len(result.InsertedIDs), result.InsertedIDs)
	return nil
}

func (r *PrarthanaDataMongoRepository) InsertManyDeities(ctx context.Context, deities []entity.DeityDocument) error {
	var documents []interface{}
	for _, deity := range deities {
		documents = append(documents, deity)
	}

	result, err := r.deityCollection.InsertMany(ctx, documents)
	if err != nil {
		return fmt.Errorf("error inserting documents: %w", err)
	}

	log.Printf("Inserted %d documents with IDs: %v\n", len(result.InsertedIDs), result.InsertedIDs)
	return nil
}

func (r *PrarthanaDataMongoRepository) InsertManyPrarthanas(ctx context.Context, prarthanas []entity.Prarthana) error {
	var documents []interface{}
	for _, prarthana := range prarthanas {
		documents = append(documents, prarthana)
	}

	result, err := r.prarthanaCollection.InsertMany(ctx, documents)
	if err != nil {
		return fmt.Errorf("error inserting documents: %w", err)
	}

	log.Printf("Inserted %d documents with IDs: %v\n", len(result.InsertedIDs), result.InsertedIDs)
	return nil
}

func (r *PrarthanaDataMongoRepository) GetTmpIdToPrarthanaIds(ctx context.Context) (map[string]string, map[string]string, error) {
	filter := bson.M{}
	projection := bson.M{
		"_id":                     1,
		"TmpId":                   1,
		"ui_info.template_number": 1,
	}
	cursor, err := r.prarthanaCollection.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, nil, fmt.Errorf("error querying the collection: %w", err)
	}
	defer cursor.Close(ctx)

	idTemplateMap := make(map[string]string)
	tmpIdToIdMap := make(map[string]string)

	for cursor.Next(ctx) {
		var result struct {
			ID     string `bson:"_id"`
			TmpId  string `bson:"TmpId"`
			UiInfo struct {
				TemplateNumber string `bson:"template_number"`
			} `bson:"ui_info"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, nil, fmt.Errorf("error decoding document: %w", err)
		}
		idTemplateMap[result.TmpId] = result.UiInfo.TemplateNumber
		tmpIdToIdMap[result.TmpId] = result.ID
	}

	if err := cursor.Err(); err != nil {
		return nil, nil, fmt.Errorf("cursor iteration error: %w", err)
	}
	return idTemplateMap, tmpIdToIdMap, nil
}

func (r *PrarthanaDataMongoRepository) GetTmpIdToDeityIdMap(ctx context.Context) (map[string]string, error) {
	filter := bson.M{}
	projection := bson.M{
		"_id":                     1,
		"TmpId":                   1,
		"ui_info.template_number": 1,
	}

	cursor, err := r.deityCollection.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, fmt.Errorf("error querying the collection: %w", err)
	}
	defer cursor.Close(ctx)

	tmpIdToDeityIdMap := make(map[string]string)
	for cursor.Next(ctx) {
		var result struct {
			ID     string `bson:"_id"`
			TmpId  string `bson:"TmpId"`
			UiInfo struct {
				TemplateNumber string `bson:"template_number"`
			} `bson:"ui_info"`
		}

		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("error decoding document: %w", err)
		}
		tmpIdToDeityIdMap[result.TmpId] = result.ID
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor iteration error: %w", err)
	}
	return tmpIdToDeityIdMap, nil
}

func (r *PrarthanaDataMongoRepository) GetAllStotras(ctx context.Context) (map[string]entity.Stotra, error) {
	cursor, err := r.stotraCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("error fetching stotras: %w", err)
	}
	defer cursor.Close(ctx)
	stotraMap := make(map[string]entity.Stotra)
	for cursor.Next(ctx) {
		var stotra entity.Stotra
		if err := cursor.Decode(&stotra); err != nil {
			return nil, fmt.Errorf("error decoding stotra: %w", err)
		}
		stotraMap[stotra.ID] = stotra
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}
	return stotraMap, nil
}

func (r *PrarthanaDataMongoRepository) GetAllDeities(ctx context.Context) ([]entity.DeityDocument, error) {
	cursor, err := r.deityCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var deities []entity.DeityDocument
	if err = cursor.All(ctx, &deities); err != nil {
		return nil, err
	}

	return deities, nil
}

func (r *PrarthanaDataMongoRepository) GeneratePrarthanaTmpIdToIdMap(ctx context.Context) (map[string]string, error) {
	// Define the map to store the TmpId -> _id mapping
	tmpIdToIdMap := make(map[string]string)
	projection := bson.M{
		"_id":   1,
		"TmpId": 1,
	}
	cursor, err := r.prarthanaCollection.Find(ctx, bson.M{}, options.Find().SetProjection(projection))
	if err != nil {
		return nil, fmt.Errorf("error fetching documents: %w", err)
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var document struct {
			ID    string `bson:"_id"`
			TmpId string `bson:"TmpId"`
		}

		if err := cursor.Decode(&document); err != nil {
			return nil, fmt.Errorf("error decoding document: %w", err)
		}
		if document.TmpId != "" {
			tmpIdToIdMap[document.TmpId] = document.ID
		}
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return tmpIdToIdMap, nil
}

func (r *PrarthanaDataMongoRepository) GenerateDeityTmpIdToIdMap(ctx context.Context) (map[string]string, error) {
	// Map to store TmpId to _id mapping
	tmpIdToIdMap := make(map[string]string)
	cursor, err := r.deityCollection.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"_id": 1, "TmpId": 1}))
	if err != nil {
		return nil, fmt.Errorf("error fetching documents from MongoDB: %w", err)
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var document struct {
			ID    string `bson:"_id"`
			TmpId string `bson:"TmpId"`
		}
		if err := cursor.Decode(&document); err != nil {
			return nil, fmt.Errorf("error decoding document: %w", err)
		}
		tmpIdToIdMap[document.TmpId] = document.ID
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over cursor: %w", err)
	}
	return tmpIdToIdMap, nil
}

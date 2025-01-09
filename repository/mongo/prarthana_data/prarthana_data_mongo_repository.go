package prarthana_data

import (
	"context"
	"fmt"
	"github.com/Out-Of-India-Theory/oit-go-commons/logging"
	mongoCommons "github.com/Out-Of-India-Theory/oit-go-commons/mongo"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/configuration"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/entity"
	"github.com/google/uuid"
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
	if shloks == nil || len(shloks) == 0 {
		log.Println("No shloks provided for insertion.")
		return nil
	}

	for _, shlok := range shloks {
		result := r.shlokCollection.FindOneAndReplace(ctx, bson.M{"_id": shlok.ID}, shlok)
		if result.Err() != nil {
			if result.Err() == mongo.ErrNoDocuments {
				log.Printf("No existing document found for ID: %v. Inserting new shlok.\n", shlok.ID)
				_, err := r.shlokCollection.InsertOne(ctx, shlok)
				if err != nil {
					log.Printf("Failed to insert shlok with ID: %v. Error: %v\n", shlok.ID, err)
					return fmt.Errorf("failed to insert shlok with ID %v: %w", shlok.ID, err)
				}
				log.Printf("Successfully inserted new shlok with ID: %v.\n", shlok.ID)
			} else {
				log.Printf("Failed to find and replace shlok with ID: %v. Error: %v\n", shlok.ID, result.Err())
				return fmt.Errorf("failed to find and replace shlok with ID %v: %w", shlok.ID, result.Err())
			}
		} else {
			log.Printf("Successfully updated shlok with ID: %v.\n", shlok.ID)
		}
	}
	return nil
}

func (r *PrarthanaDataMongoRepository) InsertManyStotras(ctx context.Context, stotras []entity.Stotra) error {
	if stotras == nil || len(stotras) == 0 {
		log.Println("No stotras provided for insertion.")
		return nil
	}

	for _, stotra := range stotras {
		result := r.stotraCollection.FindOneAndReplace(ctx, bson.M{"_id": stotra.ID}, stotra)
		if result.Err() != nil {
			if result.Err() == mongo.ErrNoDocuments {
				log.Printf("No existing document found for ID: %v. Inserting new stotras.\n", stotra.ID)
				_, err := r.stotraCollection.InsertOne(ctx, stotra)
				if err != nil {
					log.Printf("Failed to insert stotras with ID: %v. Error: %v\n", stotra.ID, err)
					return fmt.Errorf("failed to insert stotras with ID %v: %w", stotra.ID, err)
				}
				log.Printf("Successfully inserted new stotras with ID: %v.\n", stotra.ID)
			} else {
				log.Printf("Failed to find and replace stotras with ID: %v. Error: %v\n", stotra.ID, result.Err())
				return fmt.Errorf("failed to find and replace stotras with ID %v: %w", stotra.ID, result.Err())
			}
		} else {
			log.Printf("Successfully updated stotras with ID: %v.\n", stotra.ID)
		}
	}
	return nil
	//var documents []interface{}
	//for _, stotra := range stotras {
	//	documents = append(documents, stotra)
	//}
	//
	//result, err := r.stotraCollection.InsertMany(ctx, documents)
	//if err != nil {
	//	return fmt.Errorf("error inserting documents: %w", err)
	//}
	//
	//log.Printf("Inserted %d documents with IDs: %v\n", len(result.InsertedIDs), result.InsertedIDs)
	//return nil
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
	if prarthanas == nil || len(prarthanas) == 0 {
		log.Println("No prarthana provided for insertion.")
		return nil
	}

	for _, prarthana := range prarthanas {
		result := r.prarthanaCollection.FindOne(ctx, bson.M{"tmp_id": prarthana.TmpId})
		if result.Err() != nil {
			if result.Err() == mongo.ErrNoDocuments {
				log.Printf("No existing document found for ID: %v. Inserting new prarthana.\n", prarthana.Id)
				prarthana.Id = uuid.NewString()
				_, err := r.prarthanaCollection.InsertOne(ctx, prarthana)
				if err != nil {
					log.Printf("Failed to insert prarthana with ID: %v. Error: %v\n", prarthana.Id, err)
					return fmt.Errorf("failed to insert prarthana with ID %v: %w", prarthana.Id, err)
				}
				log.Printf("Successfully inserted new prarthana with ID: %v.\n", prarthana.Id)
			} else {
				log.Printf("Failed to find and replace prarthana with ID: %v. Error: %v\n", prarthana.Id, result.Err())
				return fmt.Errorf("failed to find and replace prarthana with ID %v: %w", prarthana.Id, result.Err())
			}
		} else {
			var prarthanaDoc entity.Prarthana
			err := result.Decode(&prarthanaDoc)
			if err != nil {
				return fmt.Errorf("error decoding prarthana document: %w", err)
			}
			prarthana.Id = prarthanaDoc.Id
			updateResult, err := r.prarthanaCollection.ReplaceOne(ctx, bson.M{"tmp_id": prarthana.TmpId}, prarthana)
			if err != nil {
				return fmt.Errorf("error updating prarthana with ID %v: %w", prarthana.Id, err)
			}
			if updateResult.MatchedCount == 0 {
				return fmt.Errorf("error updating prarthana with ID %v: %w", prarthana.Id, err)
			}
			log.Printf("Successfully updated prarthana with ID: %v.\n", prarthana.Id)
		}
	}
	return nil

	//var documents []interface{}
	//for _, prarthana := range prarthanas {
	//	documents = append(documents, prarthana)
	//}
	//
	//result, err := r.prarthanaCollection.InsertMany(ctx, documents)
	//if err != nil {
	//	return fmt.Errorf("error inserting documents: %w", err)
	//}
	//
	//log.Printf("Inserted %d documents with IDs: %v\n", len(result.InsertedIDs), result.InsertedIDs)
	//return nil
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

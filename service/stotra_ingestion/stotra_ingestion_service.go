package stotra_ingestion

import (
	"context"
	"errors"
	"fmt"
	"github.com/Out-Of-India-Theory/oit-go-commons/logging"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/entity"
	mongoRepo "github.com/Out-Of-India-Theory/prarthana-ingestion-script/repository/mongo/prarthana_data"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/service/zoho"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/util"
	"github.com/go-audio/wav"
	"github.com/hajimehoshi/go-mp3"
	"go.uber.org/zap"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type StotraIngestionService struct {
	logger                   *zap.Logger
	prarthanaMongoRepository mongoRepo.MongoRepository
	zohoService              zoho.Service
}

func InitStotraIngestionService(ctx context.Context,
	prarthanaMongoRepository mongoRepo.MongoRepository,
	zohoService zoho.Service,
) *StotraIngestionService {
	return &StotraIngestionService{
		logger:                   logging.WithContext(ctx),
		prarthanaMongoRepository: prarthanaMongoRepository,
		zohoService:              zohoService,
	}
}

//func (s *StotraIngestionService) StotraIngestion(ctx context.Context, startID, endID int) (map[string]entity.Stotra, error) {
//	var response entity.ShlokaSheetResponse
//	err := s.zohoService.GetSheetData(ctx, "stotra", &response)
//	if err != nil {
//		return nil, err
//	}
//	if len(response.Records) == 0 {
//		return nil, errors.New("no records found")
//	}
//
//	stotraMap := map[string]entity.Stotra{}
//	var stotras []entity.Stotra
//	for i, record := range response.Records {
//		log.Printf("Processing record %d\n", i+1) // Log the current record number
//		idf, ok := record["ID"].(float64)
//		if !ok {
//			return nil, fmt.Errorf("invalid ID")
//		}
//		id := int(idf)
//		if id < startID || id > endID {
//			continue
//		}
//
//		name, ok := record["Name (Optional)"].(string)
//		if !ok {
//			return nil, fmt.Errorf("invalid Name : %d", id)
//		}
//		name = strings.TrimSpace(name)
//		re := regexp.MustCompile(`[^a-zA-Z0-9\s]+`)
//		if re.MatchString(name) {
//			return nil, fmt.Errorf("the name '%s' contains special characters. Please remove them", name)
//		}
//		baseFilename := strings.ToLower(strings.ReplaceAll(strings.TrimSuffix(name, "|"), " ", "_"))
//		isWav := true
//		stotraUrl := "https://d161fa2zahtt3z.cloudfront.net/audio/" + baseFilename + ".wav"
//		stotraUrlmp3 := "https://d161fa2zahtt3z.cloudfront.net/audio/" + baseFilename + ".mp3"
//		if !util.UrlExists(stotraUrl) {
//			if !util.UrlExists(stotraUrlmp3) {
//				return nil, fmt.Errorf("audio URL does not exist: %s", stotraUrl)
//			}
//			isWav = false
//			stotraUrl = stotraUrlmp3
//		}
//		resp, err := http.Get(stotraUrl)
//		if err != nil || resp.StatusCode != http.StatusOK {
//			fmt.Printf("Error accessing StotraUrl: %s, Error: %v\n", stotraUrl, err)
//			continue
//		}
//		defer resp.Body.Close()
//		pattern := "*.wav"
//		if !isWav {
//			pattern = "*.mp3"
//		}
//
//		tempFile, err := os.CreateTemp("", pattern)
//		if err != nil {
//			fmt.Println("Error creating temp file:", err)
//			continue
//		}
//		defer os.Remove(tempFile.Name())
//		_, err = io.Copy(tempFile, resp.Body)
//		if err != nil {
//			fmt.Println("Error saving audio file:", err)
//			continue
//		}
//
//		durationStr, durationInSeconds, err := getDurationFromFile(tempFile.Name())
//		if err != nil {
//			fmt.Println("Error getting duration:", err)
//			continue
//		}
//
//		_, durationInMilliseconds, err := getDurationFromFileInMilliseconds(tempFile.Name())
//		if err != nil {
//			fmt.Println("Error getting duration in milliseconds:", err)
//			continue
//		}
//		shlokIds := fmt.Sprintf("%v", record["Shloka ID (Comma separated - Ordered)"])
//		stotra := entity.Stotra{
//			ID:    strconv.Itoa(id),
//			IntId: id,
//			Title: map[string]string{
//				"default": name,
//			},
//			ShlokIds:               util.GetSplittedString(shlokIds),
//			Duration:               durationStr,
//			DurationInSeconds:      durationInSeconds,
//			DurationInMilliseconds: durationInMilliseconds,
//			StotraUrl:              stotraUrl,
//		}
//
//		stotraMap[strconv.Itoa(id)] = stotra
//		stotras = append(stotras, stotra)
//	}
//	return stotraMap, s.prarthanaMongoRepository.InsertManyStotras(ctx, stotras)
//}

func (s *StotraIngestionService) StotraIngestion(ctx context.Context, startID, endID int) (map[string]entity.Stotra, error) {
	var response entity.ShlokaSheetResponse
	err := s.zohoService.GetSheetData(ctx, "stotra", &response)
	if err != nil {
		return nil, err
	}
	if len(response.Records) == 0 {
		return nil, errors.New("no records found")
	}

	stotraMap := make(map[string]entity.Stotra)
	var mu sync.Mutex
	var wg sync.WaitGroup
	errChan := make(chan error, 1)
	sem := make(chan struct{}, 10)

	for i, record := range response.Records {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			sem <- struct{}{}
			wg.Add(1)
			go func(i int, record map[string]interface{}) {
				defer func() {
					<-sem
					wg.Done()
				}()

				idf, ok := record["ID"].(float64)
				if !ok {
					errChan <- fmt.Errorf("invalid ID")
					return
				}
				id := int(idf)
				if id < startID || id > endID {
					return
				}
				log.Printf("total processed %d\n", i+1)
				log.Printf("Processing record : row number %d\n", id)

				nameDefault, ok := record["Name (Optional) (Default)"].(string)
				if !ok {
					errChan <- fmt.Errorf("invalid Name : %d", id)
					return
				}
				nameDefault = strings.TrimSpace(nameDefault)
				re := regexp.MustCompile(`[^a-zA-Z0-9\s\-]+`)
				if re.MatchString(nameDefault) {
					errChan <- fmt.Errorf("the name '%s' contains special characters. Please remove them", nameDefault)
					return
				}
				nameHindi, ok := record["Name (Optional) (Hindi)"].(string)
				nameKannada, ok := record["Name (Optional) (Kannada)"].(string)
				nameMarathi, ok := record["Name (Optional) (Marathi)"].(string)
				nameTamil, ok := record["Name (Optional) (Tamil)"].(string)
				nameTelugu, ok := record["Name (Optional) (Telugu)"].(string)
				nameGujarati, ok := record["Name (Optional) (Gujarati)"].(string)

				baseFilename := strings.ToLower(util.SanitizeString(nameDefault))
				//strings.ToLower(strings.ReplaceAll(strings.TrimSuffix(name, "|"), " ", "_"))
				isWav := true
				stotraUrl := "https://d161fa2zahtt3z.cloudfront.net/audio/" + baseFilename + ".wav"
				stotraUrlmp3 := "https://d161fa2zahtt3z.cloudfront.net/audio/" + baseFilename + ".mp3"
				if !util.UrlExists(stotraUrl) {
					if !util.UrlExists(stotraUrlmp3) {
						errChan <- fmt.Errorf("audio URL does not exist: %s", stotraUrl)
						return
					}
					isWav = false
					stotraUrl = stotraUrlmp3
				}

				resp, err := http.Get(stotraUrl)
				if err != nil || resp.StatusCode != http.StatusOK {
					log.Printf("Error accessing StotraUrl: %s, Error: %v\n", stotraUrl, err)
					return
				}
				defer resp.Body.Close()

				pattern := "*.wav"
				if !isWav {
					pattern = "*.mp3"
				}

				tempFile, err := os.CreateTemp("", pattern)
				if err != nil {
					log.Println("Error creating temp file:", err)
					return
				}
				defer os.Remove(tempFile.Name())
				_, err = io.Copy(tempFile, resp.Body)
				if err != nil {
					log.Println("Error saving audio file:", err)
					return
				}

				durationStr, durationInSeconds, err := getDurationFromFile(tempFile.Name())
				if err != nil {
					log.Println("Error getting duration:", err)
					return
				}

				_, durationInMilliseconds, err := getDurationFromFileInMilliseconds(tempFile.Name())
				if err != nil {
					log.Println("Error getting duration in milliseconds:", err)
					return
				}

				shlokIds := fmt.Sprintf("%v", record["Shloka ID (Comma separated - Ordered)"])
				stotra := entity.Stotra{
					ID:    strconv.Itoa(id),
					IntId: id,
					Title: map[string]string{
						"default": nameDefault,
						"hi":      nameHindi,
						"kn":      nameKannada,
						"mr":      nameMarathi,
						"ta":      nameTamil,
						"te":      nameTelugu,
						"gu":      nameGujarati,
					},
					ShlokIds:               util.GetSplittedString(shlokIds),
					Duration:               durationStr,
					DurationInSeconds:      durationInSeconds / 2,
					DurationInMilliseconds: durationInMilliseconds / 2,
					StotraUrl:              stotraUrl,
				}

				mu.Lock()
				stotraMap[strconv.Itoa(id)] = stotra
				mu.Unlock()
			}(i, record)
		}
	}

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Check for errors
	if err := <-errChan; err != nil {
		return nil, err
	}

	// Insert into the database
	stotras := make([]entity.Stotra, 0, len(stotraMap))
	for _, stotra := range stotraMap {
		stotras = append(stotras, stotra)
	}
	return stotraMap, s.prarthanaMongoRepository.InsertManyStotras(ctx, stotras)
}

func getDurationFromFile(filename string) (string, int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	ext := filepath.Ext(filename)
	var totalSeconds int

	switch ext {
	case ".mp3":
		decoder, err := mp3.NewDecoder(file)
		if err != nil {
			return "", 0, err
		}
		length := decoder.Length()
		sampleRate := 96000 // Most common sample rate for MP3 files
		duration := time.Duration(length) * time.Second / time.Duration(sampleRate)
		totalSeconds = int(duration.Seconds())

	case ".wav":
		decoder := wav.NewDecoder(file)
		if !decoder.IsValidFile() {
			return "", 0, fmt.Errorf("invalid WAV file")
		}
		buf, err := decoder.FullPCMBuffer()
		if err != nil {
			return "", 0, err
		}
		sampleRate := buf.Format.SampleRate
		duration := time.Duration(buf.NumFrames()) * time.Second / time.Duration(sampleRate)
		totalSeconds = int(duration.Seconds())

	default:
		return "", 0, fmt.Errorf("unsupported file type: %s", ext)
	}

	minutes := int(math.Max(1, math.Round((float64(totalSeconds/2) / float64(60)))))
	durationStr := fmt.Sprintf("%dm", minutes)

	return durationStr, totalSeconds / 2, nil
}

func getDurationFromFileInMilliseconds(filename string) (string, int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	ext := filepath.Ext(filename)
	var totalMilliseconds int

	switch ext {
	case ".mp3":
		decoder, err := mp3.NewDecoder(file)
		if err != nil {
			return "", 0, err
		}
		length := decoder.Length()
		sampleRate := 96000
		duration := time.Duration(length) * time.Second / time.Duration(sampleRate)
		totalMilliseconds = int(duration.Milliseconds())

	case ".wav":
		decoder := wav.NewDecoder(file)
		if !decoder.IsValidFile() {
			return "", 0, fmt.Errorf("invalid WAV file")
		}
		buf, err := decoder.FullPCMBuffer()
		if err != nil {
			return "", 0, err
		}
		sampleRate := buf.Format.SampleRate
		duration := time.Duration(buf.NumFrames()) * time.Second / time.Duration(sampleRate)
		totalMilliseconds = int(duration.Milliseconds())

	default:
		return "", 0, fmt.Errorf("unsupported file type: %s", ext)
	}

	minutes := int(math.Max(1, math.Round((float64(totalMilliseconds)/1000.0)/60.0)))
	durationStr := fmt.Sprintf("%dm", minutes)

	return durationStr, totalMilliseconds, nil
}

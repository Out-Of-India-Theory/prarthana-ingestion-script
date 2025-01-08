package stotra_ingestion

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/Out-Of-India-Theory/oit-go-commons/logging"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/entity"
	mongoRepo "github.com/Out-Of-India-Theory/prarthana-automated-script/repository/mongo/prarthana_data"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/service/util"
	"github.com/go-audio/wav"
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
	"time"
)

type StotraIngestionService struct {
	logger                   *zap.Logger
	prarthanaMongoRepository mongoRepo.MongoRepository
}

func InitStotraIngestionService(ctx context.Context,
	prarthanaMongoRepository mongoRepo.MongoRepository,
) *StotraIngestionService {
	return &StotraIngestionService{
		logger:                   logging.WithContext(ctx),
		prarthanaMongoRepository: prarthanaMongoRepository,
	}
}

func (s *StotraIngestionService) StotraIngestion(ctx context.Context, csvFilePath string, startID, endID int) (map[string]entity.Stotra, error) {
	// Open the CSV file
	file, err := os.Open(csvFilePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Create a CSV reader
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // Allow variable number of fields per record

	// Read the CSV header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("error reading header: %w", err)
	}

	// Map CSV header to field indices in the Prayer struct
	fieldMap := make(map[string]int)
	for i, field := range header {
		fieldMap[field] = i
	}

	stotraMap := map[string]entity.Stotra{}

	// Create a slice to store prayer objects
	var stotras []entity.Stotra

	// Read remaining records from the CSV file
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading records: %w", err)
	}

	// Iterate over each record in the CSV file
	for i, record := range records {
		log.Printf("Processing record %d\n", i+1) // Log the current record number

		// Defensive check to avoid index out of range errors
		if len(record) <= fieldMap["ID"] {
			log.Printf("Skipping record %d: Missing ID field\n", i+1)
			continue
		}

		// Convert the ID from string to an integer
		id, err := strconv.Atoi(record[fieldMap["ID"]])
		if err != nil {
			log.Printf("Skipping record %d: Invalid ID format\n", i+1)
			continue
		}

		// Check if the ID is within the specified range
		if id < startID || id > endID {
			continue
		}

		// Defensive check for the Name field
		if len(record) <= fieldMap["Name (Optional)"] {
			log.Printf("Skipping record %d: Missing Name field\n", i+1)
			continue
		}
		title := record[fieldMap["Name (Optional)"]]
		re := regexp.MustCompile(`[^a-zA-Z0-9\s]+`)
		if re.MatchString(title) {
			return nil, fmt.Errorf("the name '%s' contains special characters. Please remove them", title)
		}
		baseFilename := strings.ToLower(strings.ReplaceAll(strings.TrimSuffix(title, "|"), " ", "_"))
		stotraUrl := "https://d161fa2zahtt3z.cloudfront.net/audio/" + baseFilename + ".wav"

		// Download the file from StotraUrl
		resp, err := http.Get(stotraUrl)
		if err != nil || resp.StatusCode != http.StatusOK {
			fmt.Printf("Error accessing StotraUrl: %s, Error: %v\n", stotraUrl, err)
			continue
		}
		defer resp.Body.Close()

		// Save the WAV file temporarily for duration calculation
		tempFile, err := os.CreateTemp("", "*.wav")
		if err != nil {
			fmt.Println("Error creating temp file:", err)
			continue
		}
		defer os.Remove(tempFile.Name()) // Clean up the temp file
		_, err = io.Copy(tempFile, resp.Body)
		if err != nil {
			fmt.Println("Error saving audio file:", err)
			continue
		}

		// Calculate duration
		durationStr, durationInSeconds, err := getDurationFromFile(tempFile.Name())
		if err != nil {
			fmt.Println("Error getting duration:", err)
			continue
		}

		_, durationInMilliseconds, err := getDurationFromFileInMilliseconds(tempFile.Name())
		if err != nil {
			fmt.Println("Error getting duration in milliseconds:", err)
			continue
		}

		stotra := entity.Stotra{
			ID: strconv.Itoa(id),
			Title: map[string]string{
				"default": title,
			},
			ShlokIds:               util.GetSplittedString(record[fieldMap["Shloka ID (Comma separated - Ordered)"]]),
			Duration:               durationStr,
			DurationInSeconds:      durationInSeconds,
			DurationInMilliseconds: durationInMilliseconds,
			StotraUrl:              stotraUrl,
		}

		stotraMap[strconv.Itoa(id)] = stotra
		stotras = append(stotras, stotra)
	}
	return stotraMap, s.prarthanaMongoRepository.InsertManyStotras(ctx, stotras)
}

func getDurationFromFile(filename string) (string, int, error) {
	// Open the MP3 file
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

	minutes := int(math.Max(1, math.Round((float64(totalSeconds) / float64(60)))))
	durationStr := fmt.Sprintf("%dm", minutes)

	return durationStr, totalSeconds, nil
}

func getDurationFromFileInMilliseconds(filename string) (string, int, error) {
	// Open the audio file
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
		sampleRate := 96000 // Most common sample rate for MP3 files
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

	// Convert milliseconds to minutes for the duration string
	minutes := int(math.Max(1, math.Round((float64(totalMilliseconds)/1000.0)/60.0)))
	durationStr := fmt.Sprintf("%dm", minutes)

	return durationStr, totalMilliseconds, nil
}

package file

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/mmikhail2001/test-clever-search/internal/domain/file"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var bucketName string = "test"
var minioHost string = "localhost:9000"
var channelName string = "test-queue"

type Repository struct {
	minio           *minio.Client
	mongo           *mongo.Database
	channelRabbitMQ *amqp.Channel
}

func NewRepository(minio *minio.Client, mongo *mongo.Database, channelRabbitMQ *amqp.Channel) *Repository {
	return &Repository{
		minio:           minio,
		mongo:           mongo,
		channelRabbitMQ: channelRabbitMQ,
	}
}

func (r *Repository) CreateFile(ctx context.Context, file file.File) error {
	dto := fileDTO{
		ID:          file.ID.String(),
		Filename:    file.Filename,
		Size:        file.Size,
		ContentType: file.ContentType,
		Extension:   filepath.Ext(file.Filename),
		Status:      "uploaded",
		URL:         file.URL,
	}

	collection := r.mongo.Collection("files")
	_, err := collection.InsertOne(ctx, dto)
	if err != nil {
		log.Println("Failed to insert to mongo:", err)
		return err
	}
	return nil
}

// TODO: обращение к сервису Поиска
// тот отдаст ID в mongoDB
// func (r *Repository) Search(ctx context.Context, query file.queryString) ([]file.File, error)

func (r *Repository) GetFiles(ctx context.Context, query string) ([]file.File, error) {
	var resultsDTO []fileDTO

	filter := bson.M{}
	if query != "" {
		filter["filename"] = bson.M{"$regex": primitive.Regex{Pattern: query, Options: "i"}}
	}

	opts := options.Find().SetSort(bson.D{{"filename", 1}})

	cursor, err := r.mongo.Collection("files").Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var dto fileDTO
		err := cursor.Decode(&dto)
		if err != nil {
			return nil, err
		}
		resultsDTO = append(resultsDTO, dto)
	}

	results := make([]file.File, len(resultsDTO))
	for i, fileDTO := range resultsDTO {
		uuid, err := uuid.Parse(fileDTO.ID)
		if err != nil {
			return nil, err
		}
		results[i] = file.File{
			ID:          uuid,
			Filename:    fileDTO.Filename,
			Size:        fileDTO.Size,
			ContentType: fileDTO.ContentType,
			Status:      fileDTO.Status,
			URL:         fileDTO.URL,
		}
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *Repository) GetFileByID(ctx context.Context, uuidFile uuid.UUID) (file.File, error) {
	var resultDTO fileDTO

	filter := bson.M{"_id": uuidFile.String()}
	err := r.mongo.Collection("files").FindOne(ctx, filter).Decode(&resultDTO)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return file.File{}, fmt.Errorf("GetFileByID no file with uuid: %w", err)
		}
		return file.File{}, err
	}

	parsedUUID, err := uuid.Parse(resultDTO.ID)
	if err != nil {
		return file.File{}, err
	}

	file := file.File{
		ID:          parsedUUID,
		Filename:    resultDTO.Filename,
		Size:        resultDTO.Size,
		ContentType: resultDTO.ContentType,
		Status:      resultDTO.Status,
		URL:         resultDTO.URL,
	}

	return file, nil
}

func (r *Repository) Update(ctx context.Context, file file.File) error {
	uuidString := file.ID.String()
	update := bson.M{
		"$set": bson.M{
			"filename": file.Filename,
			"status":   file.Status,
			"url":      file.URL,
		},
	}

	filter := bson.M{"_id": uuidString}
	_, err := r.mongo.Collection("files").UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update file: %w", err)
	}

	return nil
}

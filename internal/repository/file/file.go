package file

import (
	"context"
	"encoding/json"
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

// TODO:
// можно в отдельной горутине загружать файл или несолько файлов (PutObject)
// - minio.PutObjectOptions{ Progress : myreader } следить за состоянием
// - уведомить клиента о том, что файл начал загружаться
// - а до этого всего уже создать в mongo запись о том, что файл на загрузке в s3

// можно не отвечать фронту, пока не загрузим файл PutObject
func (r *Repository) Upload(ctx context.Context, file file.File) (file.File, error) {
	objectName := file.Filename
	info, err := r.minio.PutObject(ctx, bucketName, file.Path, file.File, file.Size, minio.PutObjectOptions{ContentType: file.ContentType})
	if err != nil {
		log.Println("Failed to PutObject minio:", err)
		return file, err
	}
	log.Printf("Successfully uploaded %s of size %d\n", objectName, info.Size)
	log.Printf("Full info %+v\n", info)

	file.URL = "https://" + minioHost + "/" + bucketName + "/" + file.Path
	file.ID = uuid.New()

	return file, nil
}

func (r *Repository) PublishMessage(ctx context.Context, file file.File) error {
	fileDTO := fileForQueueDTO{
		ID:          file.ID.String(),
		URL:         file.URL,
		ContentType: file.ContentType,
	}

	fileJSON, err := json.Marshal(fileDTO)
	if err != nil {
		return err
	}

	err = r.channelRabbitMQ.Publish(
		"",
		channelName,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        fileJSON,
		},
	)
	if err != nil {
		return err
	}

	return nil
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

// обращение к сервису Поиска
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

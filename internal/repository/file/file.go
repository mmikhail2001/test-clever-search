package file

import (
	"context"
	"log"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/mmikhail2001/test-clever-search/internal/domain/file"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var bucketName string = "test"
var minioHost string = "localhost:9000"

type Repository struct {
	minio *minio.Client
	mongo *mongo.Database
}

func NewRepository(minio *minio.Client, mongo *mongo.Database) *Repository {
	return &Repository{
		minio: minio,
		mongo: mongo,
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
	info, err := r.minio.PutObject(ctx, bucketName, objectName, file.File, file.Size, minio.PutObjectOptions{ContentType: file.ContentType})
	if err != nil {
		log.Println("Failed to PutObject minio:", err)
		return file, err
	}
	log.Printf("Successfully uploaded %s of size %d\n", objectName, info.Size)
	log.Printf("Full info %+v\n", info)

	file.URL = "https://" + minioHost + "/" + bucketName + "/" + objectName
	file.ID = uuid.New()

	return file, nil
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
func (r *Repository) Search(ctx context.Context, search file.SearchQuery) ([]file.File, error) {
	var resultsDTO []fileDTO

	filter := bson.M{"filename": bson.M{"$regex": primitive.Regex{Pattern: search.Query, Options: "i"}}}
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

	results := []file.File{}
	for _, fileDTO := range resultsDTO {
		uuid, err := uuid.Parse(fileDTO.ID)
		if err == nil {
			results = append(results, file.File{
				ID:          uuid,
				Filename:    fileDTO.Filename,
				Size:        fileDTO.Size,
				ContentType: fileDTO.ContentType,
				Status:      fileDTO.Status,
				URL:         fileDTO.URL,
			})
		}
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

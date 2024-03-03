package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mmikhail2001/test-clever-search/internal/repository/file"
	"github.com/mmikhail2001/test-clever-search/pkg/client/minio"
	"github.com/mmikhail2001/test-clever-search/pkg/client/mongo"
	"github.com/mmikhail2001/test-clever-search/pkg/client/rabbitmq"

	fileDelivery "github.com/mmikhail2001/test-clever-search/internal/delivery/file"
	fileUsecase "github.com/mmikhail2001/test-clever-search/internal/usecase/file"
)

func main() {

	if err := Run(); err != nil {
		fmt.Println("Error: ", err)
	}
}

func Run() error {
	minio, err := minio.NewClient()
	if err != nil {
		return err
	}

	mongoDB, err := mongo.NewClient()
	if err != nil {
		return err
	}

	channelRabbitMQ, err := rabbitmq.NewClient()
	if err != nil {
		return err
	}

	fileRepo := file.NewRepository(minio, mongoDB, channelRabbitMQ)
	fileUsecase := fileUsecase.NewUsecase(fileRepo)
	fileHandler := fileDelivery.NewHandler(fileUsecase)

	r := mux.NewRouter()
	r.HandleFunc("/", serveHome).Methods("GET")
	r.HandleFunc("/getfiles", fileHandler.GetFiles).Methods("GET")
	r.HandleFunc("/upload", fileHandler.Upload).Methods("POST")
	http.ListenAndServe(":8080", r)
	return nil
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

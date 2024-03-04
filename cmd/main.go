package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mmikhail2001/test-clever-search/internal/repository/file"
	"github.com/mmikhail2001/test-clever-search/internal/repository/notifier"
	"github.com/mmikhail2001/test-clever-search/pkg/client/minio"
	"github.com/mmikhail2001/test-clever-search/pkg/client/mongo"
	"github.com/mmikhail2001/test-clever-search/pkg/client/rabbitmq"

	fileDelivery "github.com/mmikhail2001/test-clever-search/internal/delivery/file"
	notifyDelivery "github.com/mmikhail2001/test-clever-search/internal/delivery/notifier"
	fileUsecase "github.com/mmikhail2001/test-clever-search/internal/usecase/file"
	notifyUsecase "github.com/mmikhail2001/test-clever-search/internal/usecase/notifier"
)

// TODO:
// нужный ws.conn должен выбираться исходя из cookie пользователя (сейчас заглушка userID = 1)
// бд пользователей, авторизация
// repository vs gateway - система рассылки уведомлений
// не работает ограничение на размер файла
// в доменную сущность поместился Conn   *websocket.Conn
// конфиг файл
// контексты, таймауты

// нужна заглушка для python ML, прочитывание сообщений, sleep, webhook
// зачем UUID вообще конвертить в uuid.UUID
// может статуса в бд числовыми сделать (а не строка?...) и еще нужно статусы синхронизовать с event

// можно ли беку следить, как мл обрабатывает очередь ? нужно ли?
// если в поле директории указать /dir, то 2024/03/04 21:49:12 Failed to PutObject minio: Object name contains unsupported characters.
// - проблема в / в начале

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
	notifyGateway := notifier.NewGateway()

	notifyUsecase := notifyUsecase.NewUsecase(notifyGateway)
	fileUsecase := fileUsecase.NewUsecase(fileRepo, notifyUsecase)

	fileHandler := fileDelivery.NewHandler(fileUsecase)
	notifyDelivery := notifyDelivery.NewHandler(notifyUsecase)

	r := mux.NewRouter()
	r.HandleFunc("/", serveHome).Methods("GET")
	r.HandleFunc("/getfiles", fileHandler.GetFiles).Methods("GET")
	r.HandleFunc("/upload", fileHandler.Upload).Methods("POST")
	r.HandleFunc("/complete-ml", fileHandler.CompleteProcessingFile).Methods("GET")
	r.HandleFunc("/ws", notifyDelivery.ConnectNotifications).Methods("GET")
	http.ListenAndServe(":8080", r)
	return nil
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

package main

import (
	"database/sql"
	_ "database/sql"
	eHandlers "github.com/calendar-bot/pkg/events/handlers"
	eStorage "github.com/calendar-bot/pkg/events/storage"
	eUsecase"github.com/calendar-bot/pkg/events/usecase"

	uHandlers "github.com/calendar-bot/pkg/users/handlers"
	uStorage "github.com/calendar-bot/pkg/users/storage"
	uUsecase"github.com/calendar-bot/pkg/users/usecase"
	"github.com/labstack/echo"
	"go.uber.org/zap"
	"os"
)

type RequestHandlers struct {
	eventHandlers eHandlers.EventHandlers
	userHandlers uHandlers.UserHandlers
}

func newRequestHandler(db *sql.DB) *RequestHandlers {

	eventStorage := eStorage.NewEventStorage(db)
	eventUseCase := eUsecase.NewEventUseCase(eventStorage)
	eventHandlers := eHandlers.NewEventHandlers(eventUseCase)

	userStorage := uStorage.NewUserStorage(db)
	userUseCase := uUsecase.NewUserUseCase(userStorage)
	userHandlers := uHandlers.NewUserHandlers(userUseCase)

	return &(RequestHandlers{
		eventHandlers: eventHandlers,
		userHandlers: userHandlers,
	})
}

func connectToDB() (*sql.DB, error) {
	usernameDB := "main"
	passwordDB := os.Getenv("main")
	nameDB := os.Getenv("mainnet")
	connectString := "user=" + usernameDB + " password=" + passwordDB + " dbname=" + nameDB + " sslmode=disable"

	db, err := sql.Open("postgres", connectString)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func main() {
	server := echo.New()

	db, err := connectToDB()
	if err != nil {
		zap.S().Fatalf("failed to connect to db, %v", err)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			zap.S().Errorf("failed to close db connection, %v", err)
		}
	}()

	allHandler := newRequestHandler(db)

	allHandler.eventHandlers.InitHandlers(server)

	server.Logger.Fatal(server.Start(":8080"))
}

package main

import (
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	storage "github.com/mahadeva604/audio-storage"
	"github.com/mahadeva604/audio-storage/pkg/handler"
	"github.com/mahadeva604/audio-storage/pkg/repository"
	"github.com/mahadeva604/audio-storage/pkg/service"
	"github.com/spf13/viper"
	"log"
	"os"
)

// @title AAC Share API
// @version 1.0
// @description API Server to share aac files

// @host localhost:8000
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

const saveDir = "saved/"

func main() {

	if err := initConfig(); err != nil {
		log.Fatalf("Can't load configs: %s", err.Error())
	}

	if err := godotenv.Load(); err != nil {
		log.Fatalf("Can't load .env: %s", err.Error())
	}

	cfg := repository.Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		Username: viper.GetString("db.username"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   viper.GetString("db.dbname"),
		SSLMode:  viper.GetString("db.sslmode"),
	}

	db, err := repository.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("Can't connect to db: %s", err.Error())
	}

	secretKey, err := storage.GetSecretKey()
	if err != nil {
		log.Fatalf("Can't get secret key: %s", err.Error())
	}

	repos := repository.NewRepository(db, saveDir)
	services := service.NewService(repos, secretKey)
	handlers := handler.NewHandler(services)

	srv := new(storage.Server)

	if err := srv.Run(viper.GetString("port"), handlers.InitRoutes()); err != nil {
		log.Fatalf("Can't run http server: %s", err.Error())
	}
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}

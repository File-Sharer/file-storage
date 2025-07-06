package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/File-Sharer/file-storage/internal/config"
	"github.com/File-Sharer/file-storage/internal/handler"
	"github.com/File-Sharer/file-storage/internal/server"
	"github.com/File-Sharer/file-storage/internal/service"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	logger, err := newLogger()
	if err != nil {
		log.Panicf("failed to create zap logger: %s", err.Error())
	}

	if err := initConfig(); err != nil {
		log.Panicf("failed to initialize yaml config: %s", err.Error())
	}

	services := service.New(logger)
	handlers := handler.New(services)

	srv := server.New()
	serverCfg := config.ServerConfig{
		Port: viper.GetString("app.port"),
		Handler: handlers.Init(),
		MaxHeaderBytes: 1 << 20,
		ReadTimeout: time.Second * 10,
		WriteTimeout: time.Second * 10,
	}
	go func(srv *server.Server, cfg config.ServerConfig) {
		if err := srv.Run(cfg); err != nil {
			log.Panicf("failed to run http server: %s", err.Error())
		}
	}(srv, serverCfg)

	log.Println("Server started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Println("Server shutting down...")
}

func initConfig() error {
	viper.AddConfigPath("./")
	viper.SetConfigType("yaml")
	viper.SetConfigName("app")
	return viper.ReadInConfig()
}

func newLogger() (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{
		"./app.log",
	}
	return cfg.Build()
}

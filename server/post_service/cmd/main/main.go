package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"post_service/internal/handler"
	"post_service/internal/metrics"
	"post_service/internal/repository"
	"post_service/internal/service"
	"post_service/internal/storage"
	"syscall"
	"time"

	grpc_client "post_service/internal/clients/grpc"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func main() {
	// Init Env
	currentdir, _ := os.Getwd()
	err := godotenv.Load(filepath.Join(currentdir, ".env"))
	if err != nil {
		logrus.Fatalln("couldn't load env variables", err)
	}

	// Init Storage
	storage := storage.New()

	// Init Grpc Client
	grpcClient, err := grpc_client.New(
		context.Background(),
		fmt.Sprintf("host.docker.internal:%s", os.Getenv("GRPC_PORT")),
		10*time.Second,
		3,
	)
	if err != nil {
		logrus.WithError(err).Fatalln("failed to initialize grpc client")
	}

	// Init Repository, Service and Handler
	postRepository := &repository.PostRepository{Db: storage.Db}
	postService := &service.PostService{PostRepository: postRepository}
	handler := &handler.Handler{GrpcClient: grpcClient, PostService: postService}

	// Run Server
	server := &http.Server{
		Addr:         ":8081",
		Handler:      handler.New(),
		ReadTimeout:  4 * time.Second,
		WriteTimeout: 4 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	go server.ListenAndServe()

	// Run metrics server
	go func() {
		_ = metrics.Listen(":9081")
	}()

	// Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	stoppingSignal := <-stop
	logrus.WithField("signal", stoppingSignal).Info("stopping application")

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logrus.Errorf("HTTP server shutdown error")
	}
	storage.Stop()

	logrus.Info("application stopped")
}

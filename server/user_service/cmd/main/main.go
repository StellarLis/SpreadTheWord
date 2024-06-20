package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
	"user_service/internal/amqp"
	"user_service/internal/handlers"
	"user_service/internal/metrics"
	"user_service/internal/repository"
	"user_service/internal/services"
	"user_service/internal/storage"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func main() {
	// Init Config
	currentdir, _ := os.Getwd()
	err := godotenv.Load(filepath.Join(currentdir, ".env"))
	if err != nil {
		logrus.Fatalln(err)
	}

	// Init Db
	storage := storage.New()

	// Connect to RabbitMQ instance
	amqp_handler := amqp.New()

	// Init Handler
	userRepository := repository.UserRepository{Db: storage.Db}
	userService := services.UserService{
		UserRepository: &userRepository,
		JwtSecretKey:   os.Getenv("JWT_SECRET_KEY"),
		Amqp:           amqp_handler,
	}
	handler := handlers.HttpHandler{UserService: userService}

	// Run gRPC Handler
	grpcPort, err := strconv.ParseInt(os.Getenv("GRPC_PORT"), 10, 0)
	if err != nil {
		logrus.Fatalln(err)
	}
	grpcHandler := handlers.NewGrpcHandler(int(grpcPort), os.Getenv("JWT_SECRET_KEY"))
	go grpcHandler.MustRun()

	// Run Http Server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler.New(),
		ReadTimeout:  4 * time.Second,
		WriteTimeout: 4 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	go server.ListenAndServe()

	// Run metrics server
	go func() {
		_ = metrics.Listen(":9080")
	}()

	// Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	stoppingSignal := <-stop
	logrus.WithField("signal", stoppingSignal).Info("stopping application")

	grpcHandler.Stop()
	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logrus.Fatalf("HTTP Server shutdown error")
	}
	storage.Stop()
	amqp_handler.Close()

	logrus.Info("application stopped")
}

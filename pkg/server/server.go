package server

import (
	"context"
	"database/sql"
	"errors"
	"first-version/pkg/api"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Run настраивает маршрутизцию, статический файловый сервер и запускает НТТР-
// сервер с поддержкой корректного завершения работы (Graceful Shutdown).
func Run(todoPort string, db *sql.DB) {
	logger := log.New(os.Stdout, "[SERVER]", log.LstdFlags|log.Lshortfile)

	router := http.NewServeMux()
	taskServer := api.NewTaskServer(db)

	logger.Println("Регистрация сетевых маршрутов API планировщика...")
	router.HandleFunc("/api/task", taskServer.TaskHandler)
	router.HandleFunc("/api/tasks", taskServer.TasksHandler)
	router.HandleFunc("/api/task/done", taskServer.DoneHandler)
	router.HandleFunc("/api/nextdate", taskServer.NextDateHandler)

	webDir := "./web"
	if _, err := os.Stat(webDir); os.IsNotExist(err) { //++
		logger.Fatalf("Критическая ошибка: Директория со статикой '%s' не найдена! Проверить файловую структуру.", webDir)
	}

	fileServer := http.FileServer(http.Dir(webDir))
	router.Handle("/", fileServer)

	server := &http.Server{
		Addr:         ":" + todoPort,
		Handler:      router,
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	serverErrors := make(chan error, 1)

	// Запуск НТТР-сервера в фоновой горутине для предотвращения блокировки основного потока.
	go func() {
		logger.Printf("Веб-сервер успешно запущен. Ожидание запросов на http://localhost:%s", todoPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Fatalf("Критическая ошибка при старте сервера: %v", err)
	case sig := <-shutdown:
		logger.Printf("Получен системный сигнал [%v]. Инициирована плавная остановка...", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Printf("Превышен лимит 5 секунд на плавную остановку: %v", err)
			logger.Println("Принудительное закрытие сетевых дескрипторов...")
			_ = server.Close()
		}
	}

	logger.Println("Сетевой слой успешно остановлен. Все порты освобождены.")
}

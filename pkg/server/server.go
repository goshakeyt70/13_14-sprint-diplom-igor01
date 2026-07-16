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

// Run настраивает изолированную экосистему веб-сервера и
// блокирует основной поток выполнения до системного прерывания.
func Run(todoPort string, db *sql.DB) {
	// 1. Логгер с префексом - помогает при разборе логов понимать,
	// какая часть приложения пишет отчет.
	logger := log.New(os.Stdout, "[SERVER]", log.LstdFlags|log.Lshortfile)

	// 2. Создаем изолированный (локальный) маршрутизатор запросов.
	// Избегаем глобального http.DefaultServeMux - это стандарт
	// безопасности разработки в Go.
	router := http.NewServeMux()

	// 3. Dependency Injection: внедряем базу данных в слой бизнес-логики (API)
	taskServer := api.NewTaskServer(db)

	logger.Println("Регистрация сетевых маршрутов API планировщика...")

	// Привязываем URL-пути к методам нашей структуры-сервера.
	router.HandleFunc("/api/task", taskServer.AddHandler)          // Шаг 4: Добавление задачи
	router.HandleFunc("/api/nextdate", taskServer.NextDateHandler) // Шаг 3: Математика расчета дат

	// 4. Обслуживание статического фронтенда (HTML/JS/CSS панели задач)
	webDir := "./web"

	// Защита от дураков: если пользователь забыл склонировать папку web, сервер упадет подав сигнал,
	// а не молча.
	if _, err := os.Stat(webDir); os.IsNotExist(err) {
		logger.Fatalf("Критическая ошибка: Директория со статикой '%s' не найдена! Проверить файловую структуру.", webDir)
	}

	fileServer := http.FileServer(http.Dir(webDir))
	// Корень "/" обрабатывает все запросы к статическим файлам.
	router.Handle("/", fileServer)

	// 5. Конфигурация НТТР-сервера
	// Явные лимиты предотвращают зависание системных потоков (утечку горутин).
	server := &http.Server{
		Addr:         ":" + todoPort,
		Handler:      router, // Наш кастомный роутер со всеми маршрутами
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,  // Клиент должен успеть отправить запрос за 5 сек.
		WriteTimeout: 10 * time.Second, // Сервер обязуется отдать ответ в течение 10 сек.
		IdleTimeout:  15 * time.Second, // Время жизни неактивного Keep-Alive соединения.
	}

	// 6. Реализация Graceful Shutdown (безопасная остановка без потери транзакций БД).
	serverErrors := make(chan error, 1)

	// Запуск бесконечного цикла прослушивания сети в фоновой горутине.
	// Метод ListenAndServe блокирующий. Если не убрать его в горутину (go func),
	// код ниже никогда не выполнится, и сервер нельзя будет остановить плавно.
	go func() {
		logger.Printf("Веб-сервер успешно запущен. Ожидание запросов на http://localhost:%s", todoPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			// Записываем непредвиденную ошибку сети (например, порт занят)
			serverErrors <- err
		}
	}()

	// Создаем буферизированный канал для системных сигналов операционной системыю
	shutdown := make(chan os.Signal, 1)

	// Подписываем канал на событие Interruption (Ctrl+C) и Termination (остановка процесса
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Селектор блокирует поток и ждет наступления одного из двух событий.
	select {
	case err := <-serverErrors:
		// Если сервер упал на старте
		logger.Fatalf("Критическая ошибка при старте сервера: %v", err)

	case sig := <-shutdown:
		// Если пользователь решил остановить сервер в консоли Ubuntu.
		logger.Printf("Получен системный сигнал [%v]. Инициирована плавная остановка...", sig)

		// Выделим жесткий лимит времени (5 секунд) на завершение активных НТТР-запросов.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Даем серверу команду прекратить принимать новые соединения и дописать старые.
		if err := server.Shutdown(ctx); err != nil {
			logger.Printf("Превышен лимит 5 секунд на плавную остановку: %v", err)
			logger.Println("Принудительное закрытие сетевых дескрипторов...")
			_ = server.Close() // Жесткое закрытие принудительно
		}
	}
	logger.Println("Сетевой слой успешно остановлен. Все порты освобождены.")
}

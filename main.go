package main

import (
	"first-version/pkg/db"
	"first-version/pkg/server"
	"log"
	"os"
)

func main() {
	log.Println("[MAIN] Инициализация конфигурации планировщика...")

	dbPath := os.Getenv("TODO_DBFILE")
	if dbPath == "" {
		dbPath = "scheduler.db"
	}

	todoPort := os.Getenv("TODO_PORT")
	if todoPort == "" {
		todoPort = "7540"
	}

	database, err := db.InitDatabase(dbPath)
	if err != nil {
		log.Fatalf("[MAIN] Критическая ошибка инициализации БД: %v", err)
	}

	defer func() {
		log.Println("[MAIN]Завершение работы приложения. Закрытие пула БД...")
		if err := database.Close(); err != nil {
			log.Printf("[MAIN] Ошибка при закрытии БД: %v", err)
		}
		log.Println("[MAIN] База данных успешно сохранена и закрыта на диске.")
	}()

	server.Run(todoPort, database)
}

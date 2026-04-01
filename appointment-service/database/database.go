package database

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Параметры подключения
var (
	//host = "localhost" // "host.docker.internal" // "db" postgres
	host     = getEnv("DB_HOST", "localhost")
	port     = getEnvAsInt("DB_PORT", 5432)
	user     = getEnv("DB_USER", "postgres")
	password = getEnv("DB_PASSWORD", "87363699")
	dbname   = getEnv("DB_NAME", "isheecrm")
)

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// Вспомогательная функция для чтения числовых переменных окружения
func getEnvAsInt(key string, defaultValue int) int {
	if valueStr, exists := os.LookupEnv(key); exists {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

var Pool *pgxpool.Pool

// Close закрывает пул соединений с базой данных
func Close() {
	Pool.Close()
	fmt.Println("Соединение с базой данных закрыто")
}

// ConnectDB устанавливает соединение с базой данных PostgreSQL
func ConnectDB() error {
	var err error
	// Формируем строку подключения
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?client_encoding=UTF8", user, password, host, port, dbname)

	// Создаем контекст с таймаутом, чтобы подключение не висело вечно
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Подключаемся к базе данных PostgreSQL
	Pool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("Ошибка подключения к базе данных PostgreSQL: %w", err)
	}

	// Проверяем соединение с базой данных
	err = Pool.Ping(ctx)
	if err != nil {
		Pool.Close()
		return fmt.Errorf("Не удалось выполнить ping базы данных: %w", err)
	}

	fmt.Println("Подключение к базе данных PostgreSQL успешно выполнено")
	return nil
}

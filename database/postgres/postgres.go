package postgres

import (
	"Makhkets/internal/configs"
	"Makhkets/pkg/logging"
	"context"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Client interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

func InitDatabase() Client {
	cfg := configs.GetConfig()
	logger := logging.GetLogger()

	// Настройки подключения
	config, err := pgxpool.ParseConfig(fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.Storage.Username, cfg.Storage.Password, cfg.Storage.Host, cfg.Storage.Port, cfg.Storage.Database))
	if err != nil {
		logger.Fatal(fmt.Sprintf("Ошибка при парсинге строки подключения: %v", err))
	}

	// Создание пула соединений
	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Ошибка при создании пула соединений: %v", err))
	}

	return pool
}

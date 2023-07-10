package postgres

import (
	"Makhkets/internal/configs"
	"Makhkets/pkg/logging"
	"Makhkets/schema"
	"database/sql"
	"github.com/pressly/goose/v3"

	"context"
	"fmt"

	_ "github.com/lib/pq"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Client interface {
	// Exec INSERT, UPDATE, DELETE
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)

	// Query Select
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)

	// QueryRow Query one select
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row

	// Begin transaction
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

	migrate(cfg)

	return pool
}

func migrate(cfg *configs.Config) {
	// Открываем соединение с базой данных.
	db, err := sql.Open(
		"postgres",
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", cfg.Storage.Username, cfg.Storage.Password, cfg.Storage.Host, cfg.Storage.Port, cfg.Storage.Database),
	)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Проверяем соединение с базой данных.
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	// setup database

	goose.SetBaseFS(schema.EmbedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		panic(err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		fmt.Println(err)
	}
}

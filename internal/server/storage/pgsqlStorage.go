package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/atrian/devmetrics/internal/appconfig/serverconfig"
)

type PgSQLStorage struct {
	pgPool  *pgxpool.Pool
	metrics *MetricsDicts
	config  *serverconfig.Config
}

func NewPgSQLStorage(config *serverconfig.Config) *PgSQLStorage {
	dbPool, poolErr := pgxpool.Connect(context.Background(), config.Server.DBDSN)
	if poolErr != nil {
		log.Fatal(poolErr)
	}

	return &PgSQLStorage{
		pgPool:  dbPool,
		metrics: NewMetricsDicts(),
		config:  config,
	}
}

func upsertMetricStatement() string {
	return `
		INSERT INTO public.metrics (id, type, delta, value)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE
		SET type = $2, delta = $3, value = $4;
		`
}

func (s *PgSQLStorage) StoreGauge(name string, value float64) {
	_, err := s.pgPool.Exec(context.Background(), upsertMetricStatement(), name, "gauge", value, nil)
	if err != nil {
		fmt.Println(err)
	}
}

func (s *PgSQLStorage) StoreCounter(name string, value int64) {
	// Проверяем есть ли уже счетчик в базе, если есть, суммируем данные
	storedCounter, exist := s.GetCounter(name)
	if exist {
		value += storedCounter
	}

	// Порядок аргументов id, type, delta, value
	_, err := s.pgPool.Exec(context.Background(), upsertMetricStatement(), name, "counter", nil, value)
	if err != nil {
		fmt.Println(err)
	}
}

func (s *PgSQLStorage) GetGauge(name string) (float64, bool) {
	var delta float64

	sqlStatement := `SELECT delta FROM public.metrics WHERE id=$1;`
	row := s.pgPool.QueryRow(context.Background(), sqlStatement, name)

	switch err := row.Scan(&delta); err {
	case nil:
		return delta, true
	default:
		return delta, false
	}
}

func (s *PgSQLStorage) GetCounter(name string) (int64, bool) {
	var value int64

	sqlStatement := `SELECT value FROM public.metrics WHERE id=$1;`
	row := s.pgPool.QueryRow(context.Background(), sqlStatement, name)

	switch err := row.Scan(&value); err {
	case nil:
		return value, true
	default:
		return value, false
	}
}

func (s *PgSQLStorage) GetMetrics() *MetricsDicts {
	var (
		metricID   string
		metricType string
		delta      sql.NullFloat64
		value      sql.NullInt64
	)

	sqlStatement := `SELECT id, type, delta, value FROM public.metrics;`
	rows, err := s.pgPool.Query(context.Background(), sqlStatement)

	if err != nil {
		fmt.Println(err)
		return s.metrics
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&metricID, &metricType, &delta, &value)
		if err != nil {
			fmt.Println(err)
			continue
		}

		switch metricType {
		case "gauge":
			s.metrics.GaugeDict[metricID] = gauge(delta.Float64)
		case "counter":
			s.metrics.CounterDict[metricID] = counter(value.Int64)
		default:
			continue
		}
	}

	return s.metrics
}

// RunOnStart на старте запускаем миграции
func (s *PgSQLStorage) RunOnStart() {
	runMigrations(s.config.Server.DBDSN)
}

// RunOnClose закрываем pgPool
func (s *PgSQLStorage) RunOnClose() {
	if s.pgPool != nil {
		s.pgPool.Close()
	}
}

// runMigrations запускает миграции перед использованием приложения
func runMigrations(dsn string) {
	fmt.Println("Start migration process...")
	m, mErr := migrate.New(
		"file://internal/server/database/migrations",
		dsn)
	if mErr != nil {
		log.Fatal("Can't prepare for migrations:", mErr.Error())
	}
	mErr = m.Up()
	if mErr != nil {
		fmt.Println("Migration complete with error:", mErr.Error())
	} else {
		fmt.Println("Successfully migrated")
	}
}

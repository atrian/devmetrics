package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"

	"github.com/atrian/devmetrics/internal/appconfig/serverconfig"
	"github.com/atrian/devmetrics/internal/dto"
)

type PgSQLStorage struct {
	pgPool  *pgxpool.Pool
	metrics *MetricsDicts
	config  *serverconfig.Config
	logger  *zap.Logger
}

// Проверка имплементации интерфейса. Как это работает?
var _ Repository = (*PgSQLStorage)(nil)

func NewPgSQLStorage(config *serverconfig.Config, logger *zap.Logger) (*PgSQLStorage, error) {
	dbPool, poolErr := pgxpool.Connect(context.Background(), config.Server.DBDSN)
	if poolErr != nil {
		logger.Error("NewPgSQLStorage pgxpool.Connect", zap.Error(poolErr))
		return nil, poolErr
	}

	return &PgSQLStorage{
		pgPool:  dbPool,
		metrics: NewMetricsDicts(),
		config:  config,
		logger:  logger,
	}, nil
}

// upsertMetricQuery порядок аргументов в запросе: id, type, delta, value
func upsertMetricQuery() string {
	return `
		INSERT INTO public.metrics (id, type, delta, value)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id, type) DO UPDATE
		SET type = $2, delta = $3, value = $4;`
}

func (s *PgSQLStorage) StoreGauge(name string, value float64) error {
	_, err := s.pgPool.Exec(context.Background(), upsertMetricQuery(), name, "gauge", nil, value)
	if err != nil {
		s.logger.Error("StoreGauge pgPool.Exec", zap.Error(err))
		return fmt.Errorf(`failed store gauge: %w`, err)
	}
	return nil
}

func (s *PgSQLStorage) StoreCounter(name string, value int64) error {
	// Проверяем есть ли уже счетчик в базе, если есть, суммируем данные
	storedCounter, exist := s.GetCounter(name)
	if exist {
		value += storedCounter
	}

	// Порядок аргументов id, type, delta, value
	_, err := s.pgPool.Exec(context.Background(), upsertMetricQuery(), name, "counter", value, nil)
	if err != nil {
		s.logger.Error("StoreCounter pgPool.Exec", zap.Error(err))
		return fmt.Errorf(`failed store counter: %w`, err)
	}

	return nil
}

func (s *PgSQLStorage) GetGauge(name string) (float64, bool) {
	var value float64

	sqlQuery := `SELECT value FROM public.metrics WHERE id=$1;`
	row := s.pgPool.QueryRow(context.Background(), sqlQuery, name)

	switch err := row.Scan(&value); err {
	case nil:
		return value, true
	default:
		s.logger.Debug("GetGauge row.Scan", zap.Error(err))
		return value, false
	}
}

func (s *PgSQLStorage) GetCounter(name string) (int64, bool) {
	var delta int64

	sqlQuery := `SELECT delta FROM public.metrics WHERE id=$1;`
	row := s.pgPool.QueryRow(context.Background(), sqlQuery, name)

	switch err := row.Scan(&delta); err {
	case nil:
		return delta, true
	default:
		s.logger.Debug("GetCounter row.Scan", zap.Error(err))
		return delta, false
	}
}

func (s *PgSQLStorage) GetMetrics() *MetricsDicts {
	var (
		metricID   string
		metricType string
		value      sql.NullFloat64
		delta      sql.NullInt64
	)

	sqlStatement := `SELECT id, type, delta, value FROM public.metrics;`
	rows, err := s.pgPool.Query(context.Background(), sqlStatement)

	if err != nil {
		s.logger.Error("GetMetrics pgPool.Query", zap.Error(err))
		return s.metrics
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&metricID, &metricType, &delta, &value)
		if err != nil {
			s.logger.Error("GetMetrics rows.Scan", zap.Error(err))
			continue
		}

		switch metricType {
		case "gauge":
			s.metrics.GaugeDict[metricID] = gauge(value.Float64)
		case "counter":
			s.metrics.CounterDict[metricID] = counter(delta.Int64)
		default:
			continue
		}
	}

	return s.metrics
}

// SetMetrics сохранение слайса DTO Metrics в бд.
func (s *PgSQLStorage) SetMetrics(metrics []dto.Metrics) {
	memoryCounters := make(map[string]int64)
	ctx := context.Background()

	// начинаем транзакцию
	tx, err := s.pgPool.Begin(ctx)
	if err != nil {
		s.logger.Error("SetMetrics pgPool.Begin", zap.Error(err))
	}

	batch := &pgx.Batch{}
	for _, metric := range metrics {
		switch metric.MType {
		case "counter":
			// получаем сохраненное ранее в БД значение
			storedCounter, _ := s.GetCounter(metric.ID)
			// получаем сохраненное ранее значение в памяти в рамках одного batch запроса
			memoryCounter := memoryCounters[metric.ID]

			s.logger.Debug("SetMetrics counter value update: ",
				zap.Int64("storedCounter", storedCounter),
				zap.Int64("memoryCounter", memoryCounter),
				zap.Int64("*metric.Delta", *metric.Delta),
				zap.Int64("position sum", storedCounter+memoryCounter+*metric.Delta))

			batch.Queue(upsertMetricQuery(), metric.ID, metric.MType, *metric.Delta+storedCounter+memoryCounter, nil)
			// обновляем сумму в памяти
			memoryCounters[metric.ID] += *metric.Delta
		case "gauge":

			s.logger.Debug("SetMetrics gauge value update: ",
				zap.Float64("*metric.Value", *metric.Value))

			// записываем последнее если пришла пачка одинаковых
			batch.Queue(upsertMetricQuery(), metric.ID, metric.MType, nil, *metric.Value)
		default:
			continue
		}
	}

	result := tx.SendBatch(ctx, batch)
	var queryError error
	for queryError == nil {
		_, queryError = result.Exec()
	}

	// коммит транзакции
	err = tx.Commit(ctx)
	if err != nil {
		s.logger.Error("SetMetrics Transaction error", zap.Error(err))
	}
}

// RunOnStart на старте запускаем миграции, запускаем тикер статистики пула соединений с бд
func (s *PgSQLStorage) RunOnStart() {
	s.runMigrations(s.config.Server.DBDSN)
	s.poolStatLogger(s.pgPool)
}

// RunOnClose закрываем pgPool
func (s *PgSQLStorage) RunOnClose() {
	if s.pgPool != nil {
		s.pgPool.Close()
	}
}

// runMigrations запускает миграции перед использованием приложения
func (s *PgSQLStorage) runMigrations(dsn string) {
	s.logger.Info("Start migration process")

	m, mErr := migrate.New(
		"file://internal/server/database/migrations",
		dsn)
	if mErr != nil {
		s.logger.Fatal("Can't prepare for migrations", zap.Error(mErr))
	}
	mErr = m.Up()
	if mErr != nil {
		s.logger.Debug("Migration complete with error:", zap.Error(mErr))
	} else {
		s.logger.Info("Successfully migrated")
	}
}

func (s *PgSQLStorage) poolStatLogger(pgPool *pgxpool.Pool) {
	statInterval := 30 * time.Second

	// запускаем тикер дампа статистики пула соединений с БД
	dumpPGPoolStatTicker := time.NewTicker(statInterval)

	s.logger.Info("Start pgPool stat collection", zap.Duration("statInterval", statInterval))

	go func() {
		for statTime := range dumpPGPoolStatTicker.C {
			stat := pgPool.Stat()
			s.logger.Debug("PGPool stat",
				zap.Time("datetime", statTime),
				zap.Int32("TotalConns", stat.TotalConns()),
				zap.Int32("AcquiredConns", stat.AcquiredConns()),
				zap.Int32("IdleConns", stat.IdleConns()),
				zap.Int64("NewConnsCount", stat.NewConnsCount()),
				zap.Int32("MaxConns", stat.MaxConns()),
			)
		}
	}()
}

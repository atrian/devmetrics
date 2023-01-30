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

	"github.com/atrian/devmetrics/internal/appconfig/serverconfig"
	"github.com/atrian/devmetrics/internal/dto"
	"github.com/atrian/devmetrics/pkg/logger"
)

// PoolStatInterval интервал вывода в лог статистики работы pgxpool
const PoolStatInterval = 30 * time.Second

// PgSQLStorage PostgreSQL хранилище для метрик и счетчиков
type PgSQLStorage struct {
	pgPool  *pgxpool.Pool
	metrics *MetricsDicts
	config  *serverconfig.Config
	logger  logger.Logger
}

var _ Repository = (*PgSQLStorage)(nil)

// NewPgSQLStorage возвращает указатель на PgSQLStorage который сконфигурирован со всеми зависимостями
// для работы с БД используется pgxpool
func NewPgSQLStorage(config *serverconfig.Config, logger logger.Logger) (*PgSQLStorage, error) {
	dbPool, poolErr := pgxpool.Connect(context.Background(), config.Server.DBDSN)
	if poolErr != nil {
		logger.Error("NewPgSQLStorage pgxpool.Connect", poolErr)
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

// StoreGauge сохранение метрики в БД
func (s *PgSQLStorage) StoreGauge(name string, value float64) error {
	_, err := s.pgPool.Exec(context.Background(), upsertMetricQuery(), name, "gauge", nil, value)
	if err != nil {
		s.logger.Error("StoreGauge pgPool.Exec", err)
		return fmt.Errorf(`failed store gauge: %w`, err)
	}
	return nil
}

// StoreCounter сохранение счетчика в БД
func (s *PgSQLStorage) StoreCounter(name string, value int64) error {
	// Проверяем есть ли уже счетчик в базе, если есть, суммируем данные
	storedCounter, exist := s.GetCounter(name)
	if exist {
		value += storedCounter
	}

	// Порядок аргументов id, type, delta, value
	_, err := s.pgPool.Exec(context.Background(), upsertMetricQuery(), name, "counter", value, nil)
	if err != nil {
		s.logger.Error("StoreCounter pgPool.Exec", err)
		return fmt.Errorf(`failed store counter: %w`, err)
	}

	return nil
}

// GetGauge получение метрики по имени
func (s *PgSQLStorage) GetGauge(name string) (float64, bool) {
	var value float64

	sqlQuery := `SELECT value FROM public.metrics WHERE id=$1 AND type='gauge';`
	row := s.pgPool.QueryRow(context.Background(), sqlQuery, name)

	switch err := row.Scan(&value); err {
	case nil:
		return value, true
	default:
		s.logger.Debug(fmt.Sprintf("GetGauge row.Scan: %v", err))
		return value, false
	}
}

// GetCounter получение счетчика по имени
func (s *PgSQLStorage) GetCounter(name string) (int64, bool) {
	var delta int64

	sqlQuery := `SELECT delta FROM public.metrics WHERE id=$1 AND type='counter';`
	row := s.pgPool.QueryRow(context.Background(), sqlQuery, name)

	switch err := row.Scan(&delta); err {
	case nil:
		return delta, true
	default:
		s.logger.Debug(fmt.Sprintf("GetCounter row.Scan: %v", err))
		return delta, false
	}
}

// GetMetrics получение всех метрик и счетчиков из БД в структуре MetricsDicts
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
		s.logger.Error("GetMetrics pgPool.Query", err)
		return s.metrics
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&metricID, &metricType, &delta, &value)
		if err != nil {
			s.logger.Error("GetMetrics rows.Scan", err)
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
		s.logger.Error("SetMetrics pgPool.Begin", err)
	}

	batch := &pgx.Batch{}
	for _, metric := range metrics {
		switch metric.MType {
		case "counter":
			// получаем сохраненное ранее в БД значение
			storedCounter, _ := s.GetCounter(metric.ID)
			// получаем сохраненное ранее значение в памяти в рамках одного batch запроса
			memoryCounter := memoryCounters[metric.ID]

			s.logger.Debug(fmt.Sprintf("SetMetrics counter value update: storedCounter %v, memoryCounter: %v, *metric.Delta: %v, position sum: %v",
				storedCounter, memoryCounter, *metric.Delta, storedCounter+memoryCounter+*metric.Delta))

			batch.Queue(upsertMetricQuery(), metric.ID, metric.MType, *metric.Delta+storedCounter+memoryCounter, nil)
			// обновляем сумму в памяти
			memoryCounters[metric.ID] += *metric.Delta
		case "gauge":

			s.logger.Debug(fmt.Sprintf("SetMetrics gauge value update: %v", *metric.Value))

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
		s.logger.Error("SetMetrics Transaction error", err)
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
		s.logger.Fatal("Can't prepare for migrations", mErr)
	}
	mErr = m.Up()
	if mErr != nil {
		s.logger.Debug(fmt.Sprintf("Migration complete with error: %v", mErr))
	} else {
		s.logger.Info("Successfully migrated")
	}
}

// poolStatLogger логирование статистики работы pgxpool
func (s *PgSQLStorage) poolStatLogger(pgPool *pgxpool.Pool) {
	// запускаем тикер дампа статистики пула соединений с БД
	dumpPGPoolStatTicker := time.NewTicker(PoolStatInterval)

	s.logger.Info(fmt.Sprintf("Start pgPool stat collection. statInterval: %v", PoolStatInterval))

	go func() {
		for statTime := range dumpPGPoolStatTicker.C {
			stat := pgPool.Stat()

			s.logger.Debug(fmt.Sprintf("PGPool stat: %v :: TotalConns: %v, AcquiredConns: %v, IdleConns: %v, NewConnsCount: %v, MaxConns: %v.",
				statTime,
				stat.TotalConns(),
				stat.AcquiredConns(),
				stat.IdleConns(),
				stat.NewConnsCount(),
				stat.MaxConns()))
		}
	}()
}

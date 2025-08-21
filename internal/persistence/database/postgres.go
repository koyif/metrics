package database

import (
	"context"
	"github.com/koyif/metrics/internal/app/logger"
	models "github.com/koyif/metrics/internal/model"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

type Database struct {
	conn *pgx.Conn
}

func New(ctx context.Context, url string) *Database {
	if url == "" {
		log.Fatal("database url is empty")
	}

	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	return &Database{
		conn: conn,
	}
}

func (db *Database) StoreMetric(metric models.Metrics) error {
	sql := "INSERT INTO metrics (metric_name, metric_type, metric_value, metric_delta, updated_at) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (metric_name) DO UPDATE SET metric_value = $3, metric_delta = $4, updated_at = $5"

	_, err := db.conn.Exec(context.Background(), sql, metric.ID, metric.MType, metric.Value, metric.Delta, time.Now())
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) Metric(metricName string) (models.Metrics, error) {
	sql := "SELECT metric_name, metric_type, metric_value, metric_delta FROM metrics WHERE metric_name = $1"
	var metric models.Metrics
	row := db.conn.QueryRow(context.Background(), sql, metricName)

	if err := row.Scan(&metric.ID, &metric.MType, &metric.Value, &metric.Delta); err != nil {
		return metric, err
	}

	return metric, nil
}

func (db *Database) AllMetrics() []models.Metrics {
	sql := "SELECT metric_name, metric_type, metric_value, metric_delta FROM metrics"
	rows, err := db.conn.Query(context.Background(), sql)
	if err != nil {
		logger.Log.Error("failed to query all metrics: %v", logger.Error(err))
		return nil
	}

	defer rows.Close()

	metrics := make([]models.Metrics, 0)
	for rows.Next() {
		var metric models.Metrics
		if err := rows.Scan(&metric.ID, &metric.MType, &metric.Value, &metric.Delta); err != nil {
			logger.Log.Error("failed to scan metric: %v", logger.Error(err))
			continue
		}
		metrics = append(metrics, metric)
	}

	return metrics
}

func (db *Database) Ping(ctx context.Context) error {
	return db.conn.Ping(ctx)
}

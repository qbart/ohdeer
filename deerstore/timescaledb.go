package deerstore

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/qbart/ohdeer/deer"
	"github.com/qbart/ohowl/tea"
)

type TimescaleDB struct {
	db       *sql.DB
	inserter *sql.Stmt
}

func NewTimescaleDB(ctx context.Context, connUri string) (*TimescaleDB, error) {
	db, err := sql.Open("postgres", connUri)
	if err != nil {
		return nil, fmt.Errorf("DB error: %v\n", err)
	}
	if err != nil {
		db.Close()
		return nil, err
	}

	return &TimescaleDB{
		db: db,
	}, nil
}

func (m *TimescaleDB) Migrate(ctx context.Context) error {
	sql := `
	CREATE TABLE IF NOT EXISTS metrics(
	  id         bigint        GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
	  monitor_id varchar       NOT NULL,
	  service_id varchar       NOT NULL,
	  at         timestamptz   NOT NULL,
	  success    bool          NOT NULL DEFAULT false,
	  details    jsonb
	);
	`
	_, err := m.db.Exec(sql)

	if err != nil {
		return err
	}

	inserter, err := m.db.Prepare(
		`INSERT INTO metrics(monitor_id, service_id, at, success, details) VALUES ($1, $2, $3, $4, $5)`,
	)
	if err != nil {
		return err
	}
	m.inserter = inserter

	return nil
}

func (m *TimescaleDB) Close(ctx context.Context) {
	if m.inserter != nil {
		m.inserter.Close()
	}
	m.db.Close()
}

func (m *TimescaleDB) Save(ctx context.Context, result *deer.CheckResult) {
	var d deer.Details
	d.Trace = result.Trace

	m.inserter.Exec(
		result.MonitorID,
		result.ServiceID,
		result.At,
		result.Success,
		tea.MustJson(d),
	)
}

func (m *TimescaleDB) Read(ctx context.Context) ([]*deer.Metric, error) {
	queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	rows, err := m.db.QueryContext(queryCtx, metricsSql)
	if err != nil {
		return nil, err
	}
	res := make([]*deer.Metric, 0, 24)

	for rows.Next() {
		var metric deer.Metric
		if err := rows.Scan(
			&metric.MonitorID,
			&metric.ServiceID,
			&metric.Bucket,
			&metric.Health,
		); err != nil {
			return nil, err
		}
		res = append(res, &metric)
	}

	return res, nil
}

const metricsSql string = `
WITH data AS (
  SELECT * FROM metrics WHERE at >= NOW() - INTERVAL '24 hours'
), series_by_service AS (
SELECT
  monitor_id,
  service_id,
  time_bucket('1 hour', at) AS bucket,
  count(*) filter (where success is true) as healthy_total,
  count(*) as total
FROM data
GROUP BY monitor_id, service_id, bucket
), series_by_monitor AS (
SELECT
  monitor_id,
  null AS service_id,
  time_bucket('1 hour', at) AS bucket,
  count(*) filter (where success is true) as healthy_total,
  count(*) as total
FROM data
GROUP BY monitor_id, bucket
)
  SELECT monitor_id, service_id, bucket, (healthy_total::numeric / total) AS health FROM series_by_monitor
UNION ALL
  SELECT monitor_id, service_id, bucket, (healthy_total::numeric / total) AS health FROM series_by_service
ORDER BY monitor_id, bucket, service_id NULLS FIRST
`

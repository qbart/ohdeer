package deerstore

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq" // postgres adapter
	"github.com/qbart/ohdeer/deer"
	"github.com/qbart/ohtea/tea"
)

// TimescaleDB store impl.
type TimescaleDB struct {
	db       *sql.DB
	inserter *sql.Stmt
}

// NewTimescaleDB creates new timescale db store.
func NewTimescaleDB(ctx context.Context, connURI string) (*TimescaleDB, error) {
	db, err := sql.Open("postgres", connURI)
	if err != nil {
		return nil, fmt.Errorf("DB error: %v", err)
	}
	if err != nil {
		db.Close()
		return nil, err
	}

	return &TimescaleDB{
		db: db,
	}, nil
}

// Migrate creates metrics table and initialize prepared statements.
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

// Truncate purges data from metrics table.
func (m *TimescaleDB) Truncate(ctx context.Context) error {
	_, err := m.db.Query("DELETE FROM metrics")
	return err
}

// Close closes connection to pg.
func (m *TimescaleDB) Close(ctx context.Context) {
	if m.inserter != nil {
		m.inserter.Close()
	}
	m.db.Close()
}

// Save inserts metrics to database.
func (m *TimescaleDB) Save(ctx context.Context, result *deer.CheckResult) {
	var d deer.Details
	d.Trace = result.Trace

	if result.StatusCode != 0 {
		d.Response = &deer.ResponseDetails{StatusCode: result.StatusCode}
	}

	if result.Error != nil {
		d.Error = &deer.ErrorDetails{Message: result.Error.Error()}
	}

	m.inserter.Exec(
		result.MonitorID,
		result.ServiceID,
		result.At,
		result.Success,
		tea.MustJson(d),
	)
}

// Read fetches metrics from database based on filter.
func (m *TimescaleDB) Read(ctx context.Context, filter *deer.ReadFilter) ([]*deer.Metric, error) {
	queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	intervalStart := filter.Since
	intervalStop := filter.Until()

	bucket := fmt.Sprint(filter.TimeBucket, " ", filter.TimeBucketUnit)

	var sb strings.Builder
	mi := 0
	for k, v := range filter.ActiveServices {
		sb.WriteString("(monitor_id =")
		sb.WriteString(pq.QuoteLiteral(k))
		if len(v) > 0 {
			sb.WriteString(" AND service_id IN (")
			for i, s := range v {
				sb.WriteString(pq.QuoteLiteral(s))
				if i < len(v)-1 {
					sb.WriteString(",")
				}
			}
			sb.WriteString(")")
		}
		sb.WriteString(")")
		if mi < len(filter.ActiveServices)-1 {
			sb.WriteString(" OR ")
		}
		mi++
	}
	if sb.Len() == 0 {
		sb.WriteString("1=1")
	}
	sql := fmt.Sprintf(
		metricsSQL,
		pq.QuoteLiteral(bucket),
		pq.QuoteLiteral(intervalStart.Format(time.RFC3339)),
		pq.QuoteLiteral(intervalStop.Format(time.RFC3339)),
		pq.QuoteLiteral(intervalStart.Format(time.RFC3339)),
		pq.QuoteLiteral(intervalStop.Format(time.RFC3339)),
		sb.String(),
	)

	// fmt.Println(sql)

	rows, err := m.db.QueryContext(queryCtx, sql)
	if err != nil {
		return nil, err
	}
	res := make([]*deer.Metric, 0, 24)

	var (
		dnsLookup, tcpConnection, tlsHandshake, serverProcessing, contentTransfer, total *float64
	)

	for rows.Next() {
		var metric deer.Metric
		metric.Details.Trace = &deer.Trace{}

		if err := rows.Scan(
			&metric.MonitorID,
			&metric.ServiceID,
			&metric.Bucket,
			&metric.Health,
			&metric.PassedChecks,
			&metric.FailedChecks,
			&dnsLookup,
			&tcpConnection,
			&tlsHandshake,
			&serverProcessing,
			&contentTransfer,
			&total,
		); err != nil {
			return nil, err
		}
		unit := time.Microsecond
		if dnsLookup != nil {
			metric.Details.Trace.DNSLookup = time.Duration(*dnsLookup) / unit
		}
		if tcpConnection != nil {
			metric.Details.Trace.TCPConnection = time.Duration(*tcpConnection) / unit
		}
		if tlsHandshake != nil {
			metric.Details.Trace.TLSHandshake = time.Duration(*tlsHandshake) / unit
		}
		if serverProcessing != nil {
			metric.Details.Trace.ServerProcessing = time.Duration(*serverProcessing) / unit
		}
		if contentTransfer != nil {
			metric.Details.Trace.ContentTransfer = time.Duration(*contentTransfer) / unit
		}
		if total != nil {
			metric.Details.Trace.Total = time.Duration(*total) / unit
		}
		res = append(res, &metric)
	}

	return res, nil
}

const metricsSQL string = `
SELECT
  monitor_id,
  service_id,
  time_bucket_gapfill(%s, at, %s, %s) AS bucket,
  COALESCE(count(*) FILTER (WHERE success IS true) / count(*)::numeric, -1) AS health,
  COALESCE(count(*) FILTER (WHERE success IS true), 0) AS passed_checks,
  COALESCE(count(*) FILTER (WHERE success IS false), 0) AS failed_checks,
  AVG((details->'trace'->>'dns_lookup')::numeric) AS dns_lookup,
  AVG((details->'trace'->>'tcp_connection')::numeric) AS tcp_connection,
  AVG((details->'trace'->>'tls_handshake')::numeric) AS tls_handshake,
  AVG((details->'trace'->>'server_processing')::numeric) AS server_processing,
  AVG((details->'trace'->>'content_transfer')::numeric) AS content_transfer,
  AVG((details->'trace'->>'total')::numeric) AS total
FROM metrics
WHERE (at BETWEEN %s AND %s) AND %s
GROUP BY monitor_id, service_id, bucket
ORDER BY monitor_id, service_id, bucket
`

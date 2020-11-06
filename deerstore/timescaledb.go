package deerstore

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/qbart/ohdeer/deer"
)

type TimescaleDB struct {
	conn *pgxpool.Pool
	m    sync.RWMutex
}

func NewTimescaleDB(connUri string) (*TimescaleDB, error) {
	pool, err := pgxpool.Connect(context.Background(), connUri)
	if err != nil {
		return nil, fmt.Errorf("DB error: %v\n", err)
	}
	return &TimescaleDB{
		conn: pool,
	}, nil
}

func (m *TimescaleDB) Init() {
}

func (m *TimescaleDB) Close() {
	m.conn.Close()
}

func (m *TimescaleDB) Save(result *deer.CheckResult) {
	m.m.Lock()
	defer m.m.Unlock()

	fmt.Println(result)
}

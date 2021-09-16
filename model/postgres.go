package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"

	"github.com/jackc/pgx/v4"
	"github.com/javking07/toadlester/conf"
)

type PostgresStorage struct {
	databaseConn *pgx.Conn
	dbName       string
}

func BootstrapPostgres(config *conf.DatabaseConfig) (PostgresStorage, error) {
	// connect to database
	dbInfo := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DatabaseName,
	)
	log.Info().Msgf("attempting to connect to database with info %s", dbInfo)
	conn, err := pgx.Connect(context.Background(), dbInfo)
	if err != nil {
		return PostgresStorage{}, err
	} else {
		log.Info().Msgf("connected to database on host %s:%v", config.Host, config.Port)
	}

	if conn == nil {
		return PostgresStorage{}, fmt.Errorf("no database available: %v", conn)
	}
	_, err = conn.Exec(context.Background(), CreateTableQuery)
	if err != nil {
		return PostgresStorage{}, err
	}

	return PostgresStorage{conn, config.DatabaseName}, nil
}

func (p PostgresStorage) Init(query string) error {
	if p.databaseConn == nil {
		return fmt.Errorf("no databse available: %v", p.databaseConn)
	}
	_, err := p.databaseConn.Exec(context.Background(), query)
	if err != nil {
		return err
	}
	return nil
}

func (p PostgresStorage) Insert(id string, itemName string, payload []byte) (int64, error) {
	query := `INSERT INTO tests (id, name,data) VALUES ($1,$2, $3)`
	result, err := p.databaseConn.Exec(context.Background(), query, id, itemName, payload)

	if err != nil {
		return 0, err
	}

	if affected := result.RowsAffected(); affected <= 1 {
		return affected, err
	} else {
		return affected, nil
	}
}

func (p PostgresStorage) Select(itemId int) ([]byte, error) {
	var data Payload
	err := p.databaseConn.QueryRow(context.Background(), `SELECT id, name, data FROM tests WHERE id=$1`, itemId).Scan(&data.ID, &data.Name, &data.Data)
	if err != nil {
		return nil, err
	}

	var items []Payload
	items = append(items, data)

	return json.Marshal(items)
}

func (p PostgresStorage) SelectAll(count, start int) ([]byte, error) {
	rows, err := p.databaseConn.Query(context.Background(), "SELECT id, name, data FROM tests LIMIT $1 OFFSET $2", count, start)
	if err != nil {
		return nil, err
	}
	defer func() { rows.Close() }()

	var payload []Payload
	for rows.Next() {
		var item Payload
		err := rows.Scan(&item.ID, &item.Name, &item.Data)
		if err != nil {
			return nil, err
		}
		payload = append(payload, item)
	}

	if len(payload) == 0 {
		return nil, sql.ErrNoRows
	}
	return json.Marshal(payload)
}

func (p PostgresStorage) Update(id int, payload Payload) error {
	_, err := p.databaseConn.Exec(context.Background(), "UPDATE tests SET name=$1, data=$2 WHERE id=$3", payload.Name, payload.Data, id)
	return err
}

func (p PostgresStorage) Delete(id int) error {
	_, err := p.databaseConn.Exec(context.Background(), "DELETE FROM tests where id=$1", id)
	return err
}

func (p PostgresStorage) Healthy() error {
	err := p.databaseConn.Ping(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (p PostgresStorage) Purge(table string) error {
	if _, err := p.databaseConn.Exec(context.Background(), fmt.Sprintf("DELETE FROM %s", table)); err != nil {
		return fmt.Errorf("Error purging %s table: %v", table, err)
	}
	log.Info().Msgf("Purging %s table", table)
	if _, err := p.databaseConn.Exec(context.Background(), fmt.Sprintf("ALTER SEQUENCE %s_id_seq RESTART WITH 1", table)); err != nil {
		return err
	}
	return nil
}

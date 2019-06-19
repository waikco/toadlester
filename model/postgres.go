package model

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"

	"github.com/javking07/toadlester/conf"
)

type PostgresStorage struct {
	database *sql.DB
	dbName   string
}

func BootstrapPostgres(config *conf.DatabaseConfig) (Storage, error) {
	// connect to database
	dbInfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		config.User, config.Password, config.DatabaseName)
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return nil, err
	}

	// return db connection
	storage := &PostgresStorage{db, config.DatabaseName}
	err = storage.Init(CreateTableQuery)
	if err != nil {
		return nil, err
	} else {
		log.Info().Msg("table presence confirmed")
	}
	return storage, nil
}

func (p *PostgresStorage) Init(query string) error {
	_, err := p.database.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresStorage) Insert(itemName string, payload []byte) (int64, error) {
	query := `INSERT INTO tests (name,data) VALUES ($1,$2)`
	result, err := p.database.Exec(query, itemName, payload)

	if err != nil {
		return 0, err
	}

	if id, err := result.RowsAffected(); err != nil {
		return 0, err
	} else {
		return id, nil
	}
}

func (p *PostgresStorage) Select(itemId int) ([]byte, error) {
	var data Payload
	err := p.database.QueryRow(`SELECT id, name, data FROM tests WHERE id=$1`, itemId).Scan(&data.ID, &data.Name, &data.Data)
	if err != nil {
		return nil, err
	}

	var items []Payload
	items = append(items, data)

	return json.Marshal(items)
}

func (p *PostgresStorage) SelectAll(count, start int) ([]byte, error) {
	rows, err := p.database.Query("SELECT id, name, data FROM tests LIMIT $1 OFFSET $2", count, start)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

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

func (p *PostgresStorage) Update(id int, payload Payload) error {
	_, err := p.database.Exec("UPDATE tests SET name=$1, data=$2 WHERE id=$3", payload.Name, payload.Data, id)
	return err
}

func (p *PostgresStorage) Delete(id int) error {
	_, err := p.database.Exec("DELETE FROM tests where id=$1", id)
	return err
}

func (p *PostgresStorage) Healthy() error {
	err := p.database.Ping()
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresStorage) Purge(table string) error {
	if _, err := p.database.Exec(fmt.Sprintf("DELETE FROM %s", table)); err != nil {
		return fmt.Errorf("Error purging %s table: %v", table, err)
	}
	log.Info().Msgf("Purging %s table", table)
	p.database.Exec(fmt.Sprintf("ALTER SEQUENCE %s_id_seq RESTART WITH 1", table))
	return nil
}

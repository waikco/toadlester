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

func (p *PostgresStorage) Insert(itemName string, payload []byte) error {
	query := `INSERT INTO tests (name,data) VALUES ($1,$2)`
	_, err := p.database.Query(query, itemName, payload)

	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresStorage) SelectAll(count, start int) ([]byte, error) {
	rows, err := p.database.Query("SELECT id, name, data FROM tests LIMIT  $1 OFFSET $2", count, start)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []interface{}
	for rows.Next() {
		var item interface{}
		err := rows.Scan(&item)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return json.Marshal(items)
}

func (p *PostgresStorage) Select(itemId int) ([]byte, error) {

	// Query for data
	query := `SELECT data FROM tests WHERE name=$1`
	rows, err := p.database.Query(query, itemID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var items []interface{}
	for rows.Next() {
		var item interface{}
		err := rows.Scan(&item)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return json.Marshal(items)
}

func (p *PostgresStorage) Update(itemName string, payload []byte) error {
	return nil
}

func (p *PostgresStorage) Delete(itemName string) error {
	return nil
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
		log.Fatal().Msgf("Error purging %s table: %s", table, err.Error())
	}
	log.Printf("Purging %s table", table)
	p.database.Exec(fmt.Sprintf("ALTER SEQUENCE %s_id_seq RESTART WITH 1", table))
	return nil
}

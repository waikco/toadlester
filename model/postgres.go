package model

import (
	"database/sql"
	"fmt"

	"github.com/rs/zerolog/log"

	_ "github.com/lib/pq"

	"github.com/javking07/toadlester/conf"
)

type PostgresStorage struct {
	database *sql.DB
	dbName   string
}

func BootstrapPostgres(config *conf.DatabaseConfig) (Storage, error) {

	// connect to database
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		config.User, config.Password, config.DatabaseName)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		return nil, err
	}

	// todo create required tables if they do not exist
	_, err = db.Exec(CreateTableQuery)
	if err != nil {
		return nil, err
	} else {
		log.Info().Msg("table presence confirmed")
	}

	// return db connection
	storage := &PostgresStorage{db, config.DatabaseName}
	return storage, nil
}

func (p *PostgresStorage) Insert(itemName string, payload []byte) error {
	query := `INSERT INTO tests (name,data) VALUES ($1,$2)`
	_, err := p.database.Query(query, itemName, payload)

	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresStorage) Select(itemName string) error {

	// Query for data
	//query := fmt.Sprintf("SELECT data FROM %s where name=?", table)
	//row, err := session.Query(query, d.Name)
	//
	//// If no records found, keep Found field as false and return
	//if iterable.NumRows() > 0 {
	//	d.Found = true
	//} else {
	//	return nil
	//}
	//
	//// If records found, add to Data field
	//m := map[string]interface{}{}
	//for iterable.MapScan(m) {
	//	d.Data = append(d.Data, m["data"])
	//	// clear map as required by gocql package
	//	m = map[string]interface{}{}
	//}
	return nil
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

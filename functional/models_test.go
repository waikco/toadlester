package functional_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	uuid "github.com/satori/go.uuid"

	"github.com/javking07/toadlester/model"

	"github.com/javking07/toadlester/app"
	"github.com/javking07/toadlester/conf"
	"github.com/rs/zerolog/log"
)

var a app.App
var config *conf.Config

func TestMain(m *testing.M) {

	switch os.Getenv("CONFIG_SWITCH") {
	case "drone":
		config = conf.SaneDefaults()
		log.Info().Msg("Configuring app for drone")
		config.Database.Host = "database"
		config.Database.DatabaseName = "test"
		config.Database.User = "postgres"
		config.Database.Password = "\"\""
		config.Database.SslMode = "disable"
	case "local":
		config = conf.SaneDefaults()
	default:
		config = conf.SaneDefaults()
	}

	a.Bootstrap(config)

	log.Info().Msgf("database ssl status is %s", config.Database.SslMode)

	// table confirmation and cleanup
	defer purgeTable()

	addData(6)

	code := m.Run()

	os.Exit(code)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func TestConfirmTable(t *testing.T) {
	log.Print("confirming table exists...")
	if err := a.Storage.Init(model.TestCreateTableQuery); err != nil {
		t.Fatalf("Error creating tests table: " + err.Error())
	}
}

//purgeTable deletes items from table
func purgeTable() {
	if err := a.Storage.Purge("tests"); err != nil {
		log.Error().Msg(err.Error())
	}
}

// addData adds dummy entries to database
func addData(count int) {
	if count < 1 {
		count = 1
	}
	for i := 0; i < count; i++ {
		_, err := a.Storage.Insert(uuid.NewV4().String(), strconv.Itoa(i), []byte(fmt.Sprintf(`{"tps": 100, "url": "http://example.com", "name": "today", "method": "GET", "duration": "10s"}`)))
		if err != nil {
			log.Fatal().Msgf("error adding data: %v", err)
		}
	}
}

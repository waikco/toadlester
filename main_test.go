package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

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
	case "local":
		config = conf.SaneDefaults()
	default:
		config = conf.SaneDefaults()
	}

	a.AppConfig = config
	log.Info().Msgf("%+v", a.AppConfig)
	a.Bootstrap()

	// fallback to default config, if file, or env vars not found
	if config.Server == nil {
		log.Info().Msg("no viable config available. falling back to sane defaults.\n")
		a.AppConfig = conf.SaneDefaults()
	} else {
		a.AppConfig = config
	}

	log.Info().Msgf("database ssl status is %s", a.AppConfig.Database.SslMode)

	// table confirmation and cleanup
	confirmTableExists()
	defer purgeTable()

	code := m.Run()

	os.Exit(code)
}

func TestEmptyTable(t *testing.T) {
	purgeTable()

	req := httptest.NewRequest("GET", "/toadlester/v1/tests", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
	checkResponseBody(t, "[]", response.Body.String())

}

//ConfirmTableexists checks for existence of app database
func confirmTableExists() {
	log.Print("confirming table exists...")
	if err := a.AppStorage.Init(model.TestCreateTableQuery); err != nil {
		log.Fatal().Msgf("Error creating tests table: " + err.Error())
	}
	log.Print("database confirmed...")

}

func TestGetNonExistentTest(t *testing.T) {
	purgeTable()

	req := httptest.NewRequest("GET", "/toadlester/v1/13", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
	checkResponseBody(t, "[]", response.Body.String())
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.AppRouter.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected Status Code:%d | Received Status Code:%d", expected, actual)
	}
}

func checkResponseBody(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Errorf("Expected Body:%s | Received Body:%s", expected, actual)
	}
}

//PurgeTable deletes items from table
func purgeTable() {

}

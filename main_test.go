package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
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
	confirmTable()
	purgeTable()
	defer purgeTable()

	code := m.Run()

	os.Exit(code)
}

func TestEmptyTable(t *testing.T) {
	purgeTable()

	req := httptest.NewRequest("GET", "/toadlester/v1/tests", nil)
	response := executeRequest(req)

	if response.Code != http.StatusOK {
		t.Errorf("got %v want %v", response.Code, http.StatusOK)
	}

	if response.Body.String() != `"[]"` {
		t.Errorf("got %v want %v", response.Body.String(), `"[]"`)
	}
}

func TestGetNonExistentTest(t *testing.T) {
	purgeTable()

	req := httptest.NewRequest("GET", "/toadlester/v1/tests/3", nil)
	response := executeRequest(req)

	if response.Code != http.StatusOK {
		t.Errorf("got %v want %v", response.Code, http.StatusOK)
	}

	if response.Body.String() != `"[]"` {
		t.Errorf("got %v want %v", response.Body.String(), `"[]"`)
	}
}

func TestGet(t *testing.T) {
	purgeTable()
	addData(1)
	req := httptest.NewRequest("GET", "/toadlester/v1/tests/1", nil)
	response := executeRequest(req)

	if response.Code != http.StatusOK {
		t.Errorf("got %v want %v", response.Code, http.StatusOK)
	}

	expected := `[{"id":"1","name":"0","data":{"tps":100,"url":"http://example.com","name":"today","method":"GET","duration":"10s"}}]`
	if response.Body.String() != expected {
		t.Errorf("got %v want %v", response.Body.String(), expected)
	}
}

func TestGetAll(t *testing.T) {
	purgeTable()
	addData(2)
	req := httptest.NewRequest("GET", "/toadlester/v1/tests", nil)
	response := executeRequest(req)

	if response.Code != http.StatusOK {
		t.Errorf("got %v want %v", response.Code, http.StatusOK)
	}

	expected := `[{"id":"1","name":"0","data":{"tps":100,"url":"http://example.com","name":"today","method":"GET","duration":"10s"}},{"id":"2","name":"1","data":{"tps":100,"url":"http://example.com","name":"today","method":"GET","duration":"10s"}}]`
	if response.Body.String() != expected {
		t.Errorf("got %v want %v", response.Body.String(), expected)
	}
}

func TestUpdate(t *testing.T) {
	purgeTable()
	addData(1)
	payload := []byte(`{"id":"1","name":"updated","data":{"tps":200,"url":"http://example.com/updates","name":"updated","method":"POST","duration":"30s"}}`)
	req := httptest.NewRequest("PUT", "/toadlester/v1/tests/1", bytes.NewBuffer(payload))
	response := executeRequest(req)

	if response.Code != http.StatusOK {
		t.Errorf("got %v want %v", response.Code, http.StatusOK)
	}

	//expected := `[{"id":"1","name":"0","data":{"tps":100,"url":"http://example.com","name":"today","method":"GET","duration":"10s"}}]`
	if response.Body.String() != string(payload) {
		t.Errorf("got %v want %v", response.Body.String(), payload)
	}
}

func TestDelete(t *testing.T) {
	purgeTable()
	addData(1)

	// confirm presence
	req := httptest.NewRequest("GET", "/toadlester/v1/tests/1", nil)
	response := executeRequest(req)

	if response.Code != http.StatusOK {
		t.Errorf("got %v want %v", response.Code, http.StatusOK)
	}
	expected := `[{"id":"1","name":"0","data":{"tps":100,"url":"http://example.com","name":"today","method":"GET","duration":"10s"}}]`
	if response.Body.String() != expected {
		t.Errorf("got %v want %v", response.Body.String(), expected)
	}

	// delete and confirm result
	req = httptest.NewRequest("DELETE", "/toadlester/v1/tests/1", nil)
	response = executeRequest(req)

	if response.Code != http.StatusOK {
		t.Errorf("got %v want %v", response.Code, http.StatusOK)
	}
	expected = `{"result":"success"}`
	if response.Body.String() != expected {
		t.Errorf("got %v want %v", response.Body.String(), expected)
	}

	// confirm lack of presence
	req = httptest.NewRequest("GET", "/toadlester/v1/tests/1", nil)
	response = executeRequest(req)

	if response.Code != http.StatusOK {
		t.Errorf("got %v want %v", response.Code, http.StatusOK)
	}
	expected = `"[]"`
	if response.Body.String() != expected {
		t.Errorf("got %v want %v", response.Body.String(), expected)
	}

}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.AppRouter.ServeHTTP(rr, req)

	return rr
}

//ConfirmTable checks for existence of app database
func confirmTable() {
	log.Print("confirming table exists...")
	if err := a.AppStorage.Init(model.TestCreateTableQuery); err != nil {
		log.Fatal().Msgf("Error creating tests table: " + err.Error())
	}
	log.Print("database confirmed...")
}

//purgeTable deletes items from table
func purgeTable() {
	if err := a.AppStorage.Purge("tests"); err != nil {
		log.Error().Msg(err.Error())
	}
}

// addData adds dummy entries to database
func addData(count int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		a.AppStorage.Insert(strconv.Itoa(i), []byte(fmt.Sprintf(`{"tps": 100, "url": "http://example.com", "name": "today", "method": "GET", "duration": "10s"}`)))
	}
}

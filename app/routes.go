package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/javking07/toadlester/model"
)

func (a *App) Health(w http.ResponseWriter, r *http.Request) {
	var Health struct {
		ServerStatus   string `json:"server_status"`
		DatabaseStatus string `json:"database_status"`
		CacheCount     int64  `json:"cache_count"`
	}
	// Report on server status
	Health.ServerStatus = "ok"

	//Report on database status
	if a.AppStorage != nil {
		err := a.AppStorage.Healthy()

		if err == nil {
			Health.DatabaseStatus = "connected"
		} else {
			a.AppLogger.Error().Msg(err.Error())
			Health.DatabaseStatus = "not connected"
		}

	}

	// Report on cache status
	if a.AppCache != nil {
		Health.CacheCount = a.AppCache.EntryCount()
	} else {
		Health.CacheCount = 0
	}

	respondWithJSON(w, http.StatusOK, Health)

}

// todo create middleware for capturing trace data on request
func (a *App) WithTracing(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (a *App) PostTest(w http.ResponseWriter, r *http.Request) {
	var payload model.LoadTest
	requestBody, err := ioutil.ReadAll(r.Body)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error reading body")
		return
	}

	err = json.Unmarshal(requestBody, &payload)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("error unmarshalling body: %s", err.Error()))
		return
	}

	// add to database, then cache
	if err := a.AppStorage.Insert(payload.Name, requestBody); err != nil {
		respondWithError(w, http.StatusInternalServerError, "error setting db value"+err.Error())
		return
	} else {
		err = a.AppCache.Set([]byte(payload.Name), requestBody, 3600)
		if err != nil {
			a.AppLogger.Error().Msg(err.Error())
			respondWithError(w, http.StatusInternalServerError, "error setting cache value")
			return
		} else {
			respondWithJSON(w, http.StatusCreated, fmt.Sprintf(" test %s added to queu", payload.Name))
			return
		}
	}

}

// todo adjust interface to return data and finish this to get back data
func (a *App) GetTest(w http.ResponseWriter, r *http.Request) {

	var payload model.LoadTest
	var payloadBytes []byte

	// check cache first
	if cacheValue, err := a.AppCache.Get([]byte(payload.Name)); err == nil {
		payloadBytes = cacheValue
	} else if dbValue, err := a.AppStorage.Select(payload.Name); err == nil {
		payloadBytes = dbValue
	} else {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	respondWithJSON(w, http.StatusOK, payload)
}

func (a *App) GetTests(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}

	if start < 0 {
		start = 0
	}
	// todo flesh out get all tests methods for storage and routes
	tests, err := a.AppStorage.SelectAll(start, count)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	if len(tests) == 0 {
		respondWithJSON(w, http.StatusOK, `[]`)
	} else {
		respondWithJSON(w, http.StatusOK, tests)
	}

}

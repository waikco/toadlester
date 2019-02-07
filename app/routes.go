package app

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
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
		err := a.AppStorage.Ping()

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

	var payload TestInfo
	body := r.Body
	if reqbody, err := ioutil.ReadAll(body); err != nil {
		json.Unmarshal(reqbody, payload)
	}

}

// todo adjust interface to return data and finish this to get back data
func (a *App) GetTest(w http.ResponseWriter, r *http.Request) {

	var payload TestInfo
	body := r.Body
	if reqbody, err := ioutil.ReadAll(body); err != nil {
		json.Unmarshal(reqbody, payload)
	} else {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := a.AppStorage.Select(payload.name); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	} else {
		respondWithJSON(w, http.StatusOK, payload)
	}

}

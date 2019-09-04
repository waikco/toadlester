package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	uuid "github.com/satori/go.uuid"

	"github.com/rs/zerolog/log"

	"github.com/go-chi/chi"

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

func (a *App) PostTest(w http.ResponseWriter, r *http.Request) {
	var payload model.LoadTestSimple
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

	if id, err := a.AppStorage.Insert(uuid.NewV4().String(), payload.Name, requestBody); err != nil {
		respondWithError(w, http.StatusInternalServerError, "error setting db value"+err.Error())
		return
	} else {
		respondWithJSON(w, http.StatusCreated, fmt.Sprintf("test added to queue: %d", id))
		return

	}
}

func (a *App) GetTest(w http.ResponseWriter, r *http.Request) {
	idInt, err := strconv.Atoi(chi.URLParam(r, "testsID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("error formatting ID input: %v", idInt))
		log.Error().Msgf("error converting ID: %v", err.Error())
		return
	}

	switch dbValue, err := a.AppStorage.Select(idInt); err {
	case nil:
		var payload []model.Payload
		if err := json.Unmarshal(dbValue, &payload); err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		} else {
			respondWithJSON(w, http.StatusOK, payload)
		}
	case sql.ErrNoRows:
		respondWithJSON(w, http.StatusOK, `[]`)
	default:
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
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

	switch dbValue, err := a.AppStorage.SelectAll(count, start); err {
	case nil:
		var payload []model.Payload
		if err := json.Unmarshal(dbValue, &payload); err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		} else {
			respondWithJSON(w, http.StatusOK, payload)
		}
	case sql.ErrNoRows:
		respondWithJSON(w, http.StatusOK, `[]`)
	default:
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
}

func (a *App) UpdateTest(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "testsID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("error formatting ID input: %v", id))
		log.Error().Msgf("error converting ID: %v", err.Error())
		return
	}

	var p model.Payload
	if err = json.NewDecoder(r.Body).Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid resquest payload")
		return
	}
	defer func() { _ = r.Body.Close() }()

	switch err := a.AppStorage.Update(id, p); err {
	case nil:
		respondWithJSON(w, http.StatusOK, p)
	default:
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
}

func (a *App) DeleteTest(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "testsID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("error formatting ID input: %v", id))
		log.Error().Msgf("error converting ID: %v", err.Error())
		return
	}

	switch err := a.AppStorage.Delete(id); err {
	case nil:
		respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
	default:
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
}

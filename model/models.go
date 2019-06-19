package model

import (
	"encoding/json"
	"time"
)

type Payload struct {
	ID   string          `json:"id"`
	Name string          `json:"name"`
	Data json.RawMessage `json:"data"` // or could be []interface{}
	//Payload interface{} `json:"data"`
}

type LoadTest struct {
	Name     string `json:"name"`
	Method   string `json:"method"`
	Url      string `json:"url"`
	Duration string `json:"duration"` // in seconds
	TPS      int    `json:"tps"`
}

type LoadTestResults struct {
	Latencies struct {
		Total   int `json:"total"`
		Mean    int `json:"mean"`
		Five0Th int `json:"50th"`
		Nine5Th int `json:"95th"`
		Nine9Th int `json:"99th"`
		Max     int `json:"max"`
	} `json:"latencies"`
	BytesIn struct {
		Total int `json:"total"`
		Mean  int `json:"mean"`
	} `json:"bytes_in"`
	BytesOut struct {
		Total int `json:"total"`
		Mean  int `json:"mean"`
	} `json:"bytes_out"`
	Earliest    time.Time `json:"earliest"`
	Latest      time.Time `json:"latest"`
	End         time.Time `json:"end"`
	Duration    int       `json:"duration"`
	Wait        int       `json:"wait"`
	Requests    int       `json:"requests"`
	Rate        float64   `json:"rate"`
	Success     int       `json:"success"`
	StatusCodes struct {
		Num200 int `json:"200"`
	} `json:"status_codes"`
	Errors []interface{} `json:"errors"`
}

type Storage interface {
	Init(string) error
	Insert(string, []byte) (int64, error)
	Select(int) ([]byte, error)
	SelectAll(int, int) ([]byte, error)
	Update(int, []byte) error
	Delete(int) error
	Purge(string) error // deletes all items from table
	Healthy() error
}

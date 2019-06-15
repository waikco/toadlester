package model

import (
	"encoding/json"
	"time"
)

type Data struct {
	ID   string          `json:"id"`
	Name string          `json:"name"`
	Data json.RawMessage `json:"data"` // or could be []interface{}
	//Data []interface{} `json:"data"`
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
	Select(string) ([]byte, error)
	SelectAll(int, int) ([]byte, error)
	Insert(string, []byte) error
	Update(string, []byte) error
	Delete(string) error
	Purge(string) error // deletes all items from table
	Healthy() error
}

func (d *Data) Set(storage Storage, payload interface{}) error {
	err := storage.Insert(d.Name, d.Data)

	if err != nil {
		return err
	}
	return nil
}

func (d *Data) Get(storage Storage) ([]byte, error) {

	data, err := storage.Select(d.ID)

	if err != nil {
		return nil, err
	}
	return data, err
}

func (d *Data) GetAll(storage Storage, start, count int) ([]byte, error) {

	data, err := storage.SelectAll(start, count)

	if err != nil {
		return nil, err
	}
	return data, err
}

func (d *Data) Change(storage Storage) error {
	err := storage.Update(d.Name, d.Data)

	if err != nil {
		return err
	}
	return nil
}

func (d *Data) Remove(storage Storage) error {
	err := storage.Delete(d.Name)

	if err != nil {
		return err
	}
	return nil
}

package main

import (
	"encoding/json"
	"io/ioutil"
)

type record struct {
	User *user `json:"user"`
	Ads  []*ad `json:"ads"`
}

type user struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ad struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Price       string   `json:"price"`
	CategoryId  string   `json:"categoryId"`
	Images      []string `json:"images"`
}

func getRecords(filename string) ([]*record, error) {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	records := make([]record, 0)
	json.Unmarshal(raw, &records)

	recordsPtr := make([]*record, len(records))
	for i, r := range records {
		recordsPtr[i] = &r
	}

	return recordsPtr, nil
}

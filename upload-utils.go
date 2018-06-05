package main

import (
	"encoding/json"
	"io/ioutil"

	"bolha-utils/client"
)

type record struct {
	User user `json:"user"`
	Ads  []ad `json:"ads"`
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

func getRecords(filename string) ([]client.Record, error) {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	records := make([]record, 0)
	json.Unmarshal(raw, &records)

	clientRecords := make([]client.Record, len(records))
	for i, r := range records {
		castUser := client.User(r.User)
		castAds := make([]*client.Ad, len(r.Ads))
		for j, a := range r.Ads {
			castAd := client.Ad(a)
			castAds[j] = &castAd
		}

		newR := client.Record{
			User: &castUser,
			Ads:  castAds,
		}

		clientRecords[i] = newR
	}

	return clientRecords, nil
}

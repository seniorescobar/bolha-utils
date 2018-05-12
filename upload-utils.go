package main

import (
	"encoding/json"
	"io/ioutil"

	"bolha-utils/client"
)

type clientRecord struct {
	user *client.User
	ads  []*client.Ad
}

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

// func removeAds(users []*upload.User) {
// 	for _, u := range users {
// 		if err := upload.RemoveAllAds(u); err != nil {
// 			log.WithFields(log.Fields{
// 				"err": err,
// 			}).Error("error removing ads")
// 		}
// 	}
// }

// func uploadAds(records []*upload.Record) {
// 	var wg sync.WaitGroup
// 	wg.Add(len(ads))

// 	errChan := make(chan error)

// 	for _, ad := range ads {
// 		go func() {
// 			if err := upload.UploadAd(ad); err != nil {
// 				errChan <- err
// 			}
// 			wg.Done()
// 		}()
// 	}

// 	go func() {
// 		wg.Wait()
// 		close(errChan)
// 	}()

// 	for err := range errChan {
// 		log.Error(err)
// 	}
// }

func getRecords(filename string) ([]*clientRecord, error) {
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

	return mapRecords(recordsPtr), nil
}

func mapRecords(records []*record) []*clientRecord {
	clientRecords := make([]*clientRecord, len(records))

	for i, r := range records {
		clientRecords[i] = &clientRecord{
			user: &client.User{
				Username: r.User.Username,
				Password: r.User.Password,
			},
			ads: make([]*client.Ad, len(r.Ads)),
		}

		for j, ad := range r.Ads {
			clientRecords[i].ads[j] = &client.Ad{
				Title:       ad.Title,
				Description: ad.Description,
				Price:       ad.Price,
				CategoryId:  ad.CategoryId,
				Images:      ad.Images,
			}
		}
	}

	return clientRecords
}

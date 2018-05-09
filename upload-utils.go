package main

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/sirupsen/logrus"

	"bolha-utils/upload"
)

func uploadAds(ads []*upload.Ad) {
	var wg sync.WaitGroup
	wg.Add(len(ads))

	errChan := make(chan error)

	for _, ad := range ads {
		go func() {
			if err := upload.UploadAd(ad); err != nil {
				errChan <- err
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		log.Error(err)
	}
}

func getAds(filename string) ([]*upload.Ad, error) {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var ads []upload.Ad
	json.Unmarshal(raw, &ads)

	adsPtr := make([]*upload.Ad, len(ads))
	for i, ad := range ads {
		adsPtr[i] = &ad
	}

	return adsPtr, nil
}

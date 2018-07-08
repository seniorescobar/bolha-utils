package main

import (
	"bolha-utils/client"
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
)

var (
	uploadCommand *flag.FlagSet
	uploadFilePtr *string
)

func init() {
	uploadCommand = flag.NewFlagSet("upload", flag.ExitOnError)
	uploadFilePtr = uploadCommand.String("file", "", `JSON file containing ads to be upload to bolha.`)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("missing command argument")
	}

	switch os.Args[1] {
	case "upload":
		uploadCommand.Parse(os.Args[2:])
		if uploadCommand.Parsed() {
			records, err := getRecords(*uploadFilePtr)
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Fatal("error getting records")
			}

			uploadHelper(records)
		}
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
}

// UPLOAD
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
	for i := range records {
		recordsPtr[i] = &records[i]
	}

	return recordsPtr, nil
}

func uploadHelper(records []*record) {
	var wg sync.WaitGroup

	for _, r := range records {
		wg.Add(1)

		go func(r *record) {
			defer wg.Done()

			cUser := client.User(*r.User)
			c, err := client.New(&cUser)
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Error("error creating client")
				return
			}

			if err := c.RemoveAllAds(); err != nil {
				log.WithFields(log.Fields{"err": err}).Error("error removing all ads")
			}

			ads := make([]*client.Ad, len(r.Ads))
			for i, ad := range r.Ads {
				tmpAd := client.Ad(*ad)
				ads[i] = &tmpAd
			}

			c.UploadAds(ads)
		}(r)
	}

	wg.Wait()
}

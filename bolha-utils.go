package main

import (
	"bolha-utils/client"
	"bolha-utils/find"
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

			var wg sync.WaitGroup

			for _, r := range records {
				wg.Add(1)

				go func(r *record) {
					defer wg.Done()

					tmpUser := client.User(*r.User)
					c, err := client.New(&tmpUser)
					if err != nil {
						log.WithFields(log.Fields{"err": err}).Fatal("error creating client")
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

				wg.Wait()
			}
		}
	case "find":
		pattern := `(?s:<div class="ad">.+?title="(?P<Title>.*?)".+?href="(?P<Link>.+?)".+?<div class="price"><span>(?P<Price>\d+))`
		url := `http://www.bolha.com/racunalnistvo/igricarstvo-gaming/xbox/xbox-one/?location=Osrednjeslovenska%2F&hasImages=Oglasi+s+fotografijami&datePlaced=Zadnji+teden`

		bolhaScraper, err := find.New(pattern)
		if err != nil {
			log.Fatalln(err)
		}

		found, err := bolhaScraper.Find(url, func(*find.Ad) bool {
			return false
		})
		if err != nil {
			log.Fatalln(err)
		}

		for _, f := range found {
			log.Println(f)
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
	for i, r := range records {
		recordsPtr[i] = &r
	}

	return recordsPtr, nil
}

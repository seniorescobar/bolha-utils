package main

import (
	"bolha-utils/client"
	"flag"
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
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
}

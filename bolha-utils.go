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
			wg.Add(len(records))

			for _, r := range records {
				go func(wg *sync.WaitGroup, r client.Record) {
					defer wg.Done()

					c, err := client.New(r.User)
					if err != nil {
						log.WithFields(log.Fields{"err": err}).Fatal("error creating client")
					}

					if err := c.RemoveAllAds(); err != nil {
						log.WithFields(log.Fields{"err": err}).Error("error removing all ads")
					}

					c.UploadAds(r.Ads)
				}(&wg, r)
			}

			wg.Wait()
		}
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
}

package main

import (
	"bolha-utils/client"
	"flag"
	"os"

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

			for _, record := range records {
				tmpUser := client.User(*record.User)
				c, err := client.New(&tmpUser)
				if err != nil {
					log.WithFields(log.Fields{"err": err}).Fatal("error creating client")
				}

				if err := c.RemoveAllAds(); err != nil {
					log.WithFields(log.Fields{"err": err}).Error("error removing all ads")
				}

				ads := make([]*client.Ad, len(record.Ads))
				for i, ad := range ads {
					tmpAd := client.Ad(*ad)
					ads[i] = &tmpAd
				}

				c.UploadAds(ads)
			}
		}
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
}

package find

import (
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
)

type Scraper struct {
	regex *regexp.Regexp
}

type Ad struct {
	Title       string
	Description string
	Price       string
	Link        string
}

func New(pattern string) (*Scraper, error) {
	r, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	return &Scraper{
		regex: r,
	}, nil
}

func (s *Scraper) Find(url string, skipFunc func(*Ad) bool) ([]*Ad, error) {
	found := make([]*Ad, 0)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	matches := s.regex.FindAllStringSubmatch(string(respBody), -1)
	mapping := s.regex.SubexpNames()
	for _, m := range matches {
		ad := &Ad{}
		adVal := reflect.ValueOf(ad).Elem()

		for i, v := range m[1:] {
			name := mapping[i+1]
			field := adVal.FieldByName(name)
			if field.IsValid() {
				field.SetString(v)
			}
		}

		if skipFunc(ad) {
			continue
		}

		found = append(found, ad)
	}

	return found, nil
}

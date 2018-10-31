package client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	defaultHeaders = map[string][]string{
		"Accept":          []string{"application/json, text/javascript, */*; q=0.01"},
		"Accept-Encoding": []string{"identity"},
		"Accept-Language": []string{"en-US,en;q=0.9,sl;q=0.8,hr;q=0.7"},
		"Cache-Control":   []string{"max-age=0"},
		"Connection":      []string{"keep-alive"},
		"User-Agent":      []string{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/66.0.3359.139 Safari/537.36"},
	}
)

func (c *Client) String() string {
	return fmt.Sprint(c.username)
}

func (c *Client) newRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)

	// set default headers
	req.Header = http.Header(defaultHeaders)

	return req, err
}

func (c *Client) uploadAd(ad *Ad) error {
	log.WithFields(log.Fields{
		"ad": ad,
	}).Info("uploading ad...")

	metaInfo, err := c.getAdMetaInfo(ad)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{"metainfo": metaInfo}).Info("metainfo")

	return c.publishAd(ad, metaInfo)
}

func (c *Client) removeAds(ids []string) error {
	log.WithFields(log.Fields{
		"ids": ids,
	}).Info("removing ads...")

	values := url.Values{
		"IDS": {
			strings.Join(ids, ","),
		},
	}

	req, err := c.newRequest(http.MethodPost, "https://moja.bolha.com/adManager/ajaxRemoveActiveBulk", strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}

	addHeaders(req, map[string]string{
		"Content-Type":              "application/x-www-form-urlencoded",
		"Host":                      "moja.bolha.com",
		"Origin":                    "https://moja.bolha.com",
		"Referer":                   "https://moja.bolha.com/oglasi",
		"Upgrade-Insecure-Requests": "1",
		"X-Requested-With":          "XMLHttpRequest",
	})

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.New("ads not removed")
	}

	return nil
}

func getHttpClient() (*http.Client, error) {
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Timeout: time.Duration(3) * time.Minute,
		Jar:     cookieJar,
	}, nil
}

func (c *Client) logIn(username, password string) error {
	values := url.Values{
		"username": {
			username,
		},
		"password": {
			password,
		},
		"rememberMe": {
			"true",
		},
	}

	req, err := c.newRequest(http.MethodPost, "https://login.bolha.com/auth.php", strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}

	addHeaders(req, map[string]string{
		"Content-Type":              "application/x-www-form-urlencoded",
		"Host":                      "login.bolha.com",
		"Origin":                    "http://www.bolha.com",
		"Referer":                   "http://www.bolha.com/",
		"Upgrade-Insecure-Requests": "1",
		"X-Requested-With":          "XMLHttpRequest",
		"X-Site":                    "http://www.bolha.com/",
	})

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("login failed for '%s' ('%s')", username, password)
	}

	return nil
}

func (c *Client) getAdIds() ([]string, error) {
	req, err := c.newRequest(http.MethodGet, "https://moja.bolha.com/oglasi", nil)
	if err != nil {
		return nil, err
	}

	addHeaders(req, map[string]string{
		"Host": "moja.bolha.com",
		"Upgrade-Insecure-Requests": "1",
	})

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	r := regexp.MustCompile(`id="ad_(\d+)`)
	matches := r.FindAllStringSubmatch(string(body), -1)
	if matches == nil {
		return nil, errors.New("no ads found")
	}

	ids := make([]string, len(matches))
	for i, m := range matches {
		ids[i] = m[1]
	}

	return ids, nil
}

func (c *Client) getAdMetaInfo(ad *Ad) (map[string]string, error) {
	values := url.Values{
		"categoryId": {
			ad.CategoryId,
		},
	}

	req, err := c.newRequest(http.MethodPost, "http://objava-oglasa.bolha.com/izbor_paketa.php", strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	addHeaders(req, map[string]string{
		"Content-Type":              "application/x-www-form-urlencoded",
		"Host":                      "objava-oglasa.bolha.com",
		"Origin":                    "http://objava-oglasa.bolha.com",
		"Referer":                   "http://objava-oglasa.bolha.com/",
		"Upgrade-Insecure-Requests": "1",
	})

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	regex := map[string]*regexp.Regexp{
		"submitTakoj":         regexp.MustCompile(`<input type="hidden" name="submitTakoj" id="submitTakoj" value="(.*?)" />`),
		"listItemId":          regexp.MustCompile(`<input type="hidden" name="listItemId" id="listItemId" value="(.*?)" />`),
		"lPreverjeni":         regexp.MustCompile(`<input type="hidden" name="lPreverjeni" id="lPreverjeni" value="(.*?)" />`),
		"lShop":               regexp.MustCompile(`<input type="hidden" name="lShop" id="lShop" value="(.*?)">`),
		"uploader_id":         regexp.MustCompile(`<input type="hidden" name="uploader_id" id="uploader_id" value="(.*?)" />`),
		"novo":                regexp.MustCompile(`<input type="hidden" name="novo" value="(.*?)" />`),
		"adPlacementPrice":    regexp.MustCompile(`<input type="hidden" name="adPlacementPrice" id="adPlacementPrice" value="(.*?)" />`),
		"adPlacementDiscount": regexp.MustCompile(`<input type="hidden" name="adPlacementDiscount" id="adPlacementDiscount" value="(.*?)" />`),
		"nDays":               regexp.MustCompile(`<input type="hidden" name="nDays" value="(.*?)" />`),
		"spremeni":            regexp.MustCompile(`<input type="hidden" name="spremeni" value="(.*?)" />`),
		"new":                 regexp.MustCompile(`<input type="hidden" name="new" value="(.*?)" />`),
		"nKatID":              regexp.MustCompile(`<input name="nKatID" id="nKatID" type="hidden" size="5" value="(.*?)" />`),
		"nNadKatID":           regexp.MustCompile(`<input name="nNadKatID" id="nNadKatID" type="hidden" size="5" value="(.*?)" />`),
		"nMainKatID":          regexp.MustCompile(`<input name="nMainKatID" id="nMainKatID" type="hidden" size="5" value="(.*?)" />`),
		"nPath":               regexp.MustCompile(`<input name="nPath" id="nPath" disable="false" type="hidden" value="(.*?)" />`),
		"nHide":               regexp.MustCompile(`<input name="nHide" id="nHide" type="hidden" value="(.*?)" />`),
		"nPrekrij":            regexp.MustCompile(`<input style="display:none;" type="hidden" name="nPrekrij" value="(.*?)" />`),
		"nStep":               regexp.MustCompile(`<input style="display:none;" type="hidden" name="nStep" value="(.*?)" />`),
		"lNonJava":            regexp.MustCompile(`<input style="display:none;" type="hidden" name="lNonJava" value="(.*?)" />`),
		"ukaz":                regexp.MustCompile(`<input style="display:none;" type="hidden" name="ukaz" value="(.*?)" />`),
		"bShowForm":           regexp.MustCompile(`<input style="display:none;" type="hidden" name="bShowForm" id=bShowForm value="(.*?)" />`),
		"lEdit":               regexp.MustCompile(`<input style="display:none;" type="hidden" name="lEdit" value="(.*?)" />`),
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	matches := make(map[string]string)
	for k, r := range regex {
		m := r.FindSubmatch(body)
		if m == nil {
			return nil, errors.New("failed to get all meta data")
		}

		matches[k] = string(m[1])
	}

	return matches, nil
}

func (c *Client) publishAd(ad *Ad, metaInfo map[string]string) error {
	buff := &bytes.Buffer{}
	w := multipart.NewWriter(buff)
	defer w.Close()

	// write meta info
	for k, v := range metaInfo {
		err := w.WriteField(k, v)
		if err != nil {
			return err
		}
	}

	// write visible ad fields
	params := map[string]string{
		"cNaziv":     ad.Title,
		"cOpis":      ad.Description,
		"nCenaStart": ad.Price,
		"nKatID":     ad.CategoryId,
		"cTip":       "O",
	}
	for k, v := range params {
		if err := w.WriteField(k, v); err != nil {
			return err
		}
	}

	// write images
	for i, image := range ad.Images {
		f, err := os.Open(image)
		if err != nil {
			return err
		}

		part, err := w.CreateFormFile(fmt.Sprintf("aSlikeUpload[%d]", i), image)
		if err != nil {
			return err
		}

		if _, err = io.Copy(part, f); err != nil {
			return err
		}

		if err := f.Close(); err != nil {
			return err
		}
	}

	// last image needs to be empty
	if _, err := w.CreateFormFile(fmt.Sprintf("aSlikeUpload[%d]", len(ad.Images)), ""); err != nil {
		return err
	}

	req, err := c.newRequest(http.MethodPost, "http://objava-oglasa.bolha.com/oddaj.php", buff)
	if err != nil {
		return err
	}

	addHeaders(req, map[string]string{
		"Host":                      "objava-oglasa.bolha.com",
		"Origin":                    "http://objava-oglasa.bolha.com",
		"Referer":                   fmt.Sprintf("http://objava-oglasa.bolha.com/oddaj.php?katid=%d&days=30", ad.CategoryId),
		"Upgrade-Insecure-Requests": "1",
	})

	req.Header.Set("Content-Type", w.FormDataContentType())

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}

func addHeaders(req *http.Request, headers map[string]string) {
	for k, v := range headers {
		req.Header.Set(k, v)
	}
}

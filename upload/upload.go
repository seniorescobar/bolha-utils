package upload

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
	"strconv"
	"time"
)

const (
	// URLs
	loginURL            string = "https://login.bolha.com//novaprijava.php"
	loginFailedRedirect string = "https://login.bolha.com/"
	removeURL           string = "https://moja.bolha.com/adManager/ajaxRemoveActiveBulk"
	categoryURL         string = "http://objava-oglasa.bolha.com/izbor_paketa.php"
	publishURL          string = "http://objava-oglasa.bolha.com/oddaj.php"
)

// Errors returned by Upload package
var (
	ErrLoginFailed    = errors.New("login failed")
	ErrAdNotRemoved   = errors.New("ad not removed")
	ErrNoMatches      = errors.New("no matches")
	ErrAdNotPublished = errors.New("ad not published")
)

// User represents a user of bolha.com
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Ad represents an ad from bolha.com
type Ad struct {
	User        *User    `json:"user"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Price       int      `json:"price"`
	CategoryId  int      `json:"categoryId"`
	Images      []string `json:"images"`
}

// UploadAd uploads an ad to bolha.com
func UploadAd(ad *Ad) error {
	// initialize a new client
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	client := &http.Client{
		Timeout: time.Duration(3) * time.Minute,
		Jar:     cookieJar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// log in
	if err := login(client, ad.User); err != nil {
		return err
	}

	// get ad meta info
	metaInfo, err := getAdMetaInfo(client, ad)
	if err != nil {
		return err
	}

	// publish ad
	_, err = publishAd(client, ad, metaInfo)
	if err != nil {
		return err
	}

	return nil
}

func login(client *http.Client, user *User) error {
	values := url.Values{
		"name": {
			user.Username,
		},
		"geslo": {
			user.Password,
		},
		"zapomni_me": {
			"zapomni",
		},
	}

	res, err := client.PostForm(loginURL, values)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// check if successful
	if res.StatusCode != http.StatusFound {
		return ErrLoginFailed
	}
	location, err := res.Location()
	if err != nil {
		return err
	}
	if location.String() == loginFailedRedirect {
		return ErrLoginFailed
	}

	return nil
}

func getAdMetaInfo(client *http.Client, ad *Ad) (map[string]string, error) {
	values := url.Values{
		"categoryId": {
			strconv.Itoa(ad.CategoryId),
		},
	}

	if _, err := client.PostForm(categoryURL, values); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, publishURL, nil)
	if err != nil {
		return nil, err
	}

	values = url.Values{
		"katid": {
			strconv.Itoa(ad.CategoryId),
		},
		"days": {
			"30",
		},
	}
	req.URL.RawQuery = values.Encode()

	res, err := client.Do(req)
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
			return nil, ErrNoMatches
		}

		matches[k] = string(m[1])
	}

	return matches, nil
}

func publishAd(client *http.Client, ad *Ad, metaInfo map[string]string) (int, error) {
	buff := &bytes.Buffer{}
	w := multipart.NewWriter(buff)
	defer w.Close()

	// write meta info
	for k, v := range metaInfo {
		err := w.WriteField(k, v)
		if err != nil {
			return 0, err
		}
	}

	// write visible ad fields
	params := map[string]string{
		"cNaziv":     ad.Title,
		"cOpis":      ad.Description,
		"nCenaStart": strconv.Itoa(ad.Price),
		"nKatID":     strconv.Itoa(ad.CategoryId),
		"cTip":       "O",
	}
	for k, v := range params {
		if err := w.WriteField(k, v); err != nil {
			return 0, err
		}
	}

	// write images
	for i, image := range ad.Images {
		f, err := os.Open(image)
		if err != nil {
			return 0, err
		}

		part, err := w.CreateFormFile(fmt.Sprintf("aSlikeUpload[%d]", i), image)
		if err != nil {
			return 0, err
		}

		if _, err = io.Copy(part, f); err != nil {
			return 0, err
		}

		if err := f.Close(); err != nil {
			return 0, err
		}
	}

	// last image needs to be empty
	if _, err := w.CreateFormFile(fmt.Sprintf("aSlikeUpload[%d]", len(ad.Images)), ""); err != nil {
		return 0, err
	}

	req, err := http.NewRequest(http.MethodPost, publishURL, buff)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	// get ad id
	url, err := res.Location()
	if err != nil {
		return 0, err
	}

	adIDStr := url.Query().Get("id")
	if adIDStr == "" {
		return 0, ErrAdNotPublished
	}

	adID, err := strconv.Atoi(adIDStr)
	if err != nil {
		return 0, err
	}

	return adID, nil
}

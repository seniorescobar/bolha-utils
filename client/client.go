package client

import (
	"net/http"
)

// Ad represents an ad from bolha.com
type Ad struct {
	Title       string
	Description string
	Price       string
	CategoryId  string
	Images      []string
}

// CLIENT
// Client represents a bolha client
type Client struct {
	username   string
	httpClient *http.Client
}

// New creates a new bolha client
func New(username, password string) (*Client, error) {
	httpClient, err := getHttpClient()
	if err != nil {
		return nil, err
	}

	client := &Client{
		username:   username,
		httpClient: httpClient,
	}

	if err := client.logIn(username, password); err != nil {
		return nil, err
	}

	return client, nil
}

// UPLOAD
// UploadAd uploads a single ad
func (c *Client) UploadAd(ad *Ad) error {
	return c.UploadAds([]*Ad{ad})
}

// UploadAds uploads multiple ads
func (c *Client) UploadAds(ads []*Ad) error {
	for _, ad := range ads {
		if err := c.uploadAd(ad); err != nil {
			return err
		}
	}
	return nil
}

// REMOVE
// RemoveAd removes a single ad provided by an id
func (c *Client) RemoveAd(id string) error {
	return c.removeAds([]string{id})
}

// RemoveAds removes multiple ads provided by ids
func (c *Client) RemoveAds(ids []string) error {
	return c.removeAds(ids)
}

// RemoveAllAds removes all ads found on a user's account
func (c *Client) RemoveAllAds() error {
	ids, err := c.getAdIds()
	if err != nil {
		return err
	}

	return c.removeAds(ids)
}

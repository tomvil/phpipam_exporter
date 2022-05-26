package apiclient

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Client struct {
	apiAddress  string
	apiUsername string
	apiPassword string
}

func NewClient(apiAddress, apiUsername, apiPassword string) *Client {
	return &Client{apiAddress, apiUsername, apiPassword}
}

func (c *Client) Get(path string) ([]byte, error) {
	var client http.Client

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.apiUsername, c.apiPassword)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c *Client) GetParsed(path string, obj interface{}) error {
	body, err := c.Get(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, obj)
	return err
}

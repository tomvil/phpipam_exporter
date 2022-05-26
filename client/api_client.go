package apiclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	apiAddress  string
	apiUsername string
	apiPassword string
	apiToken    string
}

type Response struct {
	Data struct {
		Token string
	}
}

func NewClient(apiAddress, apiUsername, apiPassword string) (*Client, error) {
	apiToken, err := getApiToken(apiAddress, apiUsername, apiPassword)
	if err != nil {
		return nil, err
	}

	return &Client{apiAddress, apiUsername, apiPassword, apiToken}, nil
}

func getApiToken(apiAddress, apiUsername, apiPassword string) (string, error) {
	var client http.Client
	var response Response

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/user/", apiAddress), nil)
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(apiUsername, apiPassword)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	return response.Data.Token, nil
}

func (c *Client) Get(path string) ([]byte, error) {
	var client http.Client

	req, err := http.NewRequest("GET", c.apiAddress+path, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("token", c.apiToken)
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

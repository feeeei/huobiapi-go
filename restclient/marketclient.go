package restclient

import (
	"net/url"

	"github.com/bitly/go-simplejson"
	"github.com/feeeei/huobiapi-go/config"
)

type MarketClient struct {
	Endpoint *url.URL
}

// NewMarketClient REST格式行情Client
func NewMarketClient() (*MarketClient, error) {
	return &MarketClient{
		Endpoint: config.HuobiRestEndpoint,
	}, nil
}

// Get Get同步请求
func (client *MarketClient) Get(path string, params ...map[string]interface{}) (*simplejson.Json, error) {
	if err := isValidParams(params); err != nil {
		return nil, err
	}

	var p map[string]interface{}
	if params != nil {
		p = params[0]
	}
	url := client.Endpoint.String() + path
	return request("GET", url, p)
}

// Post Post同步请求
func (client *MarketClient) Post(path string, params ...map[string]interface{}) (*simplejson.Json, error) {
	if err := isValidParams(params); err != nil {
		return nil, err
	}

	url := client.Endpoint.String() + path
	var body map[string]interface{}
	if params != nil {
		body = params[0]
	}
	return request("POST", url, body)
}

package restclient

import (
	"net/url"

	"github.com/feeeei/huobiapi-go/utils"

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

// HandleGet 将Response解析到obj中
func (client *MarketClient) HandleGet(path string, obj interface{}, params ...map[string]interface{}) (*simplejson.Json, error) {
	if err := utils.CheckPointer(obj); err != nil {
		return nil, err
	}
	resp, err := client.Get(path, params...)
	if err != nil {
		return resp, err
	}
	return utils.Parse2Obj(resp, obj)
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

// HandlePost 将Response解析到obj中
func (client *MarketClient) HandlePost(path string, obj interface{}, params ...map[string]interface{}) (*simplejson.Json, error) {
	if err := utils.CheckPointer(obj); err != nil {
		return nil, err
	}
	resp, err := client.Post(path, params...)
	if err != nil {
		return resp, err
	}
	return utils.Parse2Obj(resp, obj)
}

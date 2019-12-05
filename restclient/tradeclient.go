package restclient

import (
	"net/url"

	"github.com/bitly/go-simplejson"
	"github.com/feeeei/huobiapi-go/config"
	"github.com/feeeei/huobiapi-go/sign"
	"github.com/feeeei/huobiapi-go/utils"
)

type TradeClient struct {
	Endpoint *url.URL
	sign     *sign.Sign
}

// NewTradeClient REST格式交易Client
func NewTradeClient(accessKeyID, accessKeySecret string) (*TradeClient, error) {
	return &TradeClient{
		Endpoint: config.HuobiRestEndpoint,
		sign:     sign.NewSign(accessKeyID, accessKeySecret),
	}, nil
}

// Get Get同步请求
func (client *TradeClient) Get(path string, params ...map[string]interface{}) (*simplejson.Json, error) {
	if err := isValidParams(params); err != nil {
		return nil, err
	}

	p := client.sign.GetSignFields()
	if params != nil {
		p = utils.MergeMap(p, params[0])
	}
	p = client.signParams("GET", path, p)
	url := client.Endpoint.String() + path
	return request("GET", url, p)
}

func (client *TradeClient) Post(path string, params ...map[string]interface{}) (*simplejson.Json, error) {
	if err := isValidParams(params); err != nil {
		return nil, err
	}
	p := client.signParams("POST", path, client.sign.GetSignFields())
	url := client.Endpoint.String() + path + "?" + utils.EncodeQueryString(p)
	if params != nil {
		p = params[0]
	} else {
		p = nil
	}
	return request("POST", url, p)
}

func (client *TradeClient) signParams(method, path string, params map[string]interface{}) map[string]interface{} {
	params["Signature"] = utils.Sign(method, client.Endpoint.Host, path, client.sign.AccessKeySecret, params)
	return params
}

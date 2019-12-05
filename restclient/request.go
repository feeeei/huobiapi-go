package restclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/feeeei/huobiapi-go/utils"

	"github.com/bitly/go-simplejson"
)

func request(method, url string, params map[string]interface{}) (*simplejson.Json, error) {
	url, body := parameters(method, url, params)
	var req *http.Request
	var err error
	if body != nil {
		req, err = http.NewRequest(method, url, body)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		return nil, err
	}
	addHeaders(method, req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	json, err := simplejson.NewJson(respBody)
	var status = json.Get("status").MustString()
	if status == "error" {
		return json, fmt.Errorf(json.Get("err-msg").MustString())
	}
	return json, nil
}

func addHeaders(method string, req *http.Request) *http.Request {
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.71 Safari/537.36")
	if isGetMethod(method) {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req.Header.Add("Content-Type", "application/json")
	}
	return req
}

func parameters(method, urlStr string, params map[string]interface{}) (string, *bytes.Buffer) {
	if !isGetMethod(method) {
		b, _ := json.Marshal(params)
		return urlStr, bytes.NewBuffer(b)
	}
	return urlStr + "?" + utils.EncodeQueryString(params), nil
}

func isValidParams(params ...interface{}) error {
	if len(params) > 1 {
		return fmt.Errorf("Request parameter error")
	}
	return nil
}

func isGetMethod(method string) bool {
	if method == "GET" || method == "HEAD" {
		return true
	}
	return false
}

package wsclient

import (
	"fmt"

	"github.com/bitly/go-simplejson"
	"github.com/feeeei/huobiapi-go/config"
	"github.com/feeeei/huobiapi-go/debug"
	"github.com/feeeei/huobiapi-go/sign"
	"github.com/feeeei/huobiapi-go/utils"
)

type TradeWSClient struct {
	ws            *safeWebSocket
	subscribeWait map[string]chan error
	responseWait  map[string]chan *simplejson.Json
	sign          *sign.Sign
}

// NewTradeWSClient WebSocket格式交易Client
func NewTradeWSClient(accessKeyID, accessKeySecret string) (*TradeWSClient, error) {
	client := &TradeWSClient{
		subscribeWait: make(map[string]chan error),
		responseWait:  make(map[string]chan *simplejson.Json),
		sign:          sign.NewSign(accessKeyID, accessKeySecret),
	}
	ws, err := newSafeWebSocket(config.HuobiWsTradeEndpoint, client)
	if err != nil {
		return nil, err
	}
	client.ws = ws
	if err := client.auth(); err != nil {
		ws.close()
		return nil, err
	}
	debug.Println("Trade websocket auth sccessful")
	return client, nil
}

// auth 鉴权
func (client *TradeWSClient) auth() error {
	message := client.authParams()
	client.ws.sendMessage(message)
	client.subscribeWait["auth"] = make(chan error)
	return <-client.subscribeWait["auth"]
}

// Request 一次性类请求，阻塞式返回结果
func (client *TradeWSClient) Request(topic string, fields ...map[string]interface{}) (*simplejson.Json, error) {
	field := getRequestFields(topic, fields...)
	field["op"] = "req"
	client.ws.sendMessage(field)
	client.responseWait[topic] = make(chan *simplejson.Json)
	json := <-client.responseWait[topic]
	return json, checkResponseError(json)
}

// Subscribe 订阅主题
func (client *TradeWSClient) Subscribe(topic string, listener Subscriber, fields ...map[string]interface{}) error {
	field := getRequestFields(topic, fields...)
	field["op"] = "sub"
	client.ws.sendMessage(field)
	client.subscribeWait[topic] = make(chan error)
	if err := <-client.subscribeWait[topic]; err != nil {
		return err
	}
	client.ws.subscribers[topic] = listener
	return nil
}

// UnSubscribe 取消订阅主题
func (client *TradeWSClient) UnSubscribe(topic string) {
	client.ws.sendMessage(map[string]interface{}{"op": "unsub", "topic": topic})
	client.ws.unsubscribe(topic)
}

// handle 处理消息
func (client *TradeWSClient) handle(json *simplejson.Json) {
	op := json.Get("op").MustString()
	topic := json.Get("topic").MustString()

	switch op {
	case "auth":
		client.handleError("auth", json)
	case "sub":
		client.handleError(topic, json)
	case "unsub":
		debug.Println("Unsub", topic)
	case "req":
		client.handleResponse(topic, json)
	case "notify":
		subscriber := client.ws.subscribers[topic]
		if subscriber != nil {
			subscriber(topic, json)
		}
	}
}

func (client *TradeWSClient) handleError(topic string, json *simplejson.Json) {
	client.subscribeWait[topic] <- checkResponseError(json)
}

func (client *TradeWSClient) handleResponse(topic string, json *simplejson.Json) {
	client.responseWait[topic] <- json
}

func (client *TradeWSClient) authParams() map[string]interface{} {
	params := client.sign.GetSignFields()
	params["Signature"] = utils.Sign("GET",
		client.ws.url.Host,
		client.ws.url.Path,
		client.sign.AccessKeySecret,
		params)
	params["op"] = "auth"
	return params
}

func checkResponseError(json *simplejson.Json) error {
	if json.Get("err-code").MustInt() == 0 {
		return nil
	} else {
		return fmt.Errorf(json.Get("err-msg").MustString())
	}
}

func getRequestFields(topic string, fields ...map[string]interface{}) map[string]interface{} {
	var m map[string]interface{}
	if fields == nil {
		m = make(map[string]interface{})
	}
	m["topic"] = topic
	return m
}

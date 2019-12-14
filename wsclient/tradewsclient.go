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
	autoReconnect bool
}

// NewTradeWSClient WebSocket格式交易Client
func NewTradeWSClient(accessKeyID, accessKeySecret string) (*TradeWSClient, error) {
	client := &TradeWSClient{
		subscribeWait: make(map[string]chan error),
		responseWait:  make(map[string]chan *simplejson.Json),
		sign:          sign.NewSign(accessKeyID, accessKeySecret, "2"),
		autoReconnect: true,
	}
	if err := client.connect(); err != nil {
		return nil, err
	}
	return client, nil
}

func (client *TradeWSClient) connect() error {
	ws, err := newSafeWebSocket(config.HuobiWsTradeEndpoint, client, client.autoReconnect, true)
	if err != nil {
		return err
	}
	client.ws = ws
	if err := client.auth(); err != nil {
		ws.close()
		return err
	}
	debug.Println("Trade websocket auth sccessful")
	return nil
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
	var field map[string]interface{}
	if fields == nil {
		field = make(map[string]interface{})
	} else {
		field = fields[0]
	}
	field["topic"] = topic
	field["op"] = "req"
	client.ws.sendMessage(field)
	client.responseWait[topic] = make(chan *simplejson.Json)
	json := <-client.responseWait[topic]
	return json, client.checkResponseError(json)
}

// HandleRequest 将Response解析到obj中
func (client *TradeWSClient) HandleRequest(topic string, obj interface{}, fields ...map[string]interface{}) (*simplejson.Json, error) {
	if err := utils.CheckPointer(obj); err != nil {
		return nil, err
	}
	resp, err := client.Request(topic, fields...)
	if err != nil {
		return resp, err
	}
	return utils.Parse2Obj(resp, obj)
}

// Subscribe 订阅主题
func (client *TradeWSClient) Subscribe(topic string, listener Subscriber) error {
	// 如果已经订阅，直接刷新 listener
	if client.ws.subscribers[topic] != nil {
		client.ws.subscribe(topic, listener)
		return nil
	}

	field := map[string]interface{}{"topic": topic, "op": "sub"}
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

// SetAutoReconnect 设置socket中断时自动重新链接，默认true
func (client *TradeWSClient) SetAutoReconnect(autoReconnect bool) {
	client.autoReconnect = autoReconnect
	client.ws.autoReconnect = autoReconnect
}

// Reconnect 重新连接，关闭旧链接，建立新链接，重新授权，重新订阅
func (client *TradeWSClient) Reconnect() {
	client.ws.reconnect()
}

// handle 处理消息
func (client *TradeWSClient) handle(json *simplejson.Json) {
	op := json.Get("op").MustString()
	topic := json.Get("topic").MustString()

	switch op {
	case "pong":
		// huobi WebSocket v1 接口没有客户端主动发起ping方式
	case "ping":
		json.Set("op", "pong")
		client.ws.sendMessage(json)
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
	client.subscribeWait[topic] <- client.checkResponseError(json)
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

func (client *TradeWSClient) checkResponseError(json *simplejson.Json) error {
	if json.Get("err-code").MustInt() == 0 {
		return nil
	}
	return fmt.Errorf(json.Get("err-msg").MustString())
}

func getRequestFields(topic string) map[string]interface{} {
	m := make(map[string]interface{})
	m["topic"] = topic
	return m
}

package wsclient

import (
	"fmt"

	"github.com/bitly/go-simplejson"
	"github.com/feeeei/huobiapi-go/config"
	"github.com/feeeei/huobiapi-go/debug"
	"github.com/feeeei/huobiapi-go/sign"
	"github.com/feeeei/huobiapi-go/utils"
)

type TradeWSV2Client struct {
	ws            *huobiWebSocket
	subscribeWait map[string]chan error
	responseWait  map[string]chan *simplejson.Json
	sign          *sign.Sign
	autoReconnect bool
}

// NewTradeWSV2Client WebSocket格式交易Client
func NewTradeWSV2Client(accessKeyID, accessKeySecret string) (*TradeWSV2Client, error) {
	client := &TradeWSV2Client{
		subscribeWait: make(map[string]chan error),
		responseWait:  make(map[string]chan *simplejson.Json),
		sign:          sign.NewSign(accessKeyID, accessKeySecret, "2.1"),
		autoReconnect: true,
	}
	if err := client.connect(); err != nil {
		return nil, err
	}
	return client, nil
}

func (client *TradeWSV2Client) connect() error {
	ws, err := newHuobiWebSocket(config.HuobiWsTradeV2Endpoint, client, client.autoReconnect, false)
	if err != nil {
		return err
	}
	client.ws = ws
	if err := client.auth(); err != nil {
		ws.close()
		return err
	}
	debug.Println("TradeV2 websocket auth sccessful")
	return nil
}

func (client *TradeWSV2Client) auth() error {
	message := client.authParams()
	authMessage := map[string]interface{}{
		"action": "req",
		"ch":     "auth",
		"params": message,
	}
	client.ws.sendMessage(authMessage)
	client.subscribeWait["auth"] = make(chan error)
	return <-client.subscribeWait["auth"]
}

// Subscribe 订阅主题
func (client *TradeWSV2Client) Subscribe(topic string, listener Subscriber) error {
	// 如果已经订阅，直接刷新 listener
	if client.ws.subscribers[topic] != nil {
		client.ws.subscribe(topic, listener)
		return nil
	}

	field := map[string]interface{}{"action": "sub", "ch": topic}
	client.ws.sendMessage(field)
	client.subscribeWait[topic] = make(chan error)
	if err := <-client.subscribeWait[topic]; err != nil {
		return err
	}
	client.ws.subscribers[topic] = listener
	return nil
}

// UnSubscribe 取消订阅主题
func (client *TradeWSV2Client) UnSubscribe(topic string) {
	// TODO 火币暂未实现该接口，先本地取消订阅
	client.ws.unsubscribe(topic)
}

// SetAutoReconnect 设置socket中断时自动重新链接，默认true
func (client *TradeWSV2Client) SetAutoReconnect(autoReconnect bool) {
	client.autoReconnect = autoReconnect
	client.ws.autoReconnect = autoReconnect
}

// Reconnect 重新连接，关闭旧链接，建立新链接，重新授权，重新订阅
func (client *TradeWSV2Client) Reconnect() {
	client.ws.reconnect()
}

// handle 处理消息
func (client *TradeWSV2Client) handle(json *simplejson.Json) {
	action := json.Get("action").MustString()
	ch := json.Get("ch").MustString()

	switch action {
	case "pong":
		// huobi WebSocket v2 接口没有客户端主动发起ping方式
	case "ping":
		json.Set("action", "pong")
		client.ws.sendMessage(json)
	case "req":
		if ch == "auth" {
			client.handleError(ch, json)
		} else {
			client.handleResponse(ch, json)
		}
	case "sub":
		client.handleError(ch, json)
	case "push":
		subscriber := client.ws.subscribers[ch]
		if subscriber != nil {
			subscriber(ch, json)
		}
	}
}

func (client *TradeWSV2Client) handleResponse(topic string, json *simplejson.Json) {
	client.responseWait[topic] <- json
}

func (client *TradeWSV2Client) handleError(topic string, json *simplejson.Json) {
	client.subscribeWait[topic] <- client.checkResponseError(json)
}

func (client *TradeWSV2Client) authParams() map[string]interface{} {
	params := client.sign.GetSignFields()
	params["signature"] = utils.Sign("GET",
		client.ws.url.Host,
		client.ws.url.Path,
		client.sign.AccessKeySecret,
		params)
	params["authType"] = "api"
	return params
}

func (client *TradeWSV2Client) checkResponseError(json *simplejson.Json) error {
	if json.Get("code").MustInt() == 200 {
		return nil
	}
	return fmt.Errorf(json.Get("message").MustString())
}

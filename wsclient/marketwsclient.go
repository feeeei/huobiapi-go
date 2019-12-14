package wsclient

import (
	"fmt"

	"github.com/bitly/go-simplejson"
	"github.com/feeeei/huobiapi-go/config"
	"github.com/feeeei/huobiapi-go/debug"
	"github.com/feeeei/huobiapi-go/utils"
)

type MarketWSClient struct {
	ws            *safeWebSocket
	subscribeWait map[string]chan error
	autoReconnect bool
}

// NewMarketWSClient WebSocket格式行情Client
func NewMarketWSClient() (*MarketWSClient, error) {
	client := &MarketWSClient{
		subscribeWait: make(map[string]chan error),
		autoReconnect: true,
	}
	if err := client.connect(); err != nil {
		return nil, err
	}
	return client, nil
}

func (client *MarketWSClient) connect() error {
	ws, err := newSafeWebSocket(config.HuobiWsEndpoint, client, client.autoReconnect, true)
	if err != nil {
		return err
	}
	client.ws = ws
	client.keepAlive()
	return nil
}

// Subscribe 订阅主题
func (client *MarketWSClient) Subscribe(topic string, listener Subscriber) error {
	// 如果已经订阅，直接刷新 listener
	if client.ws.subscribers[topic] != nil {
		client.ws.subscribe(topic, listener)
		return nil
	}

	client.ws.sendMessage(map[string]interface{}{"sub": topic, "id": topic})
	client.subscribeWait[topic] = make(chan error)
	err := <-client.subscribeWait[topic]
	if err != nil {
		return err
	}
	client.ws.subscribe(topic, listener) // 如果订阅成功，再加入监听列表
	return nil
}

// UnSubscribe 取消订阅主题
func (client *MarketWSClient) UnSubscribe(topic string) {
	if client.ws.subscribers[topic] == nil {
		return
	}

	client.ws.sendMessage(map[string]interface{}{"unsub": topic})
	client.ws.unsubscribe(topic)
	return
}

// SetAutoReconnect 设置socket中断时自动重新链接，默认true
func (client *MarketWSClient) SetAutoReconnect(autoReconnect bool) {
	client.autoReconnect = autoReconnect
	client.ws.autoReconnect = autoReconnect
}

// Reconnect 重新连接，关闭旧链接，建立新链接，重新授权，重新订阅
func (client *MarketWSClient) Reconnect() {
	client.ws.reconnect()
}

func (client *MarketWSClient) keepAlive() {
	client.ws.keepAlive(config.HeartbeatDuration, client)
}

// handle 处理消息
func (client *MarketWSClient) handle(json *simplejson.Json) {
	// 处理订阅推送消息
	if topic, isExist := json.CheckGet("ch"); isExist {
		topicStr := topic.MustString()
		subscriber := client.ws.subscribers[topicStr]
		if subscriber != nil {
			subscriber(topicStr, json)
		}
		return
	}

	// 处理 ping
	if ping, isExist := json.CheckGet("ping"); isExist {
		client.ws.sendMessage(map[string]interface{}{"pong": ping.MustInt64()})
		return
	}

	// 处理 pong
	if _, isExist := json.CheckGet("pong"); isExist {
		return
	}

	// 处理订阅成功消息
	if topic, isExist := json.CheckGet("subbed"); isExist {
		client.subscribeWait[topic.MustString()] <- nil
		return
	}

	// 处理取消订阅消息
	if topic, isExist := json.CheckGet("unsubbed"); isExist {
		debug.Println("Unsubscribe", topic.MustString(), json.Get("status").MustString())
		return
	}

	// 处理订阅失败消息
	if json.Get("status").MustString() == "error" {
		if id, isExist := json.CheckGet("id"); isExist {
			err := fmt.Errorf(json.Get("err-msg").MustString())
			client.subscribeWait[id.MustString()] <- err
		}
		return
	}
}

func (client *MarketWSClient) ping() map[string]interface{} {
	return map[string]interface{}{"ping": utils.UinxMillisecond()}
}

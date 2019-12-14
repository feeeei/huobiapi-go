package wsclient

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/feeeei/huobiapi-go/debug"
	"github.com/feeeei/huobiapi-go/utils"
	"github.com/gorilla/websocket"
)

type Subscriber func(topic string, json *simplejson.Json)
type aliver interface {
	ping() map[string]interface{}
}
type handler interface {
	connect() error
	handle(json *simplejson.Json)
	Subscribe(topic string, listener Subscriber) error
	UnSubscribe(topic string)
	Reconnect()
	SetAutoReconnect(autoReconnect bool)
}

type safeWebSocket struct {
	url           *url.URL
	ws            *websocket.Conn
	subscribers   map[string]Subscriber
	handler       handler
	alive         bool
	autoReconnect bool
	needDecrypt   bool
	m             sync.RWMutex
}

func newSafeWebSocket(u *url.URL, handler handler, autoReconnect, needDecrypt bool) (*safeWebSocket, error) {
	client := safeWebSocket{
		url:           u,
		subscribers:   make(map[string]Subscriber),
		handler:       handler,
		autoReconnect: autoReconnect,
		needDecrypt:   needDecrypt,
	}
	if err := client.newConnect(); err != nil {
		return nil, err
	}
	return &client, nil
}

func (client *safeWebSocket) newConnect() error {
	client.m.Lock()
	defer client.m.Unlock()
	ws, response, err := websocket.DefaultDialer.Dial(client.url.String(), nil)
	if response == nil || response.StatusCode >= 400 {
		return fmt.Errorf("Connection not established")
	}
	if err != nil {
		return err
	}
	client.ws = ws
	client.alive = true
	go client.handleMessageLoop()
	debug.Println("WebSocket connected")
	return nil
}

func (client *safeWebSocket) handleMessageLoop() {
	for true {
		_, rawMessage, err := client.ws.ReadMessage()
		if err != nil {
			debug.Println("handle message loop error: ", client.autoReconnect, err)
			break
		}
		var message []byte
		if client.needDecrypt {
			message, err = utils.DecodeGzip(rawMessage)
		} else {
			message = rawMessage
		}
		if err != nil {
			debug.Println("handle message loop error: ", client.autoReconnect, err)
			break
		}
		debug.Println("Receive:", string(message))
		json, _ := simplejson.NewJson(message)
		client.handler.handle(json)
	}
	if client.autoReconnect {
		client.reconnect()
	}
}

func (client *safeWebSocket) keepAlive(duration time.Duration, heartbeat aliver) {
	go func() {
		for client.alive {
			err := client.sendMessage(heartbeat.ping())
			client.alive = err == nil
			if err != nil {
				client.reconnect()
			} else {
				time.Sleep(duration)
			}
		}
	}()
}

// subscribe 注册订阅
func (client *safeWebSocket) subscribe(topic string, listener Subscriber) {
	client.m.Lock()
	defer client.m.Unlock()
	client.subscribers[topic] = listener
}

// unsubscribe 取消订阅
func (client *safeWebSocket) unsubscribe(topic string) {
	client.m.Lock()
	defer client.m.Unlock()
	delete(client.subscribers, topic)
}

// sendMessage 通过Websocket发送request
func (client *safeWebSocket) sendMessage(message interface{}) error {
	b, err := json.Marshal(message)
	if err != nil {
		return nil
	}
	debug.Println("Send message:", string(b))
	return client.send(b)
}

func (client *safeWebSocket) send(b []byte) error {
	client.m.Lock()
	defer client.m.Unlock()
	err := client.ws.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		debug.Println("Send message error:", err)
	}
	return err
}

// reconnect 循环式重新链接，如果中途失败会sleep 1s之后继续尝试
func (client *safeWebSocket) reconnect() {
	success := false
	for !success {
		debug.Println("Begin reconnecting")
		time.Sleep(time.Second * 1)
		client.close()
		if err := client.handler.connect(); err != nil {
			debug.Println("Reconneting error:", err)
			continue
		}
		for topic := range client.subscribers {
			if err := client.handler.Subscribe(topic, client.subscribers[topic]); err != nil {
				debug.Println("Reconneting subscribe error:", err)
				continue
			}
		}
		success = true
	}
	debug.Println("Reconnecting successful")
	return
}

func (client *safeWebSocket) close() {
	client.m.Lock()
	defer client.m.Unlock()
	client.alive = false
	client.ws.Close()
}

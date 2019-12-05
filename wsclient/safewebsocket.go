package wsclient

import (
	"encoding/json"
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
	handle(json *simplejson.Json)
}

type safeWebSocket struct {
	url         *url.URL
	ws          *websocket.Conn
	subscribers map[string]Subscriber
	handler     handler
	alive       bool
	m           sync.RWMutex
}

func newSafeWebSocket(u *url.URL, handler handler) (*safeWebSocket, error) {
	ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	client := safeWebSocket{
		url:         u,
		ws:          ws,
		subscribers: make(map[string]Subscriber),
		handler:     handler,
	}
	go client.handleMessageLoop()
	return &client, nil
}

func (client *safeWebSocket) handleMessageLoop() {
	for {
		_, rawMessage, err := client.ws.ReadMessage()
		if err != nil {
			debug.Println("handle message loop error: ", err)
			return
		}
		message, err := utils.DecodeGzip(rawMessage)
		if err != nil {
			debug.Println("handle message loop error: ", err)
			return
		}
		debug.Println("Receive:", string(message))
		json, _ := simplejson.NewJson(message)
		client.handler.handle(json)
	}
}

func (client *safeWebSocket) keepAlive(duration time.Duration, heartbeat aliver) {
	go func() {
		for {
			b, _ := json.Marshal(heartbeat.ping())
			err := client.send(b)
			if err != nil {
				client.alive = false
				if _, ok := err.(*websocket.CloseError); ok {
					break
				}
			}
			client.alive = true
			time.Sleep(duration)
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
	if err := client.ws.WriteMessage(websocket.TextMessage, b); err != nil {
		debug.Println("Send message error:", err)
	}
	return nil
}

func (client *safeWebSocket) reconnect() error {
	//TODO
	return nil
}

func (client *safeWebSocket) close() {
	client.m.Lock()
	defer client.m.Unlock()
	if client.ws != nil {
		client.ws.Close()
	}
}

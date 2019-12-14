package config

import (
	"net/url"
	"time"
)

var HuobiApiHost string
var HuobiRestEndpoint *url.URL
var HuobiWsEndpoint *url.URL
var HuobiWsTradeEndpoint *url.URL
var HuobiWsTradeV2Endpoint *url.URL

var HeartbeatDuration = time.Second * 5

func SetAPIHost(host string) {
	HuobiApiHost = host
	HuobiRestEndpoint, _ = url.Parse("https://" + host)
	HuobiWsEndpoint, _ = url.Parse("wss://" + host + "/ws")
	HuobiWsTradeEndpoint, _ = url.Parse("wss://" + host + "/ws/v1")
	HuobiWsTradeV2Endpoint, _ = url.Parse("wss://" + host + "/ws/v2")
}

package huobiapi

import (
	"github.com/feeeei/huobiapi-go/config"
	"github.com/feeeei/huobiapi-go/debug"
	"github.com/feeeei/huobiapi-go/restclient"
	"github.com/feeeei/huobiapi-go/wsclient"
)

type Params = map[string]interface{}

// MarketClient REST格式市场client
type MarketClient = restclient.MarketClient

// TradeClient REST格式交易client
type TradeClient = restclient.TradeClient

// MarketWSClient WebSocket格式市场client
type MarketWSClient = wsclient.MarketWSClient

// TradeWSClient WebSocket格式交易client
type TradeWSClient = wsclient.TradeWSClient

//TradeWSV2Client WebSocket格式交易clientV2
type TradeWSV2Client = wsclient.TradeWSV2Client

func init() {
	config.SetAPIHost("api.huobi.pro")
}

// UseAWSHost 使用aws域名，在aws网络下速度时延更低
func UseAWSHost() {
	config.SetAPIHost("api-aws.huobi.pro")
}

// SetAPIHost 使用自定义Host
func SetAPIHost(host string) {
	config.SetAPIHost(host)
}

// DebugMode 是否使用Debug模式，打印日志
func DebugMode(output bool) {
	debug.Debug(output)
}

// NewMarketClient 创建REST行情Client
func NewMarketClient() (*MarketClient, error) {
	return restclient.NewMarketClient()
}

// NewTradeClient 创建REST交易Client
func NewTradeClient(accessKeyID, accessKeySecret string) (*TradeClient, error) {
	return restclient.NewTradeClient(accessKeyID, accessKeySecret)
}

// NewMarketWSClient 创建WebSocket行情Client
func NewMarketWSClient() (*MarketWSClient, error) {
	return wsclient.NewMarketWSClient()
}

// NewTradeWSClient 创建WebSocket交易Client
func NewTradeWSClient(accessKeyID, accessKeySecret string) (*TradeWSClient, error) {
	return wsclient.NewTradeWSClient(accessKeyID, accessKeySecret)
}

// NewTradeWSV2Client 创建WebSocket交易Client
func NewTradeWSV2Client(accessKeyID, accessKeySecret string) (*TradeWSV2Client, error) {
	return wsclient.NewTradeWSV2Client(accessKeyID, accessKeySecret)
}

## Huobi现货API for Go

官方 API 文档地址：[https://huobiapi.github.io/docs/spot/v1/cn/](https://huobiapi.github.io/docs/spot/v1/cn/)

## 目录
* [安装](#安装)
* [RESTful Client](#RESTful-Client)
* [WebSocket 行情Client](#WebSocket-行情Client)
* [WebSocket 资产&订单Client](#WebSocket-资产&订单Client)
* [WebSocket 资产&订单ClientV2](#WebSocket-资产&订单ClientV2)
* [其它配置](#其它配置)

## 进度
- [x] RESTful 行情、账户接口
- [x] WebSocket 行情、资产&订单 接口
- [x] 替换 Host 为 aws 域名
- [x] WebSocket 断线重连
- [x] WebSocket V2版本客户端

## 安装
```
go get -u https://github.com/feeeei/huobiapi-go
```

## 使用

## RESTful Client
```go
// 行情 Client
client, err := huobiapi.NewMarketClient()
// OR
// 行情 & 交易 Client
client, err := NewTradeClient("AccessKeyID", "AccessKeySecret")


// 获取 accounts 信息
json, _ := client.Get("/v1/account/accounts")
accounts, _ := json.Get("data").Array()
for _, account := range accounts {
    log.Println(account)
}


// 获取 btcusdt 实时价格
json, _ := client.Get("/market/trade", huobiapi.Params{"symbol": "btcusdt"})
log.Println("Current btcusdt price:", json.Get("data").Get("price").MustFloat64())


// 下单
json, _ := client.Post("/v1/order/orders/place", huobiapi.Params{
	"account-id":      "xxxxxx",
	"amount":          "1.00",
	"price":           "5000.00",
	"source":          "api",
	"symbol":          "btcusdt",
	"type":            "buy-limit",
	"client-order-id": "client-order-id",
})
// dosomething
```

## WebSocket 行情Client
```go
client, _ := huobiapi.NewMarketWSClient()


// 订阅 btcusdt 实时交易明细
err := client.Subscribe("market.btcusdt.trade.detail", func(topic string, json *simplejson.Json) {
    d,_ := json.Encode()
    log.Println(topic, string(d))
})
if err == nil {
    log.Println("Subscribe btcusdt successful")
}


// 取消订阅
client.UnSubscribe("market.btcusdt.trade.detail")
```

## WebSocket 资产&订单Client
```go
client, _ := huobiapi.NewTradeWSClient("AccessKeyID", "AccessKeySecret")


// 请求用户资产数据，阻塞式直接返回结果
json, err := client.Request("accounts.list")
if err == nil {
	d, _ := json.Encode()
	log.Println(string(d))
}


// 订阅账户更新
client.Subscribe("accounts", func(topic string, json *simplejson.Json) {
    d, _ := json.Encode()
	log.Println(string(d))
}, huobiapi.Params{"mode": 0}) // 是否包含已冻结余额


// 订阅btcusdt交易对下订单变更
client.Subscribe("orders.btcusdt.update", func(topic string, json *simplejson.Json) {
    d, _ := json.Encode()
	log.Println(string(d))
})
```

## WebSocket-资产&订单ClientV2
```go
client, _ := huobiapi.NewTradeWSClient("AccessKeyID", "AccessKeySecret")

// 订阅btcusdt订阅清算后成交明细
client.Subscribe("trade.clearing#usdthusd", func(topic string, json *simplejson.Json) {
    d, _ := json.Encode()
    log.Println(string(d))
})
// 订阅账户余额变动
client.Subscribe("accounts.update#0", func(topic string, json *simplejson.Json) {
    d, _ := json.Encode()
    log.Println(string(d))
})
```

## 其它配置
```
huobiapi.UseAWSHost()     // 使用aws域名，在aws网络环境下延迟更低
huobiapi.SetAPIHost("xx") // 使用自定义Host，可以使用未被墙Host来在境内使用
huobiapi.DebugMode(true)  // 是否使用Debug模式，打印日志
```

## 感谢
[leizongmin/huobiapi](https://github.com/leizongmin/huobiapi)

## License
feeeei/huobiapi-go is released under the [MIT License](https://opensource.org/licenses/MIT).
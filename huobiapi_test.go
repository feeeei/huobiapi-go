package huobiapi_test

import (
	"testing"

	"github.com/bitly/go-simplejson"
	jsoniter "github.com/json-iterator/go"
)

var text = []byte(`{"ch":"market.btcusdt.trade.detail","ts":1575513575039,"tick":{"id":103337083910,"ts":1575513574985,"data":[{"id":10333708391058731767773,"ts":1575513574985,"tradeId":102062646875,"amount":0.043776,"price":7218.58,"direction":"sell"}]}}`)

func BenchmarkJSON(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		json, _ := simplejson.NewJson(text)
		_ = json.Get("ch").MustString()
		_ = json.Get("tick").Get("id").MustInt64()
	}
}

func BenchmarkJsoniter(b *testing.B) {
	// b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = jsoniter.Get(text, "ch").ToString()
		_ = jsoniter.Get(text, "tick", "id").ToInt64()
		// json := jsoniter.ParseString(jsoniter.ConfigDefault, text).ReadAny()
		// _ = json.Get("tick")
	}
}

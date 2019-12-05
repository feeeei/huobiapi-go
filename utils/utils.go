package utils

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"io/ioutil"
	"net/url"
	"sort"
	"strings"
	"time"
)

// EncodeQueryString 拼接query字符串
func EncodeQueryString(params map[string]interface{}) string {
	var keys = sortKeys(getMapKeys(params))
	var lines = make([]string, len(keys))
	for i := 0; i < len(lines); i++ {
		var k = keys[i]
		b, _ := json.Marshal(params[k])
		if b[0] == '"' {
			b = b[1:]
		}
		if b[len(b)-1] == '"' {
			b = b[:len(b)-1]
		}
		lines[i] = url.QueryEscape(k) + "=" + url.QueryEscape(string(b))
	}
	return strings.Join(lines, "&")
}

func getMapKeys(params map[string]interface{}) (keys []string) {
	for k := range params {
		keys = append(keys, k)
	}
	return keys
}

func sortKeys(keys []string) []string {
	sort.Strings(keys)
	return keys
}

// Sign 签名请求
func Sign(method, host, path, secretKey string, params map[string]interface{}) string {
	var str = method + "\n" + host + "\n" + path + "\n"
	str += EncodeQueryString(params)
	return ComputeHmac256(str, secretKey)
}

// ComputeHmac256 HMAC SHA256加密
func ComputeHmac256(str string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(str))

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// MergeMap 合并两个map
func MergeMap(map1, map2 map[string]interface{}) map[string]interface{} {
	for k := range map2 {
		map1[k] = map2[k]
	}
	return map1
}

// UinxMillisecond 取毫秒时间戳
func UinxMillisecond() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// DecodeGzip 解压gzip数据
func DecodeGzip(data []byte) ([]byte, error) {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.LittleEndian, data)
	r, err := gzip.NewReader(buffer)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return ioutil.ReadAll(r)
}

package sign

import "time"

type Sign struct {
	AccessKeyID      string
	AccessKeySecret  string
	SignatureMethod  string
	SignatureVersion string
}

func NewSign(accessKeyID, accessKeySecret string) *Sign {
	return &Sign{
		AccessKeyID:      accessKeyID,
		AccessKeySecret:  accessKeySecret,
		SignatureMethod:  "HmacSHA256",
		SignatureVersion: "2",
	}
}

func (s *Sign) GetSignFields() map[string]interface{} {
	return map[string]interface{}{
		"AccessKeyId":      s.AccessKeyID,
		"SignatureMethod":  s.SignatureMethod,
		"SignatureVersion": s.SignatureVersion,
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05"),
	}
}

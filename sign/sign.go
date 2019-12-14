package sign

import (
	"time"
)

type Sign struct {
	AccessKeyID      string
	AccessKeySecret  string
	SignatureMethod  string
	SignatureVersion string
}

func NewSign(accessKeyID, accessKeySecret, version string) *Sign {
	return &Sign{
		AccessKeyID:      accessKeyID,
		AccessKeySecret:  accessKeySecret,
		SignatureMethod:  "HmacSHA256",
		SignatureVersion: version,
	}
}

func (s *Sign) GetSignFields() map[string]interface{} {
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05")
	if s.SignatureVersion == "2.1" {
		return map[string]interface{}{
			"accessKey":        s.AccessKeyID,
			"signatureMethod":  s.SignatureMethod,
			"signatureVersion": s.SignatureVersion,
			"timestamp":        timestamp,
		}
	}

	// SignatureVersion == 2
	return map[string]interface{}{
		"AccessKeyId":      s.AccessKeyID,
		"SignatureMethod":  s.SignatureMethod,
		"SignatureVersion": s.SignatureVersion,
		"Timestamp":        timestamp,
	}
}

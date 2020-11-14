package crypto

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"io"
)

func Hmc(raw, key string) string {
	hm := hmac.New(sha1.New, []byte(key))
	hm.Write([]byte(raw))
	return base64.RawURLEncoding.EncodeToString(hm.Sum(nil))
}

func CheckPassWord(passwd, md5Pass string) string {
	if len(md5Pass) != 48 {
		return ""
	}
	salt := make([]byte, 16)
	orgMd5 := make([]byte, 32)
	for i, k := 0, 0; i < 48; i ++ {
		if i % 3 == 0 {
			salt[i/3] = md5Pass[i]
		} else {
			orgMd5[k] = md5Pass[i]
			k ++
		}
	}
	h := md5.New()
	io.WriteString(h, string(salt)+passwd)
	pre := hex.EncodeToString(h.Sum(nil))
	res := make([]byte, 48)
	for i, k := 0, 0; i < 48; i ++ {
		if i % 3 == 0 {
			res[i] = salt[i/3]
		} else {
			res[i] = pre[k]
			k ++
		}
	}
	return string(res)
}
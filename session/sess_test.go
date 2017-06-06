package session

import (
	"crypto/aes"
	"encoding/json"
	"testing"
)

func Test_gob(t *testing.T) {
	a := make(map[interface{}]interface{})
	a["username"] = "insionng"
	a[12] = 234
	a["user"] = User{"insion", "ng"}
	b, err := EncodeGob(a)
	if err != nil {
		t.Error(err)
	}
	c, err := DecodeGob(b)
	if err != nil {
		t.Error(err)
	}
	if len(c) == 0 {
		t.Error("decodeGob empty")
	}
	if c["username"] != "insionng" {
		t.Error("decode string error")
	}
	if c[12] != 234 {
		t.Error("decode int error")
	}
	if c["user"].(User).Username != "insion" {
		t.Error("decode struct error")
	}
}

type User struct {
	Username string
	NickName string
}

func TestGenerate(t *testing.T) {
	str := generateRandomKey(20)
	if len(str) != 20 {
		t.Fatal("generate length is not equal to 20")
	}
}

func TestCookieEncodeDecode(t *testing.T) {
	hashKey := "testhashKey"
	blockkey := generateRandomKey(16)
	block, err := aes.NewCipher(blockkey)
	if err != nil {
		t.Fatal("NewCipher:", err)
	}
	securityName := string(generateRandomKey(20))
	val := make(map[interface{}]interface{})
	val["name"] = "insionng"
	val["gender"] = "male"
	str, err := encodeCookie(block, hashKey, securityName, val)
	if err != nil {
		t.Fatal("encodeCookie:", err)
	}
	dst := make(map[interface{}]interface{})
	dst, err = decodeCookie(block, hashKey, securityName, str, 3600)
	if err != nil {
		t.Fatal("decodeCookie", err)
	}
	if dst["name"] != "insionng" {
		t.Fatal("dst get map error")
	}
	if dst["gender"] != "male" {
		t.Fatal("dst get map error")
	}
}

func TestParseConfig(t *testing.T) {
	s := `{"cookieName":"makrossSessionId","gcLifetime":3600}`
	cf := new(managerConfig)
	cf.EnableSetCookie = true
	err := json.Unmarshal([]byte(s), cf)
	if err != nil {
		t.Fatal("parse json error,", err)
	}
	if cf.CookieName != "makrossSessionId" {
		t.Fatal("parseconfig get cookiename error")
	}
	if cf.GcLifetime != 3600 {
		t.Fatal("parseconfig get gcLifetime error")
	}

	cc := `{"cookieName":"makrossSessionId","enableSetCookie":false,"gcLifetime":3600,"providerConfig":"{\"cookieName\":\"makrossSessionId\",\"securityKey\":\"makrosscookiehashkey\"}"}`
	cf2 := new(managerConfig)
	cf2.EnableSetCookie = true
	err = json.Unmarshal([]byte(cc), cf2)
	if err != nil {
		t.Fatal("parse json error,", err)
	}
	if cf2.CookieName != "makrossSessionId" {
		t.Fatal("parseconfig get cookiename error")
	}
	if cf2.GcLifetime != 3600 {
		t.Fatal("parseconfig get gcLifetime error")
	}
	if cf2.EnableSetCookie != false {
		t.Fatal("parseconfig get enableSetCookie error")
	}
	cconfig := new(cookieConfig)
	err = json.Unmarshal([]byte(cf2.ProviderConfig), cconfig)
	if err != nil {
		t.Fatal("parse providerConfig err,", err)
	}
	if cconfig.CookieName != "makrossSessionId" {
		t.Fatal("providerConfig get cookieName error")
	}
	if cconfig.SecurityKey != "makrosscookiehashkey" {
		t.Fatal("providerConfig get securityKey error")
	}
}

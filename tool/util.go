package tool

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	cfg "go-proxy/config"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

/*

  http.ReadRequest()
  http.ReadResponse()
  httputil.DumpRequest()
  httputil.DumpResponse()
*/

// 将request 的处理
func EncodeRequest(r *http.Request) ([]byte, error) {
	raw := bytes.NewBuffer([]byte{})
	// 写签名
	_ = binary.Write(raw, binary.LittleEndian, []byte("sign"))
	reqBytes, err := httputil.DumpRequest(r, true)
	if err != nil {
		return nil, err
	}
	// 写body数据长度 + 1
	_ = binary.Write(raw, binary.LittleEndian, int32(len(reqBytes)+1))
	// 判断是否为http或者https的标识1字节
	_ = binary.Write(raw, binary.LittleEndian, bool(r.URL.Scheme == "https"))
	if err := binary.Write(raw, binary.LittleEndian, reqBytes); err != nil {
		return nil, err
	}
	return raw.Bytes(), nil
}

// 将字节转为request
func DecodeRequest(data []byte, siteList []cfg.Tunnel) (*http.Request, error) {
	if len(data) <= 100 {
		return nil, errors.New("The byte length to be decoded is too small.")
	}
	req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(data[1:])))
	if err != nil {
		return nil, err
	}
	str := strings.Split(req.Host, ":")
	req.Host, err = getHost(str[0], siteList)
	if err != nil {
		return nil, err
	}
	scheme := "http"
	if data[0] == 1 {
		scheme = "https"
	}
	req.URL, _ = url.Parse(fmt.Sprintf("%s://%s%s", scheme, req.Host, req.RequestURI))
	req.RequestURI = ""

	return req, nil
}

//// 将response转为字节
func EncodeResponse(r *http.Response, domainAsProto bool, siteList []cfg.Tunnel) ([]byte, error) {
	raw := bytes.NewBuffer([]byte{})
	_ = binary.Write(raw, binary.LittleEndian, []byte("sign"))
	respBytes, err := httputil.DumpResponse(r, true)
	if domainAsProto {
		respBytes = replaceHost(respBytes, siteList)
	}
	if err != nil {
		return nil, err
	}
	_ = binary.Write(raw, binary.LittleEndian, int32(len(respBytes)))
	if err := binary.Write(raw, binary.LittleEndian, respBytes); err != nil {
		return nil, err
	}
	return raw.Bytes(), nil
}

//// 将字节转为response
func DecodeResponse(data []byte) (*http.Response, error) {

	resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(data)), nil)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func getHost(str string, tunnel []cfg.Tunnel) (string, error) {
	for _, v := range tunnel {
		if v.Domain == str {
			return v.Proto, nil
		}
	}
	return "", errors.New("The resolved host was not found.")
}

func replaceHost(resp []byte, siteList []cfg.Tunnel) []byte {
	str := string(resp)
	for _, v := range siteList {
		str = strings.Replace(str, v.Proto, v.Domain, -1)
		// str = strings.Replace(str, v.LocalHost, v.RemoteHost, -1)
	}
	return []byte(str)
}

// 一个简单的校验值
func CheckValue(vkey string) []byte {
	b := sha1.Sum([]byte(time.Now().Format("2006-01-02 15") + vkey))
	return b[:]
}

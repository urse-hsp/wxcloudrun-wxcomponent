package httputils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"

	"github.com/WeixinCloud/wxcloudrun-wxcomponent/comm/config"
	"github.com/WeixinCloud/wxcloudrun-wxcomponent/comm/log"
)

// PostJson 发起post请求，格式为JSON
func PostJson(url string, data interface{}) ([]byte, error) {
	jsonByte, _ := json.Marshal(data)
	return Post(url, jsonByte, "application/json")
}

// PostFormData 发起post请求，格式为Form-Data
func PostFormData(url string, formFile multipart.File, fileName string, fieldName string) ([]byte, error) {
	client := &http.Client{Timeout: 20 * time.Second}
	log.Debugf("[post]http url: %s", safeURL(url))

	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	if createFormFile, err := w.CreateFormFile(fieldName, fileName); err == nil {
		readAll, _ := ioutil.ReadAll(formFile)
		createFormFile.Write(readAll)
	}
	w.Close()
	req, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("User-Agent", "WxComponent/"+config.CommConf.Version)
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http code: %d", resp.StatusCode)
	}

	result, _ := ioutil.ReadAll(resp.Body)
	log.Debugf("[post]http response bytes: %d", len(result))
	return result, nil
}

// Post 发起post请求
func Post(url string, data []byte, contentType string) ([]byte, error) {
	_, body, err := RawPost(url, data, contentType)
	return body, err
}

// RawPost 发起post请求
func RawPost(url string, data []byte, contentType string) (*http.Response, []byte, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	log.Debugf("[post]http url: %s content-type: %s request bytes: %d", safeURL(url), contentType, len(data))
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		log.Error(err)
		return nil, nil, err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("User-Agent", "WxComponent/"+config.CommConf.Version)
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return resp, nil, fmt.Errorf("http code: %d", resp.StatusCode)
	}

	result, _ := ioutil.ReadAll(resp.Body)
	log.Debugf("[post]http response bytes: %d", len(result))
	return resp, result, nil
}

// Get 发起get请求
func Get(url string) ([]byte, error) {
	_, body, err := RawGet(url)
	return body, err
}

// RawGet 发起get请求
func RawGet(url string) (*http.Response, []byte, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	log.Debugf("[get]http url: %s", safeURL(url))
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Error(err)
		return nil, nil, err
	}
	req.Header.Set("User-Agent", "WxComponent/"+config.CommConf.Version)
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return resp, nil, fmt.Errorf("http code: %d", resp.StatusCode)
	}
	result, _ := ioutil.ReadAll(resp.Body)
	log.Debugf("[get]http response bytes: %d", len(result))
	return resp, result, nil
}

// safeURL omits query parameters because they often carry access tokens.
func safeURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "[invalid URL]"
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}

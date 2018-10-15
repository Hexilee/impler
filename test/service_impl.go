/*
Implement of test.Service
This file is generated by github.com/Hexilee/impler at 2018-10-15 19:23:13.416699 +0800 CST m=+0.337040922
DON'T EDIT IT!
*/

package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func NewService() Service {
	service := &serviceImpl{
		baseUrl: "https://box.zjuqsc.com/item",
		cookies: make([]*http.Cookie, 0),
		header:  make(http.Header),
	}
	service.header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36")
	service.cookies = append(service.cookies, &http.Cookie{
		Name:  "ga",
		Value: "bb78137vt73817ynyh89",
	})
	service.cookies = append(service.cookies, &http.Cookie{
		Name:  "qsc_session",
		Value: "secure_7y7y1n570y",
	})
	return service
}

type serviceImpl struct {
	baseUrl string
	header  http.Header
	cookies []*http.Cookie
}

func (service serviceImpl) UploadItem(path string, contentType string, cookie string, video io.Reader) (genResult *http.Response, genErr error) {
	var genBody io.ReadWriter
	var genRequest *http.Request
	genUri := "/upload"
	var genBodyWriter *multipart.Writer
	genBody = bytes.NewBufferString("")
	genBodyWriter = multipart.NewWriter(genBody)
	{
		var genPartWriter io.Writer
		var genFile *os.File
		var genFilePath string
		genFilePath = fmt.Sprintf("/var/log/%s", path)
		genFile, genErr = os.Open(genFilePath)
		defer genFile.Close()
		if genErr != nil {
			return
		}
		genPartWriter, genErr = genBodyWriter.CreateFormFile("avatar", genFilePath)
		if genErr != nil {
			return
		}
		_, genErr = io.Copy(genPartWriter, genFile)
		if genErr != nil {
			return
		}
	}
	{
		var genPartWriter io.Writer
		genPartWriter, genErr = genBodyWriter.CreateFormField("video")
		if genErr != nil {
			return
		}
		_, genErr = io.Copy(genPartWriter, video)
		if genErr != nil {
			return
		}
	}
	genBodyWriter.Close()
	genFinalUrl := strings.TrimRight(service.baseUrl, "/") + "/" + strings.TrimLeft(genUri, "/")
	genRequest, genErr = http.NewRequest("POST", genFinalUrl, genBody)
	if genErr != nil {
		return
	}
	for genHeaderKey, genHeaderSlice := range service.header {
		for _, genHeaderValue := range genHeaderSlice {
			genRequest.Header.Add(genHeaderKey, genHeaderValue)
		}
	}
	genRequest.Header.Set("ga", cookie)
	for _, genCookie := range service.cookies {
		genRequest.AddCookie(genCookie)
	}
	genRequest.Header.Set("Content-Type", genBodyWriter.FormDataContentType())
	var genResponse *http.Response
	genClient := new(http.Client)
	genResponse, genErr = genClient.Do(genRequest)
	if genErr != nil {
		return
	}
	genResult = genResponse
	return
}
func (service serviceImpl) GetItem(token int, page int, limit int) (genResult *http.Response, genErr error) {
	var genBody io.ReadWriter
	var genRequest *http.Request
	genUri := fmt.Sprintf("/get/%d?page=%d&limit=%d", token, page, limit)
	genFinalUrl := strings.TrimRight(service.baseUrl, "/") + "/" + strings.TrimLeft(genUri, "/")
	genRequest, genErr = http.NewRequest("GET", genFinalUrl, genBody)
	if genErr != nil {
		return
	}
	for genHeaderKey, genHeaderSlice := range service.header {
		for _, genHeaderValue := range genHeaderSlice {
			genRequest.Header.Add(genHeaderKey, genHeaderValue)
		}
	}
	for _, genCookie := range service.cookies {
		genRequest.AddCookie(genCookie)
	}
	var genResponse *http.Response
	genClient := new(http.Client)
	genResponse, genErr = genClient.Do(genRequest)
	if genErr != nil {
		return
	}
	genResult = genResponse
	return
}
func (service serviceImpl) PostInfo(id int, firstName string) (genResult *http.Request, genErr error) {
	var genBody io.ReadWriter
	var genRequest *http.Request
	genUri := ""
	genDataMap := make(url.Values)
	genDataMap.Add("name", fmt.Sprintf("%s.Lee", firstName))
	genDataMap.Add("id", fmt.Sprintf("%d", id))
	genBody = bytes.NewBufferString(genDataMap.Encode())
	genFinalUrl := strings.TrimRight(service.baseUrl, "/") + "/" + strings.TrimLeft(genUri, "/")
	genRequest, genErr = http.NewRequest("POST", genFinalUrl, genBody)
	if genErr != nil {
		return
	}
	for genHeaderKey, genHeaderSlice := range service.header {
		for _, genHeaderValue := range genHeaderSlice {
			genRequest.Header.Add(genHeaderKey, genHeaderValue)
		}
	}
	for _, genCookie := range service.cookies {
		genRequest.AddCookie(genCookie)
	}
	genRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	genResult = genRequest
	return
}
func (service serviceImpl) StatByReader(id int, body io.Reader) (genResult *http.Response, genErr error) {
	var genBody io.ReadWriter
	var genRequest *http.Request
	genUri := fmt.Sprintf("/stat/%d", id)
	genBody = bytes.NewBufferString("")
	_, genErr = io.Copy(genBody, body)
	if genErr != nil {
		return
	}
	genFinalUrl := strings.TrimRight(service.baseUrl, "/") + "/" + strings.TrimLeft(genUri, "/")
	genRequest, genErr = http.NewRequest("POST", genFinalUrl, genBody)
	if genErr != nil {
		return
	}
	for genHeaderKey, genHeaderSlice := range service.header {
		for _, genHeaderValue := range genHeaderSlice {
			genRequest.Header.Add(genHeaderKey, genHeaderValue)
		}
	}
	for _, genCookie := range service.cookies {
		genRequest.AddCookie(genCookie)
	}
	genRequest.Header.Set("Content-Type", "application/json; charset=UTF-8")
	var genResponse *http.Response
	genClient := new(http.Client)
	genResponse, genErr = genClient.Do(genRequest)
	if genErr != nil {
		return
	}
	genResult = genResponse
	return
}
func (service serviceImpl) StatItem(id int, body *StatBody) (genResult *http.Response, genErr error) {
	var genBody io.ReadWriter
	var genRequest *http.Request
	genUri := fmt.Sprintf("/stat/%d", id)
	var genData []byte
	genData, genErr = json.Marshal(body)
	if genErr != nil {
		return
	}
	genBody = bytes.NewBuffer(genData)
	genFinalUrl := strings.TrimRight(service.baseUrl, "/") + "/" + strings.TrimLeft(genUri, "/")
	genRequest, genErr = http.NewRequest("POST", genFinalUrl, genBody)
	if genErr != nil {
		return
	}
	for genHeaderKey, genHeaderSlice := range service.header {
		for _, genHeaderValue := range genHeaderSlice {
			genRequest.Header.Add(genHeaderKey, genHeaderValue)
		}
	}
	for _, genCookie := range service.cookies {
		genRequest.AddCookie(genCookie)
	}
	genRequest.Header.Set("Content-Type", "application/json; charset=UTF-8")
	var genResponse *http.Response
	genClient := new(http.Client)
	genResponse, genErr = genClient.Do(genRequest)
	if genErr != nil {
		return
	}
	genResult = genResponse
	return
}
func (service serviceImpl) UpdateItem(id int, cookie string, data *time.Time, apiKey string) (genResult *UploadResult, genStatusCode int, genErr error) {
	var genBody io.ReadWriter
	var genRequest *http.Request
	genUri := fmt.Sprintf("/change/%d", id)
	var genData []byte
	genDataMap := make(map[string]interface{})
	genDataMap["data"] = data
	genDataMap["apiKey"] = apiKey
	genData, genErr = json.Marshal(genDataMap)
	if genErr != nil {
		return
	}
	genBody = bytes.NewBuffer(genData)
	genFinalUrl := strings.TrimRight(service.baseUrl, "/") + "/" + strings.TrimLeft(genUri, "/")
	genRequest, genErr = http.NewRequest("PUT", genFinalUrl, genBody)
	if genErr != nil {
		return
	}
	for genHeaderKey, genHeaderSlice := range service.header {
		for _, genHeaderValue := range genHeaderSlice {
			genRequest.Header.Add(genHeaderKey, genHeaderValue)
		}
	}
	genRequest.Header.Set("ga", cookie)
	for _, genCookie := range service.cookies {
		genRequest.AddCookie(genCookie)
	}
	genRequest.Header.Set("Content-Type", "application/json; charset=UTF-8")
	var genResponse *http.Response
	genClient := new(http.Client)
	genResponse, genErr = genClient.Do(genRequest)
	if genErr != nil {
		return
	}
	var genResultData []byte
	genResultData, genErr = ioutil.ReadAll(genResponse.Body)
	defer genResponse.Body.Close()
	if genErr != nil {
		return
	}
	genStatusCode = genResponse.StatusCode
	genResult = &UploadResult{}
	genErr = json.Unmarshal(genResultData, genResult)
	if genErr != nil {
		return
	}
	return
}

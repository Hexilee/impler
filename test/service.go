package test

import (
	"io"
	"net/http"
	"time"
)

//go:generate go run ../main.go Service

/*
@Base https://box.zjuqsc.com/item
@Header(User-Agent) Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36
@Cookie(ga) bb78137vt73817ynyh89
@Cookie(qsc_session) secure_7y7y1n570y
*/
type Service interface {
	/*
	@Get /get/{token}?page={page}&limit={limit}
	 */
	GetItem(token int, page int, limit int) (*http.Response, error)

	/*
	@Post /upload
	@Body multipart
	@Header(Content-Type) {contentType}
	@Cookie(ga) {cookie}
	@File(avatar) /var/log/{path}
	 */
	UploadItem(path string, contentType string, cookie string, video io.Reader) (*http.Response, error)

	/*
	@Put /change/{id}
	@Body json
	@Cookie(ga) {cookie}
	@Result json
	 */
	UpdateItem(id int, cookie string, data *time.Time, apiKey string) (result *UploadResult, statusCode int, err error)

	/*
	@Get /stat/{id}
	@SingleBody json
	 */
	StatItem(id int, body *StatBody) (*http.Response, error)

	/*
	@Get /stat/{id}
	@SingleBody json
	 */
	StatByReader(id int, body io.Reader) (*http.Response, error)

	/*
	@Post /
	@Body form
	@Param(name) {firstName}.Lee
	 */
	PostInfo(id int, firstName string) (*http.Request, error)
}

type UploadResult struct {
}

type StatBody struct {
}

package test

import (
	"net/http"
	"time"
)

//go:generate go run ../main.go Service

/*
@Base https://box.zjuqsc.com/item
@Header(User-Agent) Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36
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
	@File(avatar) {path}
	 */
	UploadItem(path string, contentType string, cookie string) (*http.Response, error)

	/*
	@Put /change/{id}
	@Body json
	@Cookie(ga) {cookie}
	@Result json
	 */
	UpdateItem(id int, cookie string, data *time.Time) (result *UploadResult, statusCode int, err error)
}

type UploadResult struct {
}

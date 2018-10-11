package test

import (
	"net/http"
	"time"
)

//go:generate go run ../main.go Service

/*
@Base https://box.zjuqsc.com/item
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
	@Param(path) FilePath
	 */
	UploadItem(path string, contentType string, cookie string) (*http.Response, error)

	/*
	@Put /change/{id}
	@Body json
	@Cookie(ga) {cookie}
	 */
	UpdateItem(id int, cookie string, data *time.Time) (*http.Response, error)
}
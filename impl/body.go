package impl

const (
	// only for response

	// resp, error
	HttpResponse = "http_response"

	// req, error
	HttpRequest = "http_request"

	// resp, statusCode, error
	HTML = "html"
)

const (
	// for request or response

	// resp, statusCode, error
	JSON = "json"

	// resp, statusCode, error
	XML = "xml"
)

const (
	// only for request

	// resp, statusCode, error
	Form = "form"

	// resp, statusCode, error
	Multipart = "multipart"
)

type (
	BodyType string
)

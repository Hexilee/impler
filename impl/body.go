package impl

const (
	// only for response

	// resp, error
	HttpResponse BodyType = iota

	// req, error
	HttpRequest

	// resp, statusCode, error
	HTML
)

const (
	// for request or response

	// resp, statusCode, error
	JSON BodyType = iota + 3

	// resp, statusCode, error
	XML
)

const (
	// only for request

	// resp, statusCode, error
	Form BodyType = iota + 5

	// resp, statusCode, error
	Multipart
)

type (
	BodyType int
)

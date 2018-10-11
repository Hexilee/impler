package impl

const (
	// only for response

	// resp, error
	HttpResponse = iota

	// req, error
	HttpRequest

	// resp, statusCode, error
	HTML
)

const (
	// for request or response

	// resp, statusCode, error
	JSON = iota + 3

	// resp, statusCode, error
	XML
)

const (
	// only for request

	// resp, statusCode, error
	Form = iota + 5

	// resp, statusCode, error
	Multipart
)

type (
	BodyType int
)

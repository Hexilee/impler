package impl

const (
	// only for response

	// resp, error
	HttpResponse = iota + 1

	// req, error
	HttpRequest

	// resp, statusCode, error
	HTML
)

const (
	// for request or response

	// resp, statusCode, error
	JSON = iota + 4

	// resp, statusCode, error
	XML
)

const (
	// only for request

	// resp, statusCode, error
	Form = iota + 6

	// resp, statusCode, error
	Multipart
)

type (
	BodyType int
)

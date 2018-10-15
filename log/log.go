package log

import (
	"github.com/op/go-logging"
	"os"
)

var (
	Log    = logging.MustGetLogger("example")
	format = logging.MustStringFormatter(
		`%{color} > %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)
	backend          = logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter = logging.NewBackendFormatter(backend, format)
)

func init() {
	logging.SetBackend(backendFormatter)
}

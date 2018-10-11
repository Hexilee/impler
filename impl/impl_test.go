package impl

import (
	"fmt"
	"testing"
)

func TestGetQual(t *testing.T) {
	fmt.Printf("%#v", getQual("*net/http.Response"))
}

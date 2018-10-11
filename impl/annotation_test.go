package impl

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProcessor_Scan(t *testing.T) {
	i := 0
	NewProcessor(`
//
@Page(name) {name}
@Body json
@File(avatar)
`).Scan(func(ann, key, value string) (err error) {
		i++
		count := i
		switch count {
		case 1:
			assert.Equal(t, "@Page", ann)
			assert.Equal(t, "name", key)
			assert.Equal(t, "{name}", value)
		case 2:
			assert.Equal(t, "@Body", ann)
			assert.Equal(t, "", key)
			assert.Equal(t, "json", value)
		case 3:
			assert.Equal(t, "@File", ann)
			assert.Equal(t, "avatar", key)
			assert.Equal(t, "", value)

		}
		return
	})
}

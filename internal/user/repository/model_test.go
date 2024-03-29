package repo

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBlacklist(t *testing.T) {
	// Проверяем, есть ли запретное слово в блеклисте
	var cases = []struct {
		name  string
		field string
		check bool
	}{
		{name: "Check Id Field", field: "id", check: true},
		{name: "Check password Field", field: "password", check: true},
		{name: "Check Test Field", field: "test", check: false},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			assert.Equal(t, tCase.check, BlackListCheck(tCase.field))
		})
	}
}

package belogs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBeeLoggerDelLogger(t *testing.T) {
	prefix := "My-Cus"
	l := GetLogger(prefix)
	assert.NotNil(t, l)
	l.Print("hello")

	GetLogger().Print("hello")
	SetPrefix("aaa")
	Info("hello")
}

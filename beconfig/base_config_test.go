package config

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBaseConfiger_DefaultBool(t *testing.T) {
	bc := newBaseConfier("true")
	assert.True(t, bc.DefaultBool("key1", false))
	assert.True(t, bc.DefaultBool("key2", true))
}

func TestBaseConfiger_DefaultFloat(t *testing.T) {
	bc := newBaseConfier("12.3")
	assert.Equal(t, 12.3, bc.DefaultFloat("key1", 0.1))
	assert.Equal(t, 0.1, bc.DefaultFloat("key2", 0.1))
}

func TestBaseConfiger_DefaultInt(t *testing.T) {
	bc := newBaseConfier("10")
	assert.Equal(t, 10, bc.DefaultInt("key1", 8))
	assert.Equal(t, 8, bc.DefaultInt("key2", 8))
}

func TestBaseConfiger_DefaultInt64(t *testing.T) {
	bc := newBaseConfier("64")
	assert.Equal(t, int64(64), bc.DefaultInt64("key1", int64(8)))
	assert.Equal(t, int64(8), bc.DefaultInt64("key2", int64(8)))
}

func TestBaseConfiger_DefaultString(t *testing.T) {
	bc := newBaseConfier("Hello")
	assert.Equal(t, "Hello", bc.DefaultString("key1", "world"))
	assert.Equal(t, "world", bc.DefaultString("key2", "world"))
}

func TestBaseConfiger_DefaultStrings(t *testing.T) {
	bc := newBaseConfier("Hello;world")
	assert.Equal(t, []string{"Hello", "world"}, bc.DefaultStrings("key1", []string{"world"}))
	assert.Equal(t, []string{"world"}, bc.DefaultStrings("key2", []string{"world"}))
}

func newBaseConfier(str1 string) *BaseConfiger {
	return &BaseConfiger{
		reader: func(ctx context.Context, key string) (string, error) {
			if key == "key1" {
				return str1, nil
			} else {
				return "", errors.New("mock error")
			}
		},
	}
}

package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKey(t *testing.T) {
	type args struct {
		k string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{k: ""}, "flipt:v1:d41d8cd98f00b204e9800998ecf8427e"},
		{"non-empty", args{k: "foo"}, "flipt:v1:acbd18db4cc2f85cedef654fccc4a4d8"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Key(tt.args.k))
		})
	}
}

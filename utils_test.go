package main

import "testing"

func Test_funcName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"Ok", "github.com/patrickalin/bloomsky-client-go.Test_funcName.func1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := funcName(); got != tt.want {
				t.Errorf("funcName() = %v, want %v", got, tt.want)
			}
		})
	}
}

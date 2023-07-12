package main

import (
	"net/http"
	"testing"
)

func TestNoSurf(t *testing.T) {
	var handler myHandler
	h := NoSurf(&handler)

	switch v := h.(type) {
	case http.Handler:
		//do nothing
	default:
		t.Errorf("Type is not http.Handler, instead type is %T", v)
	}
}
func TestSessionLoad(t *testing.T) {
	var handler myHandler
	h := SessionLoad(&handler)

	switch v := h.(type) {
	case http.Handler:
		//do nothing
	default:
		t.Errorf("Type is not http.Handler, instead type is %T", v)
	}
}

// +build !linux

package main

import "net/http"

func (t *Transport) InteralTOR() *http.Transport {
	return &http.Transport{}
}

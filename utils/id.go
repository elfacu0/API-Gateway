package utils

import (
	"crypto/rand"
	"fmt"
)

func ID(path, method string) string {
	return path + "_" + method
}

func RandomStr() string {
	b := make([]byte, 10)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	str := fmt.Sprintf("%x", b)
	return str
}

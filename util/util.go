package util

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// M is a generic map
type M map[string]interface{}

var (
	Letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	Digits  = []byte("0123456789")
	Symbols = []byte("!@#$%^&*-=_.?")
)

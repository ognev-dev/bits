package util

import (
	"math/rand"
)

// RandomBool generates truth randomly.
// Just like me when I went on a date with my now wife
// The truth was big like this func, but it works
func RandomBool() bool {
	return rand.Intn(2) == 1
}

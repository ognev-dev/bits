package util

import (
	"strings"
)

func IsEmptyString(in string) bool {
	return strings.TrimSpace(in) == ""
}

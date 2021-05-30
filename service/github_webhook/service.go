package githubwebhook

import (
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
)

const PushEvent = "push"
const PullRequestEvent = "pull_request"

// SigPrefix
// GitHub adds a prefix "sha1=" to the hash
const SigPrefix = "sha1="

// MakeHMAC generates hmac of given body and key
// https://developer.github.com/webhooks/event-payloads/#delivery-headers
// The HMAC hex digest of the response body.
// This header will be sent if the webhook is configured with a secret.
// The HMAC hex digest is generated using the sha1 hash function and the secret as the HMAC key.
// Note: github also adds a prefix "sha1=" to the hash
func MakeHMAC(body, key string) (hash string, err error) {
	h := hmac.New(sha1.New, []byte(key))
	_, err = h.Write([]byte(body))
	if err != nil {
		return
	}

	hash = SigPrefix + fmt.Sprintf("%+x", h.Sum(nil))
	return
}

func IsValidHMAC(body []byte, hash, key string) (ok bool, err error) {
	h := hmac.New(sha1.New, []byte(key))
	_, err = h.Write(body)
	if err != nil {
		return
	}

	expected := fmt.Sprintf("%x", h.Sum(nil))
	ok = hash == (SigPrefix + expected)

	return
}

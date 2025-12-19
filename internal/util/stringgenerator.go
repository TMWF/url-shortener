package util

import (
	"math/rand"
	"strings"
	"time"
)

const LatinCharSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandomString(stringLength int, charset string) string {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	var sb strings.Builder
	sb.Grow(stringLength)

	for i := 0; i < stringLength; i++ {
		randomChar := charset[r.Intn(len(charset))]
		sb.WriteByte(randomChar)
	}
	return sb.String()
}

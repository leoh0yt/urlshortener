package encoder

import (
	"fmt"
	"slices"
	"strings"
)

const (
	base62Charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

type Encoder struct {
	charset string
	base    uint64
	padding string
}

func NewBase62Encoder() *Encoder {
	return &Encoder{
		charset: base62Charset,
		base:    uint64(len(base62Charset)),
		padding: "0",
	}
}

func (e *Encoder) Encode10(val uint64) string {
	if val == 0 {
		return string(e.charset[0])
	}

	var result []byte

	for val > 0 {
		rem := val % e.base
		result = append([]byte{base62Charset[rem]}, result...)
		val /= e.base
	}

	return string(result)
}

func (e *Encoder) Decode10(s string) (uint64, error) {
	trimmedStr := strings.TrimLeft(s, e.padding)

	var (
		id  uint64
		pow uint64
	)
	pow = 1

	for _, char := range slices.Backward([]rune(trimmedStr)) {
		pos := strings.IndexRune(e.charset, char)
		if pos == -1 {
			return 0, fmt.Errorf("invalid char: %c", char)
		}
		id += uint64(pos) * pow
		pow *= e.base
	}

	return id, nil
}

func (e *Encoder) Encode10WithPadding(val uint64, minSize int) string {
	encoded := e.Encode10(val)
	length := len(encoded)

	if length < minSize {
		encoded = strings.Repeat(e.padding, minSize-length) + encoded
	}

	return encoded
}

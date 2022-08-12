package layerhub

import (
	"fmt"
	"time"

	gonanoid "github.com/matoous/go-nanoid"
)

const (
	randomStringAlphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	regularIndexSize     = 14
	shortIndexSize       = 8
)

func UniqueID(prefix string) string {
	id := RandomString(regularIndexSize)
	return fmt.Sprintf("%s_%s", prefix, id)
}

func UniqueShortID() string {
	return RandomString(shortIndexSize)
}

func IsShortID(id string) bool {
	return len(id) == shortIndexSize
}

func RandomString(size int) string {
	s, err := gonanoid.Generate(randomStringAlphabet, size)
	if err != nil {
		panic(err)
	}
	return s
}

func Now() time.Time {
	return time.Now().UTC().Truncate(time.Second)
}

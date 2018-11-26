package edasim

import (
	"fmt"
	"math/rand"
	"time"
)

// implementation derived from https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go

var LetterRunes []rune = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func init() {
	fmt.Printf("run job init")
    rand.Seed(time.Now().UnixNano())
}

func RandStringRunes(n int) string {
    b := make([]rune, n)
    for i := range b {
        b[i] = LetterRunes[rand.Intn(len(LetterRunes))]
    }
    return string(b)
}
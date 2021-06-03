package util

import (
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandomInt generates a random integer number between the min and max numerric argumants passed to it
func RandomInt(min, max int64) int64 {

	return min + rand.Int63n(max-min+1)
}

// RandomString generates a random string of a speficied length as that of its give argumants
func RandomString(n int) string {

	var sb strings.Builder

	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)

	}

	return sb.String()

}

func RandomOwner() string {
	return RandomString(6)

}

func RandomMoney() int64 {
	return RandomInt(10, 100)
}

// RandomCurrency
func RandomCurrency() string {
	currencies := []string{"USD", "NGR", "CAD"}
	n := len(currencies)
	return currencies[rand.Intn(n)]
}

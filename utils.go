package partida

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"os"
	"time"
)

// Empty is a empty interface literal type to pass null as a parameter
var Empty interface{}

// ParseJSON parse a JSON file to an interface
func ParseJSON(path string, data interface{}) {
	var err error

	// try to open the JSON file
	file, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	// parse JSON file to interface
	err = json.Unmarshal([]byte(file), data)
	if err != nil {
		panic(err)
	}

}

// IsAssociative check if the informed interface is an map
func IsAssociative(v interface{}) bool {
	var is bool
	switch v.(type) {
	case map[string]interface{}:
		is = true
	default:
		is = false
	}
	return is
}

// StringInSlice check if the informed string is on the slice
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// RandomString generate a random string of the passed length
func RandomString(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano() - int64(rand.Int())))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(b)
}

// StringToSHA256 convert a string into its SHA256 hash
func StringToSHA256(str string) string {
	h := sha256.New()
	h.Write([]byte(str))
	final := hex.EncodeToString(h.Sum(nil))
	return final
}

// StringToSHA512 convert a string into its SHA512 hash
func StringToSHA512(str string) string {
	h := sha512.New()
	h.Write([]byte(str))
	final := hex.EncodeToString(h.Sum(nil))
	return final
}

package partida

import (
	"encoding/json"
	"os"
)

// an empty interface literal type, used to pass null as a parameter
var Empty interface{}

// parse a JSON file to an interface
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

// check if the informed interface is an map
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

// check if the informed string is on the slice
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

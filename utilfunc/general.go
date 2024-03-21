package utilfunc

// StringInSlice
// check if the informed string is on the passed slice
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}

	return false
}

// IsMap
// check if the informed interface is an map
func IsMap(v interface{}) bool {
	switch v.(type) {
	case map[string]interface{}:
		return true
	}

	return false
}

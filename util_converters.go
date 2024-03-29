package partida

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

func ExtractNullError(err error) (v null.String) {
	if err != nil {
		v = null.NewString(err.Error(), true)
	}
	return v
}

func ExtractNullInt(data interface{}) (v null.Int) {
	switch data.(type) {
	case float64:
		v = null.NewInt(int64(data.(float64)), true)
	}
	return v
}

func ExtractNullString(data interface{}) (v null.String) {
	switch data.(type) {
	case string:
		v = null.NewString(data.(string), true)
	}
	return v
}

func ExtractNullTime(data interface{}) (v null.Time) {
	switch data.(type) {
	case string:
		date, err := time.Parse("2006-01-02T15:04:05", data.(string))
		if err != nil {
			date, err = time.Parse("2006-01-02T15:04:05Z", data.(string))
			if err != nil {
				date, err = time.Parse("2006-01-02T15:04:05-03:00", data.(string))
				if err != nil {
					date, err = time.Parse("2006-01-02T15:04:05+03:00", data.(string))
					if err != nil {
						return
					}
				}
			}
		}
		v = null.NewTime(date, true)
	}
	return v
}

func ExtractStringArray(data interface{}) (v []string) {
	switch data.(type) {
	case []interface{}:
		for _, e := range data.([]interface{}) {
			switch e.(type) {
			case string:
				v = append(v, e.(string))
			}
		}
	}
	return v
}

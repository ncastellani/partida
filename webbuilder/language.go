package webbuilder

type Language struct {
	Default bool `json:"default"`

	// general metadata
	Path        string `json:"path"`
	ShortName   string `json:"short_name"`
	LongName    string `json:"long_name"`
	Orientation string `json:"orientation"`

	// preferences
	DatetimeComplete string `json:"datetime_complete"`
	DatetimeDayMonth string `json:"datetime_daymonth"`
	DatetimeTime     string `json:"datetime_time"`
	MomentJS         string `json:"moment.js"`

	// units
	DecimalUnit   string `json:"decimal_unit"`
	ThousandsUnit string `json:"thousands_unit"`

	// translations
	Translations map[string]string `json:"translations"`
}

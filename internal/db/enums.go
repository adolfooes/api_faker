package db

// Define ENUM mappings for different tables and columns.
// Now that ENUM values in PostgreSQL are stored as strings, we map them directly as string-to-string.
var enumMappings = map[string]map[string]map[string]string{
	"url_config": {
		"method": {
			"GET":     "GET",
			"POST":    "POST",
			"PUT":     "PUT",
			"DELETE":  "DELETE",
			"PATCH":   "PATCH",
			"OPTIONS": "OPTIONS",
			"HEAD":    "HEAD",
		},
	},
	"project_users": {
		"access_level": {
			"read":  "Read",
			"write": "Write",
			"admin": "Admin",
		},
	},
}

// ConvertASCIIToString takes an array of int (ASCII codes) and converts it to a string.
func ConvertASCIIToString(asciiCodes []byte) string {
	return string(asciiCodes)
}

func TranslateEnumValue(table string, column string, value interface{}) interface{} {
	if tableEnums, tableExists := enumMappings[table]; tableExists {
		if columnEnums, columnExists := tableEnums[column]; columnExists {
			switch v := value.(type) {
			case []byte:
				convertedStr := ConvertASCIIToString(v)
				if str, valueExists := columnEnums[convertedStr]; valueExists {
					return str
				}
			}
		}
	}

	return value
}

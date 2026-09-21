package handler

import (
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var tokenRe = regexp.MustCompile(`\{\{([^}]+)\}\}`)

type templateContext struct {
	body       map[string]interface{}
	pathParams map[string]string
}

func generateUUID() string {
	b := make([]byte, 16)
	_, _ = crand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func resolveToken(token string, ctx templateContext) interface{} {
	switch token {
	case "uuid":
		return generateUUID()
	case "now":
		return time.Now().UTC().Format(time.RFC3339)
	default:
		if strings.HasPrefix(token, "request.body.") {
			field := token[len("request.body."):]
			val := getNestedField(ctx.body, strings.Split(field, "."))
			if val != nil {
				return val
			}
		} else if strings.HasPrefix(token, "path.") {
			name := token[len("path."):]
			if v, ok := ctx.pathParams[name]; ok {
				return v
			}
		}
		return "{{" + token + "}}"
	}
}

func getNestedField(obj map[string]interface{}, path []string) interface{} {
	if len(path) == 0 || obj == nil {
		return nil
	}
	val, ok := obj[path[0]]
	if !ok {
		return nil
	}
	if len(path) == 1 {
		return val
	}
	nested, ok := val.(map[string]interface{})
	if !ok {
		return nil
	}
	return getNestedField(nested, path[1:])
}

// interpolateString processes a single string value.
// If the entire string is exactly one token and the resolved value is non-string, the typed value is returned.
func interpolateString(s string, ctx templateContext) interface{} {
	if m := tokenRe.FindStringSubmatch(s); m != nil && m[0] == s {
		return resolveToken(m[1], ctx)
	}
	return tokenRe.ReplaceAllStringFunc(s, func(match string) string {
		parts := tokenRe.FindStringSubmatch(match)
		val := resolveToken(parts[1], ctx)
		switch v := val.(type) {
		case string:
			return v
		default:
			b, _ := json.Marshal(v)
			return string(b)
		}
	})
}

// interpolateModel recursively traverses a JSON-decoded value and replaces template tokens.
func interpolateModel(data interface{}, ctx templateContext) interface{} {
	switch v := data.(type) {
	case string:
		return interpolateString(v, ctx)
	case map[string]interface{}:
		result := make(map[string]interface{}, len(v))
		for k, val := range v {
			result[k] = interpolateModel(val, ctx)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = interpolateModel(item, ctx)
		}
		return result
	default:
		return v
	}
}

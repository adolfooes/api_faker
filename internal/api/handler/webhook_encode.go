package handler

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// interpolateHeaders resolves template tokens in each header value.
func interpolateHeaders(headers map[string]string, ctx templateContext) map[string]string {
	if headers == nil {
		return nil
	}
	result := make(map[string]string, len(headers))
	for k, v := range headers {
		switch resolved := interpolateString(v, ctx).(type) {
		case string:
			result[k] = resolved
		default:
			b, _ := json.Marshal(resolved)
			result[k] = string(b)
		}
	}
	return result
}

// encodeWebhookBody serializes body for the given content_type. A string body is
// sent as-is (raw). application/x-www-form-urlencoded serializes a nested map in
// the data[field]=value style (Iugu's format); anything else is JSON-encoded.
func encodeWebhookBody(contentType string, body interface{}) ([]byte, error) {
	if body == nil {
		return []byte{}, nil
	}
	if raw, ok := body.(string); ok {
		return []byte(raw), nil
	}
	if contentType == "application/x-www-form-urlencoded" {
		return []byte(encodeFormBody(body)), nil
	}
	return json.Marshal(body)
}

// encodeFormBody flattens a JSON-decoded value into a query-string-like body
// using bracket notation for nested fields (e.g. data[id]=1, data[a][b]=2),
// matching how Iugu sends form-encoded webhooks.
func encodeFormBody(body interface{}) string {
	var pairs []string
	flattenFormFields("", body, &pairs)
	return strings.Join(pairs, "&")
}

func flattenFormFields(prefix string, value interface{}, pairs *[]string) {
	switch v := value.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			key := k
			if prefix != "" {
				key = prefix + "[" + k + "]"
			}
			flattenFormFields(key, v[k], pairs)
		}
	case []interface{}:
		for i, item := range v {
			flattenFormFields(fmt.Sprintf("%s[%d]", prefix, i), item, pairs)
		}
	case nil:
		*pairs = append(*pairs, prefix+"=")
	default:
		*pairs = append(*pairs, prefix+"="+url.QueryEscape(fmt.Sprintf("%v", v)))
	}
}

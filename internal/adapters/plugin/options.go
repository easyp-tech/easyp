package plugin

import (
	"sort"
	"strings"
)

func flattenOptions(options map[string][]string) (string, bool) {
	if len(options) == 0 {
		return "", false
	}

	keys := make([]string, 0, len(options))
	for key := range options {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	flattened := make([]string, 0, len(options))
	for _, key := range keys {
		values := options[key]
		for _, value := range values {
			if value == "" {
				flattened = append(flattened, key)
				continue
			}

			flattened = append(flattened, key+"="+value)
		}
	}

	if len(flattened) == 0 {
		return "", false
	}

	return strings.Join(flattened, ","), true
}

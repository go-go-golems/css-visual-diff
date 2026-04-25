package jsapi

import "encoding/json"

func decodeInto[T any](raw any) (T, error) {
	var out T
	b, err := json.Marshal(raw)
	if err != nil {
		return out, err
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return out, err
	}
	return out, nil
}

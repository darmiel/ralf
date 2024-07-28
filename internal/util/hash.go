package util

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
)

func CreateCacheKey(of any) (string, error) {
	data, err := json.Marshal(of)
	if err != nil {
		return "", err
	}
	hasher := fnv.New128a()
	if _, err = hasher.Write(data); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

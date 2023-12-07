package util

import "hash/fnv"

func Hash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))

	return int(h.Sum32())
}

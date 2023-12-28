package memdb

import (
	"strconv"
)

type Hash struct {
	table map[string][]byte
}

func NewHash() *Hash {
	return &Hash{table: make(map[string][]byte)}
}

func (h *Hash) Set(key string, v []byte) int {
	added := 0
	if !h.Exist(key) {
		added = 1
	}
	h.table[key] = v
	return added
}

func (h *Hash) Get(key string) []byte {
	return h.table[key]
}

func (h *Hash) Del(key string) int {
	if _, ok := h.table[key]; !ok {
		return 0
	}
	delete(h.table, key)
	return 1
}

func (h *Hash) Len() int {
	return len(h.table)
}

func (h *Hash) Exist(key string) bool {
	_, ok := h.table[key]
	return ok
}

func (h *Hash) Keys() []string {
	res := make([]string, 0, h.Len())
	for key := range h.table {
		res = append(res, key)
	}
	return res
}

func (h *Hash) Values() [][]byte {
	res := make([][]byte, 0, h.Len())
	for _, val := range h.table {
		res = append(res, val)
	}
	return res
}

func (h *Hash) Clear() {
	h.table = make(map[string][]byte)
}

func (h *Hash) StrLen(key string) int {
	return len(h.table[key])
}

func (h *Hash) IsEmpty() bool {
	return len(h.table) == 0
}

func (h *Hash) Table() map[string][]byte {
	return h.table
}

func (h *Hash) IncrBy(key string, increment int) (int, bool) {
	v := h.Get(key)
	if v == nil {
		h.Set(key, []byte(strconv.Itoa(increment)))
		return increment, true
	} else {
		n, err := strconv.Atoi(string(v))
		if err != nil {
			return 0, false
		}
		value := n + increment
		h.Set(key, []byte(strconv.Itoa(value)))
		return value, true
	}
}

func (h *Hash) IncrByFloat(key string, increment float64) (float64, bool) {
	v := h.Get(key)
	if v == nil {
		h.Set(key, []byte(strconv.FormatFloat(increment, 'f', -1, 64)))
		return increment, true
	} else {
		f, err := strconv.ParseFloat(string(v), 64)
		if err != nil {
			return 0, false
		}
		value := f + increment
		h.Set(key, []byte(strconv.FormatFloat(value, 'f', -1, 64)))
		return value, true
	}
}

func (h *Hash) Random(count int) []string {
	var keys []string

	if count == 0 || h.Len() == 0 {
		return make([]string, 0)
	}

	if count > 0 {
		if count > h.Len() {
			count = h.Len()
		}

		keys = make([]string, 0, count)

		for key := range h.table {
			// If the provided count argument is positive, return an array of distinct fields.
			if len(keys) >= count {
				break
			}
			keys = append(keys, key)
		}
	} else if count < 0 {
		keys = make([]string, 0, -count)
		for {
			for key := range h.table {
				// If called with a negative count, the behavior changes and the command is allowed to return the same field multiple times.
				if len(keys) >= -count {
					return keys
				}
				keys = append(keys, key)
				break
			}
		}
	}

	return keys
}

func (h *Hash) RandomWithValue(count int) ([]string, [][]byte) {
	var keys []string
	var vals [][]byte

	if count == 0 || h.Len() == 0 {
		return make([]string, 0), make([][]byte, 0)
	}

	if count > 0 {
		if count > h.Len() {
			count = h.Len()
		}

		keys = make([]string, 0, count)
		vals = make([][]byte, 0, count)

		for key, val := range h.table {
			// If the provided count argument is positive, return an array of distinct fields.
			if len(keys) >= count {
				break
			}
			keys = append(keys, key)
			vals = append(vals, val)
		}
	} else if count < 0 {
		keys = make([]string, 0, -count)
		vals = make([][]byte, 0, -count)
		for {
			for key, val := range h.table {
				// If called with a negative count, the behavior changes and the command is allowed to return the same field multiple times.
				if len(keys) >= -count {
					return keys, vals
				}
				keys = append(keys, key)
				vals = append(vals, val)
				break
			}
		}
	}

	return keys, vals
}

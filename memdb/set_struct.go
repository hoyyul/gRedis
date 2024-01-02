package memdb

type void struct{}

type Set struct {
	table map[string]void
}

func NewSet() *Set {
	return &Set{table: make(map[string]void)}
}

func (s *Set) Has(key string) bool {
	_, ok := s.table[key]
	return ok
}

func (s *Set) Len() int {
	return len(s.table)
}

func (s *Set) Add(key string) int {
	if s.Has(key) {
		return 0
	}
	s.table[key] = void{}
	return 1
}

func (s *Set) Remove(key string) int {
	if s.Has(key) {
		delete(s.table, key)
		return 1
	}
	return 0
}

func (s *Set) Pop() string {
	for key := range s.table {
		delete(s.table, key)
		return key
	}
	return ""
}

func (s *Set) Members() []string {
	res := make([]string, 0, s.Len())
	for key := range s.table {
		res = append(res, key)
	}
	return res
}

func (s *Set) Move(set *Set, key string) int {
	if !s.Has(key) {
		return 0
	}
	s.Remove(key)
	set.Add(key)
	return 1
}

func (s *Set) Clear() {
	s.table = make(map[string]void)
}

// set operations

func (s *Set) Union(sets ...*Set) *Set {
	res := &Set{table: make(map[string]void)}
	for key := range s.table {
		res.Add(key)
	}
	for _, set := range sets {
		for key := range set.table {
			res.Add(key)
		}
	}
	return res
}

func (s *Set) Intersect(sets ...*Set) *Set {
	res := &Set{table: make(map[string]void)}
	for key := range s.table {
		res.Add(key)
	}
	for _, set := range sets {
		for key := range res.table {
			if !set.Has(key) {
				res.Remove(key)
			}
		}
	}
	return res
}

func (s *Set) Difference(sets ...*Set) *Set {
	res := &Set{table: make(map[string]void)}
	for key := range s.table {
		res.Add(key)
	}
	for _, set := range sets {
		for key := range set.table {
			res.Remove(key)
		}
	}
	return res
}

func (s *Set) isSubset(set *Set) bool {
	for key := range set.table {
		if !s.Has(key) {
			return false
		}
	}
	return true
}

func (s *Set) Random(count int) []string {
	var res []string

	if count == 0 || s.Len() == 0 {
		return make([]string, 0)
	}

	if count > 0 {
		if count > s.Len() {
			count = s.Len()
		}

		res = make([]string, 0, count)

		for key := range s.table {
			// If the provided count argument is positive, return an array of distinct fields.
			if len(res) >= count {
				break
			}
			res = append(res, key)
		}
	} else if count < 0 {
		res = make([]string, 0, -count)
		for {
			for key := range s.table {
				// If called with a negative count, the behavior changes and the command is allowed to return the same field multiple times.
				if len(res) >= -count {
					return res
				}
				res = append(res, key)
				break
			}
		}
	}

	return res
}

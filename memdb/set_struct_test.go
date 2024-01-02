package memdb

import "testing"

func TestDifference1(t *testing.T) {
	s1 := NewSet()
	s2 := NewSet()

	s1.Add("a")
	s1.Add("b")
	s1.Add("c")

	s2.Add("c")
	s2.Add("d")
	s2.Add("e")

	res := s1.Difference(s2)
	var a, b, c bool
	for _, key := range res.Members() {
		if key == "a" {
			a = true
		}
		if key == "b" {
			b = true
		}
		if key == "c" {
			c = true
		}
	}

	if !a || !b || c {
		t.Error(res.Members())
	}
}

func TestDifference2(t *testing.T) {
	s1 := NewSet()
	s2 := NewSet()

	res := s1.Difference(s2)

	if res.Len() != 0 {
		t.Error(res.Members())
	}
}
